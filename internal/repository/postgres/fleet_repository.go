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

type FleetRepository struct {
	pool *pgxpool.Pool
}

func NewFleetRepository(pool *pgxpool.Pool) *FleetRepository {
	return &FleetRepository{
		pool: pool,
	}
}

func (r *FleetRepository) Create(ctx context.Context, name string) (domain.Fleet, error) {
	const query = `
		INSERT INTO fleets (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`

	var fleet domain.Fleet

	err := r.pool.QueryRow(ctx, query, name).Scan(
		&fleet.ID,
		&fleet.Name,
		&fleet.CreatedAt,
		&fleet.UpdatedAt,
	)
	if err != nil {
		return domain.Fleet{}, fmt.Errorf("create fleet: %w", err)
	}

	return fleet, nil
}

func (r *FleetRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Fleet, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM fleets
		WHERE id = $1
	`

	var fleet domain.Fleet

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&fleet.ID,
		&fleet.Name,
		&fleet.CreatedAt,
		&fleet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Fleet{}, apperrors.ErrNotFound
		}

		return domain.Fleet{}, fmt.Errorf("get fleet by id: %w", err)
	}

	return fleet, nil
}

func (r *FleetRepository) List(ctx context.Context) ([]domain.Fleet, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM fleets
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list fleets: %w", err)
	}
	defer rows.Close()

	fleets := make([]domain.Fleet, 0)

	for rows.Next() {
		var fleet domain.Fleet

		if err := rows.Scan(
			&fleet.ID,
			&fleet.Name,
			&fleet.CreatedAt,
			&fleet.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan fleet: %w", err)
		}

		fleets = append(fleets, fleet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fleets: %w", err)
	}

	return fleets, nil
}
