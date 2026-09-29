package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"iot-backend/internal/repository"
)

var errUnsupportedRange = errors.New("unsupported range: must be one of today, 7d, 30d")

type StatsHandler struct {
	detections *repository.DetectionRepo
}

func NewStatsHandler(detections *repository.DetectionRepo) *StatsHandler {
	return &StatsHandler{detections: detections}
}

func (h *StatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "today"
	}

	since, err := rangeToSince(rangeParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var deviceID *int64
	if v := r.URL.Query().Get("device_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid device_id")
			return
		}
		deviceID = &id
	}

	stats, err := h.detections.StatsSince(r.Context(), since, deviceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"range":            rangeParam,
		"total_detections": stats.TotalDetections,
		"by_type":          stats.ByType,
		"avg_confidence":   stats.AvgConfidence,
	})
}

func rangeToSince(rangeParam string) (time.Time, error) {
	now := time.Now()
	switch rangeParam {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "7d":
		return now.AddDate(0, 0, -7), nil
	case "30d":
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, errUnsupportedRange
	}
}

