package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type VehicleRepository struct {
	pool *pgxpool.Pool
}

func NewVehicleRepository(pool *pgxpool.Pool) *VehicleRepository {
	return &VehicleRepository{
		pool: pool,
	}
}

func (r *VehicleRepository) Create(
	ctx context.Context,
	fleetID uuid.UUID,
	plateNumber string,
	vin *string,
	vehicleType string,
) (domain.Vehicle, error) {
	const query = `
		INSERT INTO vehicles (fleet_id, plate_number, vin, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, fleet_id, plate_number, vin, type, status, created_at, updated_at
	`

	var vehicle domain.Vehicle

	err := r.pool.QueryRow(ctx, query, fleetID, plateNumber, vin, vehicleType).Scan(
		&vehicle.ID,
		&vehicle.FleetID,
		&vehicle.PlateNumber,
		&vehicle.VIN,
		&vehicle.Type,
		&vehicle.Status,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		return domain.Vehicle{}, fmt.Errorf("create vehicle: %w", err)
	}

	return vehicle, nil
}

func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Vehicle, error) {
	const query = `
		SELECT id, fleet_id, plate_number, vin, type, status, created_at, updated_at
		FROM vehicles
		WHERE id = $1
	`

	var vehicle domain.Vehicle

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&vehicle.ID,
		&vehicle.FleetID,
		&vehicle.PlateNumber,
		&vehicle.VIN,
		&vehicle.Type,
		&vehicle.Status,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Vehicle{}, apperrors.ErrNotFound
		}

		return domain.Vehicle{}, fmt.Errorf("get vehicle by id: %w", err)
	}

	return vehicle, nil
}

func (r *VehicleRepository) List(ctx context.Context) ([]domain.Vehicle, error) {
	const query = `
		SELECT id, fleet_id, plate_number, vin, type, status, created_at, updated_at
		FROM vehicles
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer rows.Close()

	vehicles := make([]domain.Vehicle, 0)

	for rows.Next() {
		var vehicle domain.Vehicle

		if err := rows.Scan(
			&vehicle.ID,
			&vehicle.FleetID,
			&vehicle.PlateNumber,
			&vehicle.VIN,
			&vehicle.Type,
			&vehicle.Status,
			&vehicle.CreatedAt,
			&vehicle.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}

		vehicles = append(vehicles, vehicle)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicles: %w", err)
	}

	return vehicles, nil
}
