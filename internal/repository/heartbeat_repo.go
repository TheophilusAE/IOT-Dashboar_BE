package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HeartbeatRepo struct {
	pool *pgxpool.Pool
}

func NewHeartbeatRepo(pool *pgxpool.Pool) *HeartbeatRepo {
	return &HeartbeatRepo{pool: pool}
}

type InsertHeartbeatParams struct {
	DeviceID      int64
	DeviceTS      *time.Time
	RSSI          *int32
	UptimeSeconds *int64
	FreeHeapBytes *int32
}

func (r *HeartbeatRepo) Insert(ctx context.Context, p InsertHeartbeatParams) (int64, time.Time, error) {
	var id int64
	var receivedAt time.Time
	err := r.pool.QueryRow(ctx, `
		INSERT INTO device_heartbeats (device_id, device_ts, rssi, uptime_seconds, free_heap_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, received_at
	`, p.DeviceID, p.DeviceTS, p.RSSI, p.UptimeSeconds, p.FreeHeapBytes).Scan(&id, &receivedAt)
	return id, receivedAt, err
}

// LastReceivedAt returns the timestamp of the most recent heartbeat for a
// device, or (zero, ErrNotFound) if the device has never sent one — the
// "never_connected" case, distinct from "offline".
func (r *HeartbeatRepo) LastReceivedAt(ctx context.Context, deviceID int64) (time.Time, error) {
	var t time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT received_at FROM device_heartbeats
		WHERE device_id = $1
		ORDER BY received_at DESC LIMIT 1
	`, deviceID).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrNotFound
	}
	return t, err
}
