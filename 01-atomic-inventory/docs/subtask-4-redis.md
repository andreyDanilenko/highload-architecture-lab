# 4. Redis Atomic Counter

**What:** Stock in Redis, atomic deduction via Lua; PG for persistence and source of truth when key is cold.  
**Why:** High RPS; 100 requests (including with empty key) → stock 900, no race.

---

## Implementation steps

1. Redis key: `inventory:stock:{sku}`, value is a number. Interface: get/set and decrementIfSufficient; plus decrementIfSufficientOrInit(sku, initialValue, quantity) — single call for “no key → init from PG → deduct”.
2. Lua: “init + decrement” script — if key missing, SET to initialValue; then if current >= quantity, DECRBY and return new balance, else -1. One EVAL = one atomic operation.
3. Redis repository: call Lua via client.eval (or sendCommand), response >= 0 → new balance, else null. Implement decrementIfSufficientOrInit.
4. Service `reserveStockRedis`: idempotency by requestId in PG. Load product from PG (for validation and initialValue). Single call redisStore.decrementIfSufficientOrInit(sku, product.stockQuantity, quantity). On null — 409. Insert transaction record, update stock in products (sync to PG).
5. Redis connection (config REDIS_URL), DI: RedisStockRepository → InventoryService. Route `POST /reserve/redis`.
6. Load test: before run delete key (redis-cli DEL or docker exec inventory-redis redis-cli DEL). Reset DB, 100 requests to /reserve/redis. Expect: 100 successful, stock 900. In test script — fallback to docker exec if redis-cli not in PATH.

---

## What was done

- Lua “init + decrement” in one EVAL (no race on cold key).
- RedisStockRepository with decrementIfSufficientOrInit and sendCommand fallback.
- reserveStockRedis: one OrInit call, then write to PG and sync stock in products.
- Endpoint /reserve/redis.
- Load test with key cleanup (redis-cli or docker exec); 100 success, actual = 900.
