DROP INDEX IF EXISTS idx_devices_external_id;
DROP INDEX IF EXISTS idx_devices_vehicle_id;

DROP INDEX IF EXISTS idx_vehicles_plate_number;
DROP INDEX IF EXISTS idx_vehicles_fleet_id;

DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS fleets;

DROP EXTENSION IF EXISTS "pgcrypto";
