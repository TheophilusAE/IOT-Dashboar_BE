package api

import (
	"net/http"
	"time"

	"iot-backend/internal/models"
	"iot-backend/internal/repository"
	"iot-backend/internal/status"
)

type DashboardHandler struct {
	devices    *repository.DeviceRepo
	sensors    *repository.SensorRepo
	detections *repository.DetectionRepo
	resolver   *status.Resolver
}

func NewDashboardHandler(devices *repository.DeviceRepo, sensors *repository.SensorRepo, detections *repository.DetectionRepo, resolver *status.Resolver) *DashboardHandler {
	return &DashboardHandler{devices: devices, sensors: sensors, detections: detections, resolver: resolver}
}

type dashboardBinDTO struct {
	DeviceID      int64          `json:"device_id"`
	DeviceUID     string         `json:"device_uid"`
	BinType       models.BinType `json:"bin_type"`
	Status        string         `json:"status"`
	FillPercent   *float64       `json:"fill_percent"`
	DistanceCM    *float64       `json:"distance_cm"`
	LastReadingAt *time.Time     `json:"last_reading_at"`
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	devices, err := h.devices.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	onlineCount, offlineCount := 0, 0
	bins := make([]dashboardBinDTO, 0, len(devices)*2)

	for _, d := range devices {
		connStatus, _, _, err := h.resolver.Resolve(ctx, d.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		switch connStatus {
		case models.StatusOnline:
			onlineCount++
		case models.StatusOffline, models.StatusNeverConnected:
			offlineCount++
		}

		latest, err := h.sensors.LatestForDevice(ctx, d.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, lbr := range latest {
			b := dashboardBinDTO{
				DeviceID:  d.ID,
				DeviceUID: d.DeviceUID,
				BinType:   lbr.BinType,
				Status:    string(lbr.Status),
			}
			if lbr.Reading != nil {
				b.FillPercent = lbr.Reading.FillPercent
				dc := lbr.Reading.DistanceCM
				b.DistanceCM = &dc
				ra := lbr.Reading.ReceivedAt
				b.LastReadingAt = &ra
			}
			bins = append(bins, b)
		}
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayCounts, err := h.detections.CountsSince(ctx, todayStart)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"generated_at": time.Now(),
		"devices": map[string]int{
			"total":   len(devices),
			"online":  onlineCount,
			"offline": offlineCount,
		},
		"bins":         bins,
		"today_counts": todayCounts,
	})
}
