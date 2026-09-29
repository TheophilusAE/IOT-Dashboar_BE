package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"iot-backend/internal/models"
)

type DetectionRepo struct {
	pool *pgxpool.Pool
}

func NewDetectionRepo(pool *pgxpool.Pool) *DetectionRepo {
	return &DetectionRepo{pool: pool}
}

type InsertDetectionParams struct {
	DeviceID   int64
	WasteType  models.BinType
	Confidence float64
	DeviceTS   *time.Time
	RawPayload []byte
}

func (r *DetectionRepo) Insert(ctx context.Context, p InsertDetectionParams) (models.DetectionEvent, error) {
	var e models.DetectionEvent
	err := r.pool.QueryRow(ctx, `
		INSERT INTO detection_events (device_id, waste_type, confidence, device_ts, raw_payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, device_id, waste_type, confidence, received_at, device_ts, raw_payload
	`, p.DeviceID, p.WasteType, p.Confidence, p.DeviceTS, p.RawPayload).Scan(
		&e.ID, &e.DeviceID, &e.WasteType, &e.Confidence, &e.ReceivedAt, &e.DeviceTS, &e.RawPayload,
	)
	return e, err
}

type ListDetectionsParams struct {
	DeviceID  *int64
	WasteType *models.BinType
	From      *time.Time
	To        *time.Time
	Page      int
	PageSize  int
}

func (r *DetectionRepo) List(ctx context.Context, p ListDetectionsParams) ([]models.DetectionEvent, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return placeholder(len(args))
	}

	if p.DeviceID != nil {
		where += " AND device_id = " + arg(*p.DeviceID)
	}
	if p.WasteType != nil {
		where += " AND waste_type = " + arg(*p.WasteType)
	}
	if p.From != nil {
		where += " AND received_at >= " + arg(*p.From)
	}
	if p.To != nil {
		where += " AND received_at < " + arg(*p.To)
	}

	pageSize := p.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, device_id, waste_type, confidence, received_at, device_ts, raw_payload,
		       COUNT(*) OVER() AS total_count
		FROM detection_events
		` + where + `
		ORDER BY received_at DESC
		LIMIT ` + arg(pageSize) + ` OFFSET ` + arg(offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []models.DetectionEvent{}
	var total int64
	for rows.Next() {
		var e models.DetectionEvent
		if err := rows.Scan(&e.ID, &e.DeviceID, &e.WasteType, &e.Confidence, &e.ReceivedAt, &e.DeviceTS, &e.RawPayload, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, rows.Err()
}

type Stats struct {
	TotalDetections int64
	ByType          map[models.BinType]int64
	AvgConfidence   *float64
}

// CountsSince returns real GROUP BY counts per waste type since the given
// time. Types with zero rows are explicitly present with count 0 (a real
// aggregate result), not omitted.
func (r *DetectionRepo) CountsSince(ctx context.Context, since time.Time) (map[models.BinType]int64, error) {
	counts := map[models.BinType]int64{
		models.BinOrganic:   0,
		models.BinAnorganic: 0,
	}
	rows, err := r.pool.Query(ctx, `
		SELECT waste_type, COUNT(*) FROM detection_events
		WHERE received_at >= $1
		GROUP BY waste_type
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bt models.BinType
		var c int64
		if err := rows.Scan(&bt, &c); err != nil {
			return nil, err
		}
		counts[bt] = c
	}
	return counts, rows.Err()
}

// StatsSince returns total count, per-type counts, and average confidence
// over the given window. AvgConfidence is nil when there are zero rows,
// preserving SQL NULL rather than coercing to a fabricated 0.0.
func (r *DetectionRepo) StatsSince(ctx context.Context, since time.Time, deviceID *int64) (Stats, error) {
	where := "WHERE received_at >= $1"
	args := []any{since}
	if deviceID != nil {
		where += " AND device_id = $2"
		args = append(args, *deviceID)
	}

	byType, err := r.countsSinceWhere(ctx, where, args)
	if err != nil {
		return Stats{}, err
	}

	var total int64
	var avgConfidence *float64
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*), AVG(confidence) FROM detection_events `+where, args...,
	).Scan(&total, &avgConfidence)
	if err != nil {
		return Stats{}, err
	}

	return Stats{TotalDetections: total, ByType: byType, AvgConfidence: avgConfidence}, nil
}

func (r *DetectionRepo) countsSinceWhere(ctx context.Context, where string, args []any) (map[models.BinType]int64, error) {
	counts := map[models.BinType]int64{
		models.BinOrganic:   0,
		models.BinAnorganic: 0,
	}
	rows, err := r.pool.Query(ctx, `
		SELECT waste_type, COUNT(*) FROM detection_events `+where+` GROUP BY waste_type`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bt models.BinType
		var c int64
		if err := rows.Scan(&bt, &c); err != nil {
			return nil, err
		}
		counts[bt] = c
	}
	return counts, rows.Err()
}
