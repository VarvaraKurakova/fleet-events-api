package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/VarvaraKurakova/fleet-events-api/internal/apperrors"
	"github.com/VarvaraKurakova/fleet-events-api/internal/domain"
)

const vehicleStateTTL = 24 * time.Hour

type VehicleStateCache struct {
	client *goredis.Client
}

func NewVehicleStateCache(client *goredis.Client) *VehicleStateCache {
	return &VehicleStateCache{
		client: client,
	}
}

type vehicleStateDTO struct {
	VehicleID    string   `json:"vehicle_id"`
	DeviceID     string   `json:"device_id"`
	EventID      string   `json:"event_id"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	BatteryLevel *float64 `json:"battery_level,omitempty"`
	Ignition     *bool    `json:"ignition,omitempty"`
	EventTime    string   `json:"event_time"`
	ReceivedAt   string   `json:"received_at"`
}

func (c *VehicleStateCache) Set(ctx context.Context, state domain.VehicleState) error {
	key := vehicleStateKey(state.VehicleID)

	dto := vehicleStateDTO{
		VehicleID:    state.VehicleID.String(),
		DeviceID:     state.DeviceID.String(),
		EventID:      state.EventID.String(),
		Lat:          state.Lat,
		Lon:          state.Lon,
		Speed:        state.Speed,
		BatteryLevel: state.BatteryLevel,
		Ignition:     state.Ignition,
		EventTime:    state.EventTime.Format(time.RFC3339),
		ReceivedAt:   state.ReceivedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal vehicle state: %w", err)
	}

	if err := c.client.Set(ctx, key, data, vehicleStateTTL).Err(); err != nil {
		return fmt.Errorf("set vehicle state: %w", err)
	}

	return nil
}

func (c *VehicleStateCache) Get(ctx context.Context, vehicleID uuid.UUID) (domain.VehicleState, error) {
	key := vehicleStateKey(vehicleID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.VehicleState{}, apperrors.ErrNotFound
		}

		return domain.VehicleState{}, fmt.Errorf("get vehicle state: %w", err)
	}

	var dto vehicleStateDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return domain.VehicleState{}, fmt.Errorf("unmarshal vehicle state: %w", err)
	}

	state, err := dto.toDomain()
	if err != nil {
		return domain.VehicleState{}, err
	}

	return state, nil
}

func vehicleStateKey(vehicleID uuid.UUID) string {
	return fmt.Sprintf("vehicle_state:%s", vehicleID.String())
}

func (d vehicleStateDTO) toDomain() (domain.VehicleState, error) {
	vehicleID, err := uuid.Parse(d.VehicleID)
	if err != nil {
		return domain.VehicleState{}, fmt.Errorf("parse vehicle id: %w", err)
	}

	deviceID, err := uuid.Parse(d.DeviceID)
	if err != nil {
		return domain.VehicleState{}, fmt.Errorf("parse device id: %w", err)
	}

	eventID, err := uuid.Parse(d.EventID)
	if err != nil {
		return domain.VehicleState{}, fmt.Errorf("parse event id: %w", err)
	}

	eventTime, err := time.Parse(time.RFC3339, d.EventTime)
	if err != nil {
		return domain.VehicleState{}, fmt.Errorf("parse event time: %w", err)
	}

	receivedAt, err := time.Parse(time.RFC3339, d.ReceivedAt)
	if err != nil {
		return domain.VehicleState{}, fmt.Errorf("parse received at: %w", err)
	}

	return domain.VehicleState{
		VehicleID:    vehicleID,
		DeviceID:     deviceID,
		EventID:      eventID,
		Lat:          d.Lat,
		Lon:          d.Lon,
		Speed:        d.Speed,
		BatteryLevel: d.BatteryLevel,
		Ignition:     d.Ignition,
		EventTime:    eventTime,
		ReceivedAt:   receivedAt,
	}, nil
}
