package status

import (
	"context"
	"errors"
	"time"

	"iot-backend/internal/models"
	"iot-backend/internal/repository"
)

// Resolver computes device connection status purely from real heartbeat
// timestamps. It is the single implementation shared by REST handlers and
// the WebSocket layer so "online"/"offline"/"never_connected" can never
// drift or be represented inconsistently.
type Resolver struct {
	heartbeats             *repository.HeartbeatRepo
	onlineThresholdSeconds int
}

func NewResolver(heartbeats *repository.HeartbeatRepo, onlineThresholdSeconds int) *Resolver {
	return &Resolver{heartbeats: heartbeats, onlineThresholdSeconds: onlineThresholdSeconds}
}

func (r *Resolver) Resolve(ctx context.Context, deviceID int64) (models.ConnectionStatus, *time.Time, *int64, error) {
	lastSeen, err := r.heartbeats.LastReceivedAt(ctx, deviceID)
	if errors.Is(err, repository.ErrNotFound) {
		return models.StatusNeverConnected, nil, nil, nil
	}
	if err != nil {
		return "", nil, nil, err
	}

	elapsed := int64(time.Since(lastSeen).Seconds())
	if elapsed > int64(r.onlineThresholdSeconds) {
		return models.StatusOffline, &lastSeen, &elapsed, nil
	}
	return models.StatusOnline, &lastSeen, &elapsed, nil
}
