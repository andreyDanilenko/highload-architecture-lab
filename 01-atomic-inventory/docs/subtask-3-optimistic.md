# 3. Optimistic Locking

**What:** Reserve without long-held locks: read row with version, update only if version unchanged; retry on conflict.  
**Why:** Correct outcome (1000 → 900) with fewer locks; faster than pessimistic when contention is low.

---

## Implementation steps

1. `products` already has version column. Repository: `updateStock(sku, newQuantity, expectedVersion)` — UPDATE with WHERE version = $expectedVersion, SET version = version + 1. Return “exactly one row updated”.
2. Config: maxOptimisticRetries (e.g. 5).
3. Service `reserveStockOptimistic`: idempotency as before. Loop: load product (findBySku), check stock, call updateStock(sku, newQuantity, product.version); if updated — create transaction, return result; else retry. After N attempts — error “too many retries”.
4. Route `POST /reserve/optimistic` → controller → reserveStockOptimistic.
5. Load test: 100 requests to /reserve/optimistic. Expect: 100 successful, stock 900.

---

## What was done

- UPDATE with version check and version increment.
- Retry loop with maxOptimisticRetries limit.
- Endpoint /reserve/optimistic.
- Load test: no lost updates; fewer locks than pessimistic under low contention.
