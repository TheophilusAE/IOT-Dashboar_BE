CREATE TABLE sensor_readings (
    id           BIGSERIAL PRIMARY KEY,
    device_id    BIGINT NOT NULL REFERENCES devices(id),
    bin_type     bin_type NOT NULL,
    distance_cm  NUMERIC(6,2) NOT NULL,
    fill_percent NUMERIC(5,2),
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    device_ts    TIMESTAMPTZ,
    raw_payload  JSONB
);
