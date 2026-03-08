# 1. Naive Reserve

**What:** Reserve via read-modify-write with no locking.  
**Why:** Demonstrates the race: under 100 concurrent requests some updates are lost (e.g. 1000 → 915). Do not use in production.

---

## Implementation steps

1. DB schema: `products` (sku, stock_quantity, version), `inventory_transactions` (sku, quantity, request_id UNIQUE). Seed: SKU-TEST-001 with stock 1000.
2. Endpoint `POST /reserve`: body `{ sku, quantity, requestId }`, validation (Zod).
3. Idempotency: look up by requestId in transactions; if found and sku/quantity match — return 200 without deducting again; on mismatch — 409.
4. Service: load product, check stock, compute newQuantity = stock - quantity, single UPDATE with no lock or version, INSERT into transactions.
5. Repository: `updateStockNaive(sku, newQuantity)` — one UPDATE by sku, no FOR UPDATE and no version check.
6. Route and controller: parse body → reserveStock → response { success, duplicated?, newBalance }. Errors via shared handler (404, 409).
7. Load test: reset DB, 100 concurrent POST /reserve. Expect: actual stock > (1000 - success) — visible lost updates.

---

## What was done

- Endpoint `/reserve` with naive read-modify-write.
- Idempotency by requestId.
- updateStockNaive with no locking.
- Load-test script that shows lost updates.
- Marked in code and README: demo and tests only.
