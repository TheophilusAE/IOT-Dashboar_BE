#!/usr/bin/env bash
# Integration-test helper only — NOT part of the production app. Simulates
# one AI waste-classification event, shaped like the assumed ingest contract
# in internal/ingest/dto.go. Never run this against a production database.
set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
DEVICE_UID="${DEVICE_UID:-TEST-ESP32-0001}"
WASTE_TYPE="${1:-organic}"
CONFIDENCE="${2:-0.947}"

curl -sS -X POST "$HOST/api/v1/ingest/detection" \
  -H "X-Device-UID: $DEVICE_UID" \
  -H "Content-Type: application/json" \
  -d "{
    \"waste_type\": \"$WASTE_TYPE\",
    \"confidence\": $CONFIDENCE
  }"
echo
