# 4. Redis Atomic Counter

**What:** Redis as a fast layer for check-and-decrement; PostgreSQL remains the source of truth. Atomic decrement in Redis (Lua); persist to PG using **delta only** (no overwrite).  
**Why:** High RPS; most "insufficient stock" requests are rejected in Redis without touching PG. Under concurrency, PG must apply the same **delta** (subtract quantity), not the absolute balance from Redis, so parallel requests do not overwrite each other.

---

## Implementation steps

1. **Redis key:** `inventory:stock:{sku}`, value is a number. Interface: get, set, and `decrementIfSufficientOrInit(sku, initialValue, quantity)` — single call for "no key → init from PG → deduct if sufficient".
2. **Lua script:** "init + decrement" — if key missing, SET to initialValue; then if current >= quantity, DECRBY and return new balance, else return -1. One EVAL = one atomic operation in Redis.
3. **Redis repository:** Call Lua via client.eval (or sendCommand). Response >= 0 → new balance; negative or invalid → null. Implement `decrementIfSufficientOrInit` and `increment` (for compensation).
4. **Service `reserveStockRedis`:**
   - Idempotency by requestId in PG.
   - Load product from PG (for validation and initialValue when key is cold).
   - Single call `redisStore.decrementIfSufficientOrInit(sku, product.stockQuantity, quantity)`. On null → 409 InsufficientStock.
   - Inside a PG transaction: create transaction record, then **decrement in PG by delta** (see below), not by writing the Redis balance.
   - On any PG failure (network, constraint, or insufficient stock in PG): **compensating transaction** — `increment(sku, quantity)` in Redis so the reserved amount is restored.
5. **PG update — critical:** Do **not** write the absolute value from Redis to PG. Use a **delta** update: `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`. Repository method: `decrementStockWithClient(client, sku, quantity)`. If no row is updated (insufficient stock in PG), treat as error and run Redis compensation.
6. Redis connection (config REDIS_URL), DI: RedisStockRepository → InventoryService. Route `POST /reserve/redis`.
7. Load test: delete key before run if needed (redis-cli DEL or docker exec). Reset DB, 100 requests to /reserve/redis. Expect: 100 successful, stock 900.

---

## Why delta in PG, not absolute value from Redis

- If we write to PG the **absolute** balance computed in Redis (e.g. `newBalance = 7`), two concurrent requests can both decrement in Redis and then both write their own absolute value to PG. Whichever writes last **overwrites** the other; one deduction is lost. Example: A reserves 3 (Redis 10→7), B reserves 5 (Redis 10→5); A writes 7 to PG, B writes 5 to PG → final PG is 5, but it should be 2 (10−3−5).
- With **delta**: each request asks PG to "subtract my quantity from the current row". The SQL uses the column value: `stock_quantity = stock_quantity - quantity`. PG serializes updates; each request subtracts its delta from the current value, so all deductions are applied. No overwrite.

---

## Compensating transaction

If the PG transaction fails (timeout, deadlock, or `decrementStockWithClient` returns false because stock in PG was insufficient), we must undo the Redis decrement. In the service `catch` block: `await redisStore.increment(sku, quantity)`. This keeps Redis in sync with the fact that the reservation did not complete in PG.

---

## What was done

- Lua "init + decrement" in one EVAL (no race on cold key).
- RedisStockRepository: `decrementIfSufficientOrInit`, `increment` (for compensation).
- `reserveStockRedis`: Redis decrement first; then PG transaction with **decrementStockWithClient** (delta), not updateStockWithClient(absolute). On PG failure, Redis compensation via `increment`.
- Product repository: `decrementStockWithClient(client, sku, quantity)` — `UPDATE ... stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`.
- Endpoint /reserve/redis.
- Load test with key cleanup; 100 success, actual stock = 900.

---

## Reconciliation (recommended)

Redis can drift (e.g. server crashed before compensation ran). A background job should periodically sync Redis from PG: for each SKU, `redis.set(sku, pg.stock_quantity)`. See [redis-stabilization-plan.md](redis-stabilization-plan.md).
