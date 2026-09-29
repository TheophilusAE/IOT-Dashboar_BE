package models

import "time"

type DetectionEvent struct {
	ID         int64
	DeviceID   int64
	WasteType  BinType
	Confidence float64
	ReceivedAt time.Time
	DeviceTS   *time.Time
	RawPayload []byte
}

// BinStatus is "has_data" only when a real row exists for that device+bin.
type BinReadingStatus string

const (
	BinStatusHasData BinReadingStatus = "has_data"
	BinStatusNoData  BinReadingStatus = "no_data"
)
