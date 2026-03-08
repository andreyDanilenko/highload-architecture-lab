# Redis: Compensating Transaction and Stabilization Plan

Description of the task, risks, and plan for improving strategy 4 (Redis) without tying it to code.

---

## Task (description)

Strategy 4 (Redis) gives maximum reserve speed by doing an atomic decrement in Redis, then syncing the result to PostgreSQL. To keep the system consistent on failures, we need a **compensating transaction** pattern:

1. **Happy path:** decrement in Redis (atomically) → save transaction and update stock in the DB. Redis and PG match.
2. **Error on DB write:** after a successful decrement in Redis, the write to PostgreSQL fails (timeout, disk, deadlock). Without a rollback in Redis, cache stock would be lower than in the DB — desync. So in `catch` we must add the quantity back in Redis (method `increment(sku, quantity)`). That way only the operation that never committed in the DB is rolled back.
3. **Critical failure:** if both Redis and the app go down (e.g. `increment` in `catch` never ran due to lost connection to Redis), consistency is restored by **background reconciliation**: PostgreSQL is the source of truth; a worker periodically compares stock in the DB and Redis and fixes Redis when needed. Result: eventual consistency.

The `increment` method in the Redis store implements the rollback (Redis `INCRBY`). Without it, the compensating transaction is not possible.

---

## Potential issues (Redis strategy)

- **`increment` failing in the `catch` block** (no connection to Redis, timeout): compensation will not run, Redis will stay undercounted. Solution — reconciliation: a background process compares PG and Redis and aligns stock; source of truth is the DB.
- **Double rollback:** if `increment` is called twice by mistake for the same failed operation, Redis stock will be overstated. Compensation must be idempotent (e.g. a rollbacks table keyed by `requestId`, or a flag on the transaction).
- **Order of operations:** compensation must run only when the decrement in Redis has already happened and the PG write has not. Otherwise we risk rolling back something that was never written to the DB.
- **Syncing PG after Redis:** the current implementation writes to PG after each reserve. Under very high load, batching or async write to a queue with later PG write could be considered — with the understanding that PG will lag until that write (read from Redis for reserves, from PG for reports after reconciliation).

---

## Stabilization plan (implementation)

Steps only, no code.

1. **Compensating transaction in reserve**  
   In the reserve handler (Redis strategy): after a successful decrement in Redis and creating a row in `inventory_transactions`, update stock in PG. On any error during the PG write, call rollback in Redis (`increment`). Ensure rollback is invoked at most once per failed operation (idempotent compensation).

2. **Logging and monitoring**  
   Log every compensation call (sku, quantity, requestId, PG error reason). Metrics: compensations per time unit, errors when calling `increment`. Alert on a rise in compensations or on `increment` failures.

3. **Reconciliation**  
   Scheduled background worker: for each product (or list of SKUs with active stock in Redis) read `stock_quantity` from PG and the value from Redis. If Redis is less than PG — set Redis to the PG value. If greater — by policy: either align to PG (source of truth) or log the discrepancy. Log reconciliation results.

4. **Handling `increment` failure**  
   If Redis is unavailable at compensation time: log the failed compensation, write to a "pending_rollbacks" table or queue. A separate process periodically retries `increment` for these entries. Reconciliation will eventually fix Redis stock even if some rollbacks never succeed.

5. **Tests**  
   Scenario "Redis decremented successfully, PG write fails" — assert that rollback is called and Redis stock is restored. Scenario "rollback in catch also fails" — assert that state is logged and/or enqueued for deferred compensation; on the next run of reconciliation or the rollback worker, consistency is restored.

6. **Documentation and runbook**  
   Describe: when compensation runs, how to read logs and metrics, what to do on mass discrepancies. Short runbook for prolonged Redis unavailability (temporary switch to pessimistic/optimistic, full reconciliation after Redis is back).

---

## Production strategy summary

| Step | Path | Action |
|------|------|--------|
| 1 | Happy path | Decrement in Redis (fast) → persist transaction and update stock in DB (reliable). Redis and PG match. |
| 2 | DB error | Catch failure → call `increment(sku, quantity)` in Redis (rollback). |
| 3 | Critical failure | If everything fails (e.g. `increment` in catch fails) → scheduled reconciliation restores consistency (PG is source of truth). |

---

## Interview deep dive: "What if `increment` in the catch block also fails?"

**Example question:** "What if Redis is down or the connection is lost when we try to roll back in `catch`?"

**Senior-level answer:**

> This is a classic distributed-systems problem. In that case we rely on **reconciliation**. PostgreSQL is our source of truth. A background worker runs on a schedule and compares stock in the DB with stock in Redis. If Redis has less than the DB (e.g. we never managed to run `increment`), the worker corrects Redis. That way we achieve **eventual consistency**: after the next reconciliation run, Redis and PG are aligned again.

---

## Data flow checklist (reserve via Redis)

1. **Reserve request** → idempotency check (e.g. by `requestId`).
2. **Redis:** `decrementIfSufficientOrInit(sku, pgStock, quantity)` → new balance or `null` (insufficient).
3. If `null` → return "insufficient stock".
4. **PG (single DB transaction):** create row in `inventory_transactions`, update `products.stock_quantity` to `newBalance` (sync). Both must succeed or both roll back.
5. If step 4 throws → **catch:** call `increment(sku, quantity)` to roll back Redis; then rethrow.
6. **Reconciliation (scheduled):** PG vs Redis; if Redis &lt; PG → set Redis to PG value.

---

## Sync and desync (списания и расхождение)

- **Списание (decrement)** происходит только в Redis (атомарно). Затем результат **синкается в PG**: запись в `inventory_transactions` и обновление `products.stock_quantity` до `newBalance`.
- **Потенциальное расхождение:** если после успешного списания в Redis запись в PG не удалась (или выполнилась только частично — транзакция есть, а `stock_quantity` не обновлён), Redis будет меньше, чем PG — десинхрон. Поэтому:
  1. Оба шага PG (create transaction + update stock) выполняются в **одной транзакции БД** — либо оба успешны, либо оба откатываются.
  2. При любой ошибке на этапе PG вызывается **компенсация**: `increment(sku, quantity)` в Redis, чтобы вернуть списанное количество.
- Так мы избегаем ситуации «списали в Redis, в PG не записали и не откатили».
