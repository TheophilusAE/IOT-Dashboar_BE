package ws

// OutgoingMessage is the envelope for every server -> client push. Every
// broadcast is triggered by an actual successful DB write in the ingest
// handlers — nothing here is generated on a timer.
type OutgoingMessage struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type SubscribeMessage struct {
	Type      string  `json:"type"`
	DeviceIDs []int64 `json:"device_ids"`
}

const (
	TypeSubscribed    = "subscribed"
	TypeDetection     = "detection_event"
	TypeSensorReading = "sensor_reading"
	TypeDeviceStatus  = "device_status"
	TypePing          = "ping"
)
