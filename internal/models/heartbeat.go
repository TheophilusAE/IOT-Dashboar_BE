package models

import "time"

type Heartbeat struct {
	ID             int64
	DeviceID       int64
	ReceivedAt     time.Time
	DeviceTS       *time.Time
	RSSI           *int32
	UptimeSeconds  *int64
	FreeHeapBytes  *int32
}
