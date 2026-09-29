package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"iot-backend/internal/models"
)

var ErrNotFound = errors.New("not found")

type DeviceRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

// GetOrRegisterByUID returns the device with the given UID, auto-registering
// it (inactive, for later review) if it has never been seen before. This
// keeps ingestion working during bring-up without ever losing real data
// because a device wasn't pre-provisioned.
func (r *DeviceRepo) GetOrRegisterByUID(ctx context.Context, uid string) (models.Device, error) {
	d, err := r.GetByUID(ctx, uid)
	if err == nil {
		return d, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return models.Device{}, err
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO devices (device_uid, is_active)
		VALUES ($1, false)
		ON CONFLICT (device_uid) DO UPDATE SET device_uid = EXCLUDED.device_uid
		RETURNING id, device_uid, display_name, location, firmware_version, registered_at, is_active
	`, uid)
	return scanDevice(row)
}

func (r *DeviceRepo) GetByUID(ctx context.Context, uid string) (models.Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, device_uid, display_name, location, firmware_version, registered_at, is_active
		FROM devices WHERE device_uid = $1
	`, uid)
	d, err := scanDevice(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Device{}, ErrNotFound
	}
	return d, err
}

func (r *DeviceRepo) GetByID(ctx context.Context, id int64) (models.Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, device_uid, display_name, location, firmware_version, registered_at, is_active
		FROM devices WHERE id = $1
	`, id)
	d, err := scanDevice(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Device{}, ErrNotFound
	}
	return d, err
}

func (r *DeviceRepo) List(ctx context.Context) ([]models.Device, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, device_uid, display_name, location, firmware_version, registered_at, is_active
		FROM devices ORDER BY registered_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}
	defer rows.Close()

	devices := []models.Device{}
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// UpdateFirmwareVersion records the firmware version reported by a heartbeat,
// so /devices reflects what the device actually last announced.
func (r *DeviceRepo) UpdateFirmwareVersion(ctx context.Context, deviceID int64, version string) error {
	_, err := r.pool.Exec(ctx, `UPDATE devices SET firmware_version = $2 WHERE id = $1`, deviceID, version)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDevice(row rowScanner) (models.Device, error) {
	var d models.Device
	err := row.Scan(&d.ID, &d.DeviceUID, &d.DisplayName, &d.Location, &d.FirmwareVersion, &d.RegisteredAt, &d.IsActive)
	return d, err
}
