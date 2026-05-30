package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type DeviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{
		pool: pool,
	}
}

func (r *DeviceRepository) Create(
	ctx context.Context,
	vehicleID uuid.UUID,
	externalID string,
	model *string,
) (domain.Device, error) {
	const query = `
		INSERT INTO devices (vehicle_id, external_id, model)
		VALUES ($1, $2, $3)
		RETURNING id, vehicle_id, external_id, model, status, last_seen_at, created_at, updated_at
	`

	var device domain.Device

	err := r.pool.QueryRow(ctx, query, vehicleID, externalID, model).Scan(
		&device.ID,
		&device.VehicleID,
		&device.ExternalID,
		&device.Model,
		&device.Status,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		return domain.Device{}, fmt.Errorf("create device: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Device, error) {
	const query = `
		SELECT id, vehicle_id, external_id, model, status, last_seen_at, created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	var device domain.Device

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&device.ID,
		&device.VehicleID,
		&device.ExternalID,
		&device.Model,
		&device.Status,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Device{}, apperrors.ErrNotFound
		}

		return domain.Device{}, fmt.Errorf("get device by id: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) GetByExternalID(ctx context.Context, externalID string) (domain.Device, error) {
	const query = `
		SELECT id, vehicle_id, external_id, model, status, last_seen_at, created_at, updated_at
		FROM devices
		WHERE external_id = $1
	`

	var device domain.Device

	err := r.pool.QueryRow(ctx, query, externalID).Scan(
		&device.ID,
		&device.VehicleID,
		&device.ExternalID,
		&device.Model,
		&device.Status,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Device{}, apperrors.ErrNotFound
		}

		return domain.Device{}, fmt.Errorf("get device by external id: %w", err)
	}

	return device, nil
}

func (r *DeviceRepository) List(ctx context.Context) ([]domain.Device, error) {
	const query = `
		SELECT id, vehicle_id, external_id, model, status, last_seen_at, created_at, updated_at
		FROM devices
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	devices := make([]domain.Device, 0)

	for rows.Next() {
		var device domain.Device

		if err := rows.Scan(
			&device.ID,
			&device.VehicleID,
			&device.ExternalID,
			&device.Model,
			&device.Status,
			&device.LastSeenAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}

		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices: %w", err)
	}

	return devices, nil
}

func (r *DeviceRepository) UpdateLastSeen(
	ctx context.Context,
	id uuid.UUID,
	lastSeenAt time.Time,
) error {
	const query = `
		UPDATE devices
		SET last_seen_at = $2,
		    updated_at = now()
		WHERE id = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, id, lastSeenAt)
	if err != nil {
		return fmt.Errorf("update device last seen: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
