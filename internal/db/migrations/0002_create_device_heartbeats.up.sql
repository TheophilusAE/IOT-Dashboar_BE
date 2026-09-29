CREATE TABLE device_heartbeats (
    id              BIGSERIAL PRIMARY KEY,
    device_id       BIGINT NOT NULL REFERENCES devices(id),
    received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    device_ts       TIMESTAMPTZ,
    rssi            INTEGER,
    uptime_seconds  BIGINT,
    free_heap_bytes INTEGER
);
