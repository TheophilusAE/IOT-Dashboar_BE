package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"iot-backend/internal/models"
)

type SensorRepo struct {
	pool *pgxpool.Pool
}

func NewSensorRepo(pool *pgxpool.Pool) *SensorRepo {
	return &SensorRepo{pool: pool}
}

type InsertSensorReadingParams struct {
	DeviceID    int64
	BinType     models.BinType
	DistanceCM  float64
	FillPercent *float64
	DeviceTS    *time.Time
	RawPayload  []byte
}

func (r *SensorRepo) Insert(ctx context.Context, p InsertSensorReadingParams) (models.SensorReading, error) {
	var sr models.SensorReading
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sensor_readings (device_id, bin_type, distance_cm, fill_percent, device_ts, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, device_id, bin_type, distance_cm, fill_percent, received_at, device_ts, raw_payload
	`, p.DeviceID, p.BinType, p.DistanceCM, p.FillPercent, p.DeviceTS, p.RawPayload).Scan(
		&sr.ID, &sr.DeviceID, &sr.BinType, &sr.DistanceCM, &sr.FillPercent, &sr.ReceivedAt, &sr.DeviceTS, &sr.RawPayload,
	)
	return sr, err
}

// LatestForDevice returns the latest reading for each known bin type for a
// device. A bin with no row yet comes back with Status=no_data and a nil
// Reading — never a fabricated zero reading.
func (r *SensorRepo) LatestForDevice(ctx context.Context, deviceID int64) ([]models.LatestBinReading, error) {
	binTypes := []models.BinType{models.BinOrganic, models.BinAnorganic}
	results := make([]models.LatestBinReading, 0, len(binTypes))

	for _, bt := range binTypes {
		var sr models.SensorReading
		err := r.pool.QueryRow(ctx, `
			SELECT id, device_id, bin_type, distance_cm, fill_percent, received_at, device_ts, raw_payload
			FROM sensor_readings
			WHERE device_id = $1 AND bin_type = $2
			ORDER BY received_at DESC LIMIT 1
		`, deviceID, bt).Scan(
			&sr.ID, &sr.DeviceID, &sr.BinType, &sr.DistanceCM, &sr.FillPercent, &sr.ReceivedAt, &sr.DeviceTS, &sr.RawPayload,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			results = append(results, models.LatestBinReading{BinType: bt, Status: models.BinStatusNoData, Reading: nil})
			continue
		}
		if err != nil {
			return nil, err
		}
		reading := sr
		results = append(results, models.LatestBinReading{BinType: bt, Status: models.BinStatusHasData, Reading: &reading})
	}

	return results, nil
}
