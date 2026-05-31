package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool: pool,
	}
}

func (r *EventRepository) Create(
	ctx context.Context,
	deviceID uuid.UUID,
	vehicleID uuid.UUID,
	eventType string,
	eventTime time.Time,
	lat *float64,
	lon *float64,
	speed *float64,
	batteryLevel *float64,
	ignition *bool,
	payload json.RawMessage,
) (domain.Event, error) {
	const query = `
		INSERT INTO events (
			device_id,
			vehicle_id,
			event_type,
			event_time,
			lat,
			lon,
			speed,
			battery_level,
			ignition,
			payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING
			id,
			device_id,
			vehicle_id,
			event_type,
			event_time,
			received_at,
			lat,
			lon,
			speed,
			battery_level,
			ignition,
			payload,
			created_at
	`

	var event domain.Event

	err := r.pool.QueryRow(
		ctx,
		query,
		deviceID,
		vehicleID,
		eventType,
		eventTime,
		lat,
		lon,
		speed,
		batteryLevel,
		ignition,
		payload,
	).Scan(
		&event.ID,
		&event.DeviceID,
		&event.VehicleID,
		&event.EventType,
		&event.EventTime,
		&event.ReceivedAt,
		&event.Lat,
		&event.Lon,
		&event.Speed,
		&event.BatteryLevel,
		&event.Ignition,
		&event.Payload,
		&event.CreatedAt,
	)
	if err != nil {
		return domain.Event{}, fmt.Errorf("create event: %w", err)
	}

	return event, nil
}
