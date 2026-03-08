#!/bin/bash
# Optimistic locking (version + retry) — expect no lost updates.
cd "$(dirname "$0")"
./race-test.sh "${1:-http://localhost:3000}" "/api/v1/inventory/reserve/optimistic"
