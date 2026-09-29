package api

import (
	"time"

	"iot-backend/internal/models"
)

type binStatusDTO struct {
	BinType       models.BinType `json:"bin_type"`
	Status        string         `json:"status"` // "has_data" | "no_data"
	FillPercent   *float64       `json:"fill_percent"`
	DistanceCM    *float64       `json:"distance_cm"`
	LastReadingAt *time.Time     `json:"last_reading_at"`
}

func toBinStatusDTO(lbr models.LatestBinReading) binStatusDTO {
	dto := binStatusDTO{
		BinType: lbr.BinType,
		Status:  string(lbr.Status),
	}
	if lbr.Reading != nil {
		dto.FillPercent = lbr.Reading.FillPercent
		dc := lbr.Reading.DistanceCM
		dto.DistanceCM = &dc
		ra := lbr.Reading.ReceivedAt
		dto.LastReadingAt = &ra
	}
	return dto
}

type deviceStatusDTO struct {
	DeviceID                  int64          `json:"device_id"`
	DeviceUID                 string         `json:"device_uid"`
	DisplayName               string         `json:"display_name"`
	ConnectionStatus          string         `json:"connection_status"`
	LastHeartbeatAt           *time.Time     `json:"last_heartbeat_at"`
	SecondsSinceLastHeartbeat *int64         `json:"seconds_since_last_heartbeat"`
	FirmwareVersion           *string        `json:"firmware_version"`
	Bins                      []binStatusDTO `json:"bins"`
}

type detectionDTO struct {
	ID         int64          `json:"id"`
	DeviceID   int64          `json:"device_id"`
	WasteType  models.BinType `json:"waste_type"`
	Confidence float64        `json:"confidence"`
	ReceivedAt time.Time      `json:"received_at"`
}

func toDetectionDTO(e models.DetectionEvent) detectionDTO {
	return detectionDTO{
		ID:         e.ID,
		DeviceID:   e.DeviceID,
		WasteType:  e.WasteType,
		Confidence: e.Confidence,
		ReceivedAt: e.ReceivedAt,
	}
}
