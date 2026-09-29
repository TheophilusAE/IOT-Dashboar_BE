package models

import "time"

type Device struct {
	ID              int64
	DeviceUID       string
	DisplayName     string
	Location        *string
	FirmwareVersion *string
	RegisteredAt    time.Time
	IsActive        bool
}

// ConnectionStatus is derived at query time from heartbeat recency,
// never stored, so it can never drift out of sync with reality.
type ConnectionStatus string

const (
	StatusOnline         ConnectionStatus = "online"
	StatusOffline        ConnectionStatus = "offline"
	StatusNeverConnected ConnectionStatus = "never_connected"
)

type DeviceStatus struct {
	Device                     Device
	ConnectionStatus           ConnectionStatus
	LastHeartbeatAt            *time.Time
	SecondsSinceLastHeartbeat  *int64
}
