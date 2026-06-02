package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type AlertRepository struct {
	pool *pgxpool.Pool
}

func NewAlertRepository(pool *pgxpool.Pool) *AlertRepository {
	return &AlertRepository{
		pool: pool,
	}
}

func (r *AlertRepository) Create(
	ctx context.Context,
	vehicleID uuid.UUID,
	deviceID uuid.UUID,
	eventID *uuid.UUID,
	alertType string,
	severity string,
	message string,
) (domain.Alert, error) {
	const query = `
		INSERT INTO alerts (
			vehicle_id,
			device_id,
			event_id,
			type,
			severity,
			message
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			vehicle_id,
			device_id,
			event_id,
			type,
			severity,
			message,
			status,
			created_at,
			resolved_at
	`

	var alert domain.Alert

	err := r.pool.QueryRow(
		ctx,
		query,
		vehicleID,
		deviceID,
		eventID,
		alertType,
		severity,
		message,
	).Scan(
		&alert.ID,
		&alert.VehicleID,
		&alert.DeviceID,
		&alert.EventID,
		&alert.Type,
		&alert.Severity,
		&alert.Message,
		&alert.Status,
		&alert.CreatedAt,
		&alert.ResolvedAt,
	)
	if err != nil {
		return domain.Alert{}, fmt.Errorf("create alert: %w", err)
	}

	return alert, nil
}

func (r *AlertRepository) List(ctx context.Context, filter domain.AlertListFilter) ([]domain.Alert, error) {
	query := `
		SELECT
			id,
			vehicle_id,
			device_id,
			event_id,
			type,
			severity,
			message,
			status,
			created_at,
			resolved_at
		FROM alerts
	`

	conditions := make([]string, 0)
	args := make([]any, 0)

	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}

	if filter.Type != "" {
		args = append(args, filter.Type)
		conditions = append(conditions, fmt.Sprintf("type = $%d", len(args)))
	}

	if filter.VehicleID != nil {
		args = append(args, *filter.VehicleID)
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, filter.Limit)
	limitPlaceholder := len(args)

	args = append(args, filter.Offset)
	offsetPlaceholder := len(args)

	query += fmt.Sprintf(
		" ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		limitPlaceholder,
		offsetPlaceholder,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)

	for rows.Next() {
		var alert domain.Alert

		if err := rows.Scan(
			&alert.ID,
			&alert.VehicleID,
			&alert.DeviceID,
			&alert.EventID,
			&alert.Type,
			&alert.Severity,
			&alert.Message,
			&alert.Status,
			&alert.CreatedAt,
			&alert.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate alerts: %w", err)
	}

	return alerts, nil
}

func (r *AlertRepository) Resolve(ctx context.Context, id uuid.UUID) (domain.Alert, error) {
	const query = `
		UPDATE alerts
		SET status = $2,
		    resolved_at = now()
		WHERE id = $1
		RETURNING
			id,
			vehicle_id,
			device_id,
			event_id,
			type,
			severity,
			message,
			status,
			created_at,
			resolved_at
	`

	var alert domain.Alert

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		domain.AlertStatusResolved,
	).Scan(
		&alert.ID,
		&alert.VehicleID,
		&alert.DeviceID,
		&alert.EventID,
		&alert.Type,
		&alert.Severity,
		&alert.Message,
		&alert.Status,
		&alert.CreatedAt,
		&alert.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Alert{}, apperrors.ErrNotFound
		}

		return domain.Alert{}, fmt.Errorf("resolve alert: %w", err)
	}

	return alert, nil
}
