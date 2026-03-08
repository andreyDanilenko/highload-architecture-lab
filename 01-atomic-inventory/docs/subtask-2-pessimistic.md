# 2. Pessimistic Locking

**What:** Reserve with row lock inside a single transaction (SELECT FOR UPDATE).  
**Why:** Eliminates the race: 100 requests → stock 900, no lost updates. Cost: locks; slower under high contention.

---

## Implementation steps

1. Transaction helper: `withTransaction(pool, fn)` — get client, BEGIN, run fn(client), COMMIT or ROLLBACK, return client to pool.
2. Repository: `findBySkuWithLock(client, sku)` — SELECT FOR UPDATE by sku using the given client. `updateStockWithClient(client, sku, newQuantity)` — UPDATE using same client. `createWithClient(client, dto)` — INSERT into transactions via same client.
3. Service `reserveStockPessimistic`: idempotency as in naive. Inside withTransaction: load product with lock, check stock, compute newQuantity, update stock, insert transaction record, return result.
4. Route `POST /reserve/pessimistic` → controller → reserveStockPessimistic → { success, duplicated?, newBalance }.
5. Load test: 100 concurrent requests to /reserve/pessimistic. Expect: 100 successful, stock 900.

---

## What was done

- SELECT FOR UPDATE + UPDATE + INSERT in one transaction.
- withTransaction and repository methods that take client.
- Endpoint /reserve/pessimistic.
- Load test confirms: no lost updates (actual = 1000 - success).
