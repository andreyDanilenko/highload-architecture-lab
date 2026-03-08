# Redis Strategy: Stabilization and Consistency

This document describes how the Redis reserve strategy stays consistent with PostgreSQL, and what to do when things go wrong. No code — plan and concepts only.

---

## Two layers: Redis and PostgreSQL

- **Redis:** Fast check and atomic decrement (Lua). Reduces load on PG by rejecting "insufficient stock" early.
- **PostgreSQL:** Source of truth for durable state. We persist each reservation (transaction record + stock decrement by **delta**).

We do **not** treat Redis as the source of truth for the final balance. We apply the same **delta** (quantity to subtract) in both: Redis decrements by quantity; PG runs `stock_quantity = stock_quantity - quantity` so parallel requests do not overwrite each other.

---

## Compensating transaction (rollback Redis on PG failure)

If the PG write fails after we have already decremented in Redis (network error, deadlock, or PG has insufficient stock due to drift), Redis would be "ahead" (it already subtracted). To fix:

- In the service **catch** block: call `redisStore.increment(sku, quantity)` to restore the amount we reserved in Redis.
- So: PG failed ⇒ we did not actually complete the reservation ⇒ we put the quantity back in Redis. This keeps Redis aligned with the fact that no durable reservation was made.

---

## When Redis and PG can drift

1. **App crashed after Redis decrement, before PG commit.** Compensation never ran. Redis is lower than PG.
2. **Compensation failed** (e.g. Redis timeout in catch). Redis stays lower than PG for that SKU.
3. **Manual or external change in PG** (e.g. admin corrected stock). Redis was not updated.

In all cases, Redis may not match PG. Reads of "current stock" from PG (e.g. getBalance) will show PG; the Redis counter is used only for the fast reserve path.

---

## Reconciliation job (recommended)

Run a periodic job (e.g. every hour or after deployment):

- For each product (or each SKU that has a Redis key), read `stock_quantity` from PG.
- Set Redis: `redis.set(inventory:stock:{sku}, pgStock)`.

This resets Redis to PG’s truth. It does not fix "Redis was ahead" (we already compensate on failure); it fixes "Redis was behind or wrong" due to crashes or missed compensation.

Optional: only overwrite Redis if the key exists (to avoid creating keys for SKUs that never used the Redis path), or always set for a defined set of SKUs.

---

## Summary

| Situation | Action |
|-----------|--------|
| PG write fails after Redis decrement | **Compensate:** `increment(sku, quantity)` in Redis (in catch). |
| Redis and PG may have drifted (crash, missed compensation, admin edit) | **Reconcile:** periodic job sets Redis from PG (`redis.set(sku, pg.stock_quantity)`). |
| Parallel successful reserves | **Delta in PG:** use `decrementStockWithClient` (subtract quantity), not "write Redis balance to PG". |

See [subtask-4-redis.md](subtask-4-redis.md) for implementation details.
