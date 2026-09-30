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

type dailyCountDTO struct {
	Date      string `json:"date"`
	Organic   int64  `json:"organic"`
	Anorganic int64  `json:"anorganic"`
}

// DailyStats returns one point per calendar day (7 or 30 days, ending today)
// with organic/anorganic counts, for rendering a per-day trend chart.
func (h *StatsHandler) DailyStats(w http.ResponseWriter, r *http.Request) {
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "7d"
	}

	var numDays int
	switch rangeParam {
	case "7d":
		numDays = 7
	case "30d":
		numDays = 30
	default:
		writeError(w, http.StatusBadRequest, "unsupported range: must be one of 7d, 30d")
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

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	since := todayStart.AddDate(0, 0, -(numDays - 1))

	days, err := h.detections.DailyCountsSince(r.Context(), since, deviceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]dailyCountDTO, len(days))
	for i, d := range days {
		out[i] = dailyCountDTO{
			Date:      d.Date.Format("2006-01-02"),
			Organic:   d.Organic,
			Anorganic: d.Anorganic,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"range": rangeParam,
		"days":  out,
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

