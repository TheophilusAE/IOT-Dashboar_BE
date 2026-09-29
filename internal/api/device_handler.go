package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"iot-backend/internal/repository"
	"iot-backend/internal/status"
)

type DeviceHandler struct {
	devices  *repository.DeviceRepo
	sensors  *repository.SensorRepo
	resolver *status.Resolver
}

func NewDeviceHandler(devices *repository.DeviceRepo, sensors *repository.SensorRepo, resolver *status.Resolver) *DeviceHandler {
	return &DeviceHandler{devices: devices, sensors: sensors, resolver: resolver}
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.devices.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]deviceStatusDTO, 0, len(devices))
	for _, d := range devices {
		dto, err := h.buildStatusDTO(r, d.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dtos = append(dtos, dto)
	}

	writeJSON(w, http.StatusOK, map[string]any{"devices": dtos})
}

func (h *DeviceHandler) Status(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	dto, err := h.buildStatusDTO(r, id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

func (h *DeviceHandler) buildStatusDTO(r *http.Request, deviceID int64) (deviceStatusDTO, error) {
	device, err := h.devices.GetByID(r.Context(), deviceID)
	if err != nil {
		return deviceStatusDTO{}, err
	}

	connStatus, lastSeen, secondsSince, err := h.resolver.Resolve(r.Context(), deviceID)
	if err != nil {
		return deviceStatusDTO{}, err
	}

	bins, err := h.sensors.LatestForDevice(r.Context(), deviceID)
	if err != nil {
		return deviceStatusDTO{}, err
	}
	binDTOs := make([]binStatusDTO, 0, len(bins))
	for _, b := range bins {
		binDTOs = append(binDTOs, toBinStatusDTO(b))
	}

	return deviceStatusDTO{
		DeviceID:                  device.ID,
		DeviceUID:                 device.DeviceUID,
		DisplayName:               device.DisplayName,
		ConnectionStatus:          string(connStatus),
		LastHeartbeatAt:           lastSeen,
		SecondsSinceLastHeartbeat: secondsSince,
		FirmwareVersion:           device.FirmwareVersion,
		Bins:                      binDTOs,
	}, nil
}
