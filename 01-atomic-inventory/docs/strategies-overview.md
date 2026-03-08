# Reservation Strategies Overview

Short comparison of all four strategies: how each works, how it avoids lost updates, and when to use it.

---

## 1. Naive

- **Flow:** Read product from PG → check stock → compute `newQuantity = stock - quantity` → single `UPDATE products SET stock_quantity = newQuantity` (no lock) → insert transaction record.
- **PG write:** Absolute value (`newQuantity`). No lock, no version.
- **Race:** Two requests can read the same stock (e.g. 10), both compute their new value (7 and 5), both run UPDATE. The last write wins; one deduction is lost.
- **Use:** Demo and load-test only, to show lost updates. Do not use in production.

---

## 2. Pessimistic

- **Flow:** Inside one PG transaction: `SELECT ... FOR UPDATE` (lock row) → check stock → compute `newQuantity` → `UPDATE stock` with same client → insert transaction → commit.
- **PG write:** Absolute value, but the row is locked. Only one transaction at a time can hold the lock for that SKU.
- **Race:** No lost update — the second request waits for the lock, then reads the already-updated value (e.g. 7) and writes its result (2). Updates are serialized.
- **Cost:** Locks block other transactions; under high contention, latency grows. Deadlocks possible if different requests lock rows in different order.
- **Use:** When you need strong consistency and can accept lock contention.

---

## 3. Optimistic

- **Flow:** No long-held lock. Loop: read product (with `version`) → check stock → `UPDATE ... SET stock_quantity = newQuantity, version = version + 1 WHERE sku = $1 AND version = $2`. If no row updated (version changed), retry; otherwise insert transaction.
- **PG write:** Absolute value, but only if `version` has not changed. So we overwrite only when nobody else updated the row since we read it.
- **Race:** If someone else updated the row, our UPDATE touches 0 rows → we retry with a fresh read. No lost update; multiple retries under contention.
- **Cost:** Many concurrent updates to the same SKU can exhaust `maxOptimisticRetries`. Good when conflict rate is low.
- **Use:** When you want to avoid locks and conflicts are rare.

---

## 4. Redis

- **Flow:** Decrement in Redis first (Lua: init from PG if key missing, then decrement if sufficient). Then inside a PG transaction: insert transaction record → **decrement in PG by delta** (`stock_quantity = stock_quantity - quantity`), not by writing the Redis balance. If PG fails (or PG has insufficient stock), **compensate:** `increment(sku, quantity)` in Redis.
- **PG write:** **Delta only:** `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`. So we do not write the absolute value from Redis; we ask PG to subtract the same quantity. Parallel requests each subtract their delta; no overwrite.
- **Race:** Redis serializes its own decrements (Lua is atomic). In PG, each request subtracts its quantity from the current row value; the DB serializes row updates, so all deltas are applied. No lost update.
- **Consistency:** If PG write fails, we roll back Redis via `increment`. Redis can still drift (e.g. crash before compensation); a reconciliation job (periodically set Redis from PG) is recommended.
- **Use:** High RPS; Redis as a fast filter; PG remains source of truth. See [subtask-4-redis.md](subtask-4-redis.md) and [redis-stabilization-plan.md](redis-stabilization-plan.md).

---

## Summary table

| Strategy    | PG update style     | Concurrency fix              | Main trade-off                    |
|------------|---------------------|------------------------------|-----------------------------------|
| Naive      | Absolute, no lock   | None                         | Lost updates; demo only           |
| Pessimistic| Absolute, with lock | Row lock (FOR UPDATE)         | Blocking under contention         |
| Optimistic | Absolute, version    | Version check + retry         | Retries under contention          |
| Redis      | **Delta** (subtract) | Delta in PG + Redis compensate| Two stores; need reconciliation   |
