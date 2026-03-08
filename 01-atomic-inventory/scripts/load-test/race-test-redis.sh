#!/bin/bash
# Redis atomic counter — expect no lost updates.
# Ensure DB is reset (e.g. scripts/reset-db.sh) and Redis key is cleared so first request seeds from PG.
REDIS_KEY="inventory:stock:SKU-TEST-001"
if command -v redis-cli &>/dev/null; then
  redis-cli DEL "$REDIS_KEY" 2>/dev/null || true
elif docker ps --format '{{.Names}}' 2>/dev/null | grep -q '^inventory-redis$'; then
  docker exec inventory-redis redis-cli DEL "$REDIS_KEY" 2>/dev/null || true
fi
cd "$(dirname "$0")"
./race-test.sh "${1:-http://localhost:3000}" "/api/v1/inventory/reserve/redis"
