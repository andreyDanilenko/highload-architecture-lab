# Redis Strategy: Stabilization and Consistency

This document describes drift risks and planned recovery for the Redis reserve strategy. Redis and PostgreSQL do not form one atomic transaction. No code — plan and concepts only.

---

## Two layers: Redis and PostgreSQL

- **Redis:** Fast check and atomic decrement (Lua). Reduces load on PG by rejecting "insufficient stock" early.
- **PostgreSQL:** Source of truth for durable state. We persist each reservation (transaction record + stock decrement by **delta**).

We do **not** treat Redis as the source of truth for the final balance. We apply the same **delta** (quantity to subtract) in both: Redis decrements by quantity; PG runs `stock_quantity = stock_quantity - quantity` so parallel requests do not overwrite each other.

---

## Compensating transaction (rollback Redis on PG failure)

After a confirmed PG rollback, Redis may still contain the earlier decrement. The current best-effort recovery is:

- In the service **catch** block: call `redisStore.increment(sku, quantity)` to restore the amount we reserved in Redis.
- A timeout during commit does not establish rollback. Resolve an unknown PG outcome using the durable request ID before compensating; otherwise a committed reservation can be credited back in Redis. Compensation can also fail or be repeated.

---

## When Redis and PG can drift

1. **App crashed after Redis decrement, before PG commit.** Compensation never ran. Redis is lower than PG.
2. **Compensation failed** (e.g. Redis timeout in catch). Redis stays lower than PG for that SKU.
3. **Manual or external change in PG** (e.g. admin corrected stock). Redis was not updated.
4. **Commit outcome unknown or compensation repeated.** Redis can become higher than PG.

In all cases, Redis may not match PG. Reads of "current stock" from PG (e.g. getBalance) will show PG; the Redis counter is used only for the fast reserve path.

---

## Reconciliation job (recommended)

For a controlled repair, first pause and drain reservations for the SKU (including writes through other strategies), and resolve unknown operation outcomes:

- For each product (or each SKU that has a Redis key), read `stock_quantity` from PG.
- Set Redis: `redis.set(inventory:stock:{sku}, pgStock)`.

With writes quiesced, this copies the authoritative balance. A periodic read-and-SET while reservations continue can overwrite fresh decrements; online reconciliation requires coordination/versioning and is not implemented by this recipe.

Optional: only overwrite Redis if the key exists (to avoid creating keys for SKUs that never used the Redis path), or always set for a defined set of SKUs.

---

## Summary

| Situation | Action |
|-----------|--------|
| Confirmed PG rollback after Redis decrement | Attempt compensation; record and retry recovery safely if it fails. |
| PG commit outcome unknown | Resolve by request ID before deciding whether to compensate. |
| Redis and PG may have drifted | Reconcile after draining writes, or design a coordinated online repair; blind periodic SET is unsafe under concurrent reservations. |
| Parallel successful reserves | **Delta in PG:** use `decrementStockWithClient` (subtract quantity), not "write Redis balance to PG". |

See [subtask-4-redis.md](subtask-4-redis.md) for implementation details.
