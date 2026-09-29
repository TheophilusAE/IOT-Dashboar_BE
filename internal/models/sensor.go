package models

import "time"

type SensorReading struct {
	ID          int64
	DeviceID    int64
	BinType     BinType
	DistanceCM  float64
	FillPercent *float64
	ReceivedAt  time.Time
	DeviceTS    *time.Time
	RawPayload  []byte
}

// LatestBinReading is the result of the "latest reading per device+bin"
// query. Reading is nil when no row exists yet for that bin, in which case
// Status must be BinStatusNoData — the API layer must not invent zeros.
type LatestBinReading struct {
	BinType BinType
	Status  BinReadingStatus
	Reading *SensorReading
}
