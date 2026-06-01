package messaging

import "time"

type EventCreatedMessage struct {
	EventID      string   `json:"event_id"`
	VehicleID    string   `json:"vehicle_id"`
	DeviceID     string   `json:"device_id"`
	EventType    string   `json:"event_type"`
	EventTime    string   `json:"event_time"`
	Speed        *float64 `json:"speed,omitempty"`
	BatteryLevel *float64 `json:"battery_level,omitempty"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`
}

func NewEventCreatedMessage(
	eventID string,
	vehicleID string,
	deviceID string,
	eventType string,
	eventTime time.Time,
	speed *float64,
	batteryLevel *float64,
	lat *float64,
	lon *float64,
) EventCreatedMessage {
	return EventCreatedMessage{
		EventID:      eventID,
		VehicleID:    vehicleID,
		DeviceID:     deviceID,
		EventType:    eventType,
		EventTime:    eventTime.Format(time.RFC3339),
		Speed:        speed,
		BatteryLevel: batteryLevel,
		Lat:          lat,
		Lon:          lon,
	}
}
