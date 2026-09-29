CREATE TABLE devices (
    id               BIGSERIAL PRIMARY KEY,
    device_uid       TEXT NOT NULL UNIQUE,
    display_name     TEXT NOT NULL DEFAULT '',
    location         TEXT,
    firmware_version TEXT,
    registered_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active        BOOLEAN NOT NULL DEFAULT true
);
