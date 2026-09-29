#!/usr/bin/env bash
# Integration-test helper only — NOT part of the production app. Simulates
# one ESP32 heartbeat POST, shaped like the assumed ingest contract in
# internal/ingest/dto.go. Never run this against a production database.
set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
DEVICE_UID="${DEVICE_UID:-TEST-ESP32-0001}"

curl -sS -X POST "$HOST/api/v1/ingest/heartbeat" \
  -H "X-Device-UID: $DEVICE_UID" \
  -H "Content-Type: application/json" \
  -d '{
    "uptime_seconds": 120,
    "rssi": -55,
    "free_heap_bytes": 145200,
    "firmware_version": "test-1.0"
  }'
echo
