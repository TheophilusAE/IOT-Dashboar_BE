package ingest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"iot-backend/internal/models"
	"iot-backend/internal/repository"
	"iot-backend/internal/ws"
)

type Handlers struct {
	devices    *repository.DeviceRepo
	heartbeats *repository.HeartbeatRepo
	detections *repository.DetectionRepo
	sensors    *repository.SensorRepo
	hub        *ws.Hub
}

func NewHandlers(
	devices *repository.DeviceRepo,
	heartbeats *repository.HeartbeatRepo,
	detections *repository.DetectionRepo,
	sensors *repository.SensorRepo,
	hub *ws.Hub,
) *Handlers {
	return &Handlers{devices: devices, heartbeats: heartbeats, detections: detections, sensors: sensors, hub: hub}
}

// resolveDevice reads X-Device-UID and auto-registers unknown devices
// (inactive, pending review) rather than rejecting the request — dropping
// real ingested data because a device wasn't pre-provisioned would violate
// the "never fabricate, never silently discard real data" requirement.
func (h *Handlers) resolveDevice(w http.ResponseWriter, r *http.Request) (models.Device, bool) {
	uid := r.Header.Get("X-Device-UID")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "missing X-Device-UID header")
		return models.Device{}, false
	}
	d, err := h.devices.GetOrRegisterByUID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "resolving device: "+err.Error())
		return models.Device{}, false
	}
	return d, true
}

func readRawBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func (h *Handlers) Heartbeat(w http.ResponseWriter, r *http.Request) {
	device, ok := h.resolveDevice(w, r)
	if !ok {
		return
	}

	var dto HeartbeatDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	_, receivedAt, err := h.heartbeats.Insert(r.Context(), repository.InsertHeartbeatParams{
		DeviceID:      device.ID,
		DeviceTS:      dto.DeviceTS,
		RSSI:          dto.RSSI,
		UptimeSeconds: dto.UptimeSeconds,
		FreeHeapBytes: dto.FreeHeapBytes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "storing heartbeat: "+err.Error())
		return
	}

	if dto.FirmwareVersion != "" {
		_ = h.devices.UpdateFirmwareVersion(r.Context(), device.ID, dto.FirmwareVersion)
	}

	h.hub.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeDeviceStatus,
		Data: map[string]any{
			"device_id":         device.ID,
			"connection_status": "online",
			"last_heartbeat_at": receivedAt,
		},
	}, &device.ID)

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "server_time": receivedAt})
}

func (h *Handlers) Detection(w http.ResponseWriter, r *http.Request) {
	device, ok := h.resolveDevice(w, r)
	if !ok {
		return
	}

	raw, err := readRawBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading body: "+err.Error())
		return
	}

	var dto DetectionDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	mapped, err := mapDetection(device.ID, dto, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := h.detections.Insert(r.Context(), repository.InsertDetectionParams{
		DeviceID:   mapped.DeviceID,
		WasteType:  mapped.WasteType,
		Confidence: mapped.Confidence,
		DeviceTS:   mapped.DeviceTS,
		RawPayload: mapped.RawPayload,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "storing detection: "+err.Error())
		return
	}

	h.hub.Broadcast(ws.OutgoingMessage{
		Type: ws.TypeDetection,
		Data: map[string]any{
			"id":          event.ID,
			"device_id":   event.DeviceID,
			"waste_type":  event.WasteType,
			"confidence":  event.Confidence,
			"received_at": event.ReceivedAt,
		},
	}, &device.ID)

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "id": event.ID})
}

func (h *Handlers) Sensor(w http.ResponseWriter, r *http.Request) {
	device, ok := h.resolveDevice(w, r)
	if !ok {
		return
	}

	raw, err := readRawBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reading body: "+err.Error())
		return
	}

	var batch SensorBatchDTO
	if err := json.Unmarshal(raw, &batch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(batch.Readings) == 0 {
		writeError(w, http.StatusBadRequest, "readings array is empty")
		return
	}

	inserted := 0
	for _, dto := range batch.Readings {
		mapped, err := mapSensorReading(device.ID, dto, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		reading, err := h.sensors.Insert(r.Context(), repository.InsertSensorReadingParams{
			DeviceID:    mapped.DeviceID,
			BinType:     mapped.BinType,
			DistanceCM:  mapped.DistanceCM,
			FillPercent: mapped.FillPercent,
			DeviceTS:    mapped.DeviceTS,
			RawPayload:  mapped.RawPayload,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "storing sensor reading: "+err.Error())
			return
		}
		inserted++

		h.hub.Broadcast(ws.OutgoingMessage{
			Type: ws.TypeSensorReading,
			Data: map[string]any{
				"device_id":    reading.DeviceID,
				"bin_type":     reading.BinType,
				"distance_cm":  reading.DistanceCM,
				"fill_percent": reading.FillPercent,
				"received_at":  reading.ReceivedAt,
			},
		}, &device.ID)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "inserted": inserted})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
