CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,

    event_type TEXT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    lat DOUBLE PRECISION,
    lon DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    battery_level DOUBLE PRECISION,
    ignition BOOLEAN,

    payload JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_vehicle_time ON events(vehicle_id, event_time DESC);
CREATE INDEX idx_events_device_time ON events(device_id, event_time DESC);
CREATE INDEX idx_events_type ON events(event_type);
