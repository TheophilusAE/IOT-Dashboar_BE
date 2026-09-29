package api

import (
	"net/http"
	"strconv"
	"time"

	"iot-backend/internal/models"
	"iot-backend/internal/repository"
)

type DetectionsHandler struct {
	detections *repository.DetectionRepo
}

func NewDetectionsHandler(detections *repository.DetectionRepo) *DetectionsHandler {
	return &DetectionsHandler{detections: detections}
}

func (h *DetectionsHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := repository.ListDetectionsParams{}

	if v := q.Get("device_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid device_id")
			return
		}
		params.DeviceID = &id
	}

	if v := q.Get("waste_type"); v != "" {
		bt, ok := models.NormalizeBinType(v)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid waste_type")
			return
		}
		params.WasteType = &bt
	}

	if v := q.Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from: must be RFC3339")
			return
		}
		params.From = &t
	}

	if v := q.Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to: must be RFC3339")
			return
		}
		params.To = &t
	}

	params.Page = 1
	if v := q.Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
		params.Page = p
	}

	params.PageSize = 20
	if v := q.Get("page_size"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 || ps > 200 {
			writeError(w, http.StatusBadRequest, "invalid page_size (1-200)")
			return
		}
		params.PageSize = ps
	}

	items, total, err := h.detections.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]detectionDTO, 0, len(items))
	for _, e := range items {
		dtos = append(dtos, toDetectionDTO(e))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"page":        params.Page,
		"page_size":   params.PageSize,
		"total_count": total,
		"items":       dtos,
	})
}
