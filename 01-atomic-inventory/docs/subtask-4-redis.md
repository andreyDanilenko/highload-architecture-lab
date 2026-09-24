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
   - The current catch path attempts **compensation** with `increment(sku, quantity)`. This is best effort: a network/commit timeout can leave the PG outcome unknown, and compensation itself can fail.
5. **PG update — critical:** Do **not** write the absolute value from Redis to PG. Use a **delta** update: `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`. Repository method: `decrementStockWithClient(client, sku, quantity)`. If no row is updated (insufficient stock in PG), treat as error and run Redis compensation.
6. Redis connection (config REDIS_URL), DI: RedisStockRepository → InventoryService. Route `POST /reserve/redis`.
7. Load test: delete key before run if needed (redis-cli DEL or docker exec). Reset DB, 100 requests to /reserve/redis. Expect: 100 successful, stock 900.

---

## Why delta in PG, not absolute value from Redis

- If we write to PG the **absolute** balance computed in Redis (e.g. `newBalance = 7`), two concurrent requests can both decrement in Redis and then both write their own absolute value to PG. Whichever writes last **overwrites** the other; one deduction is lost. Example: A reserves 3 (Redis 10→7), B reserves 5 (Redis 7→2); B writes 2 to PG, then A writes its older balance 7 → final PG is 7, but it should be 2 (10−3−5).
- With **delta**: each request asks PG to "subtract my quantity from the current row". The SQL uses the column value: `stock_quantity = stock_quantity - quantity`. PG serializes updates; each request subtracts its delta from the current value, so all deductions are applied. No overwrite.

---

## Compensating transaction

The current service `catch` block attempts `await redisStore.increment(sku, quantity)`. This can restore the counter after a confirmed PG rollback, but does not guarantee synchronization: the process or compensation can fail. A timeout during commit is an unknown outcome, not proof of rollback; a future recovery flow must resolve it by request ID before compensating.

---

## What was done

- Lua "init + decrement" in one EVAL (atomic inside Redis). The PG value was read earlier, so a cold-key initialization can still use stale stock.
- RedisStockRepository: `decrementIfSufficientOrInit`, `increment` (for compensation).
- `reserveStockRedis`: Redis decrement first; then PG transaction with **decrementStockWithClient** (delta), not updateStockWithClient(absolute). On PG failure, Redis compensation via `increment`.
- Product repository: `decrementStockWithClient(client, sku, quantity)` — `UPDATE ... stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`.
- Endpoint /reserve/redis.
- Load test with key cleanup; 100 success, actual stock = 900.

---

## Reconciliation (recommended)

Redis can drift (e.g. server crashed before compensation ran). Reconciliation is a planned recovery step. Reading PG and blindly setting Redis during live reservations can overwrite in-flight decrements; coordinate or pause writes for the repair and verify the resulting balance. See [redis-stabilization-plan.md](redis-stabilization-plan.md).
