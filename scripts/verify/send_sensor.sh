#!/usr/bin/env bash
# Integration-test helper only — NOT part of the production app. Simulates
# one ultrasonic reading cycle for both bins, shaped like the assumed ingest
# contract in internal/ingest/dto.go. Never run this against a production
# database.
set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
DEVICE_UID="${DEVICE_UID:-TEST-ESP32-0001}"
ORGANIC_DISTANCE_CM="${1:-22.5}"
ANORGANIC_DISTANCE_CM="${2:-40.1}"

curl -sS -X POST "$HOST/api/v1/ingest/sensor" \
  -H "X-Device-UID: $DEVICE_UID" \
  -H "Content-Type: application/json" \
  -d "{
    \"readings\": [
      { \"bin_type\": \"organic\",   \"distance_cm\": $ORGANIC_DISTANCE_CM },
      { \"bin_type\": \"anorganic\", \"distance_cm\": $ANORGANIC_DISTANCE_CM }
    ]
  }"
echo
