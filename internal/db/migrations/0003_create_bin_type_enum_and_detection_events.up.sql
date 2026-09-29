CREATE TYPE bin_type AS ENUM ('organic', 'anorganic');

CREATE TABLE detection_events (
    id          BIGSERIAL PRIMARY KEY,
    device_id   BIGINT NOT NULL REFERENCES devices(id),
    waste_type  bin_type NOT NULL,
    confidence  NUMERIC(5,4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    device_ts   TIMESTAMPTZ,
    raw_payload JSONB
);
