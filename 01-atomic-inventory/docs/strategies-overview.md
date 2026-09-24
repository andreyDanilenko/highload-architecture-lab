# Reservation Strategies Overview

Short comparison of all four strategies: how each works, how it avoids lost updates, and when to use it.

---

## 1. Naive

- **Flow:** Read product from PG → check stock → compute `newQuantity = stock - quantity` → single `UPDATE products SET stock_quantity = newQuantity` (no lock across read/compute/write) → insert transaction record.
- **PG write:** Absolute value (`newQuantity`). UPDATE takes a row lock, but the earlier read is unprotected and there is no version check.
- **Race:** Two requests can read the same stock (e.g. 10), both compute their new value (7 and 5), both run UPDATE. The last write wins; one deduction is lost.
- **Use:** Demo and load-test only, to show lost updates. Do not use in production.

---

## 2. Pessimistic

- **Flow:** Inside one PG transaction: `SELECT ... FOR UPDATE` (lock row) → check stock → compute `newQuantity` → `UPDATE stock` with same client → insert transaction → commit.
- **PG write:** Absolute value, but the row is locked. Only one transaction at a time can hold the lock for that SKU.
- **Race:** No lost update — the second request waits for the lock, then reads the already-updated value (e.g. 7) and writes its result (2). Updates are serialized.
- **Cost:** Locks block other transactions; under high contention, latency grows. Deadlocks possible if different requests lock rows in different order.
- **Use:** When competing writers follow the same row-locking protocol and lock contention is acceptable. Cross-row and cross-store invariants need separate analysis.

---

## 3. Optimistic

- **Flow:** No long-held lock. Loop: read product (with `version`) → check stock → `UPDATE ... SET stock_quantity = newQuantity, version = version + 1 WHERE sku = $1 AND version = $2`. If no row updated (version changed), retry; otherwise insert transaction.
- **PG write:** Absolute value, but only if `version` has not changed. So we overwrite only when nobody else updated the row since we read it.
- **Race:** If someone else updated the row, our UPDATE touches 0 rows → we retry with a fresh read. No lost update; multiple retries under contention.
- **Cost:** Many concurrent updates to the same SKU can exhaust `maxOptimisticRetries`. Good when conflict rate is low.
- **Use:** When conflicts are rare and you want to avoid holding a lock during the read/compute phase. The conditional UPDATE still uses database locks.

---

## 4. Redis

- **Flow:** Decrement in Redis first (Lua: initialize from a value previously read from PG if the key is missing, then decrement if sufficient). Then inside a PG transaction: insert transaction record → **decrement in PG by delta** (`stock_quantity = stock_quantity - quantity`), not by writing the Redis balance. If PG fails (or PG has insufficient stock), **compensate:** `increment(sku, quantity)` in Redis.
- **PG write:** **Delta only:** `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE sku = $2 AND stock_quantity >= $1`. So we do not write the absolute value from Redis; we ask PG to subtract the same quantity. Parallel requests each subtract their delta; no overwrite.
- **Race:** Redis serializes its own decrements (Lua is atomic). In PG, each request subtracts its quantity from the current row value; the DB serializes row updates, so committed deltas do not overwrite each other. This does not make Redis and PG one atomic transaction.
- **Consistency:** A confirmed PG rollback can be followed by best-effort Redis compensation. A timeout may hide a committed transaction; first resolve it by request ID. Compensation can fail, and reconciliation must coordinate with in-flight reservations rather than blindly overwrite Redis.
- **Use:** High RPS; Redis as a fast filter; PG remains source of truth. See [subtask-4-redis.md](subtask-4-redis.md) and [redis-stabilization-plan.md](redis-stabilization-plan.md).

---

## Summary table

| Strategy    | PG update style     | Concurrency fix              | Main trade-off                    |
|------------|---------------------|------------------------------|-----------------------------------|
| Naive      | Absolute, unguarded read| None                         | Lost updates; demo only           |
| Pessimistic| Absolute, with lock | Row lock (FOR UPDATE)         | Blocking under contention         |
| Optimistic | Absolute, version    | Version check + retry         | Retries under contention          |
| Redis      | **Delta** (subtract) | Delta in PG + Redis compensate| Two stores; need reconciliation   |
