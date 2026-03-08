#!/bin/bash
# Naive reserve (no locking) — expect race condition / lost updates under load.
cd "$(dirname "$0")"
./race-test.sh "${1:-http://localhost:3000}" "/api/v1/inventory/reserve"
