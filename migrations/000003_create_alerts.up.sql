CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,

    type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_alerts_vehicle_status ON alerts(vehicle_id, status);
CREATE INDEX idx_alerts_status_created ON alerts(status, created_at DESC);
CREATE INDEX idx_alerts_event_id ON alerts(event_id);
