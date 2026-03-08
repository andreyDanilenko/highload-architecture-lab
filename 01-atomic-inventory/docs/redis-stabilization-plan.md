<<<<<<< HEAD
# Redis: компенсирующая транзакция и план стабилизации

Описание задачи, риски и план доработки стратегии 4 (Redis) без привязки к коду.

---

## Задача (описание)

Стратегия 4 (Redis) даёт максимальную скорость резерва за счёт атомарного списания в Redis, после чего результат синхронизируется в PostgreSQL. Чтобы система оставалась консистентной при сбоях, нужен **паттерн компенсирующей транзакции**:

1. **Счастливый путь:** списали в Redis (атомарно) → сохранили транзакцию и обновили остаток в БД. Redis и PG совпадают.
2. **Ошибка при записи в БД:** после успешного списания в Redis запись в PostgreSQL падает (таймаут, диск, deadlock). Без отката в Redis остаток в кеше будет меньше, чем в БД — рассинхрон. Поэтому в `catch` необходимо вернуть количество обратно в Redis (метод `increment(sku, quantity)`). Так откатывается только та операция, которая не зафиксировалась в БД.
3. **Критический сбой:** если упали и Redis, и приложение (например, `increment` в `catch` не выполнился из‑за потери связи с Redis), консистентность восстанавливается **фоновой сверкой** (reconciliation): PostgreSQL — источник правды; воркер периодически сравнивает остатки в БД и Redis и при необходимости правит Redis. Итог — eventual consistency.

Метод `increment` в хранилище Redis реализует откат (Redis `INCRBY`). Без него компенсирующая транзакция невозможна.

---

## Потенциальные проблемы (Redis-стратегия)

- **Падение `increment` в блоке `catch`** (нет связи с Redis, таймаут): компенсация не выполнится, Redis останется заниженным. Решение — reconciliation: фоновый процесс сверяет PG и Redis и выравнивает остатки; источник правды — БД.
- **Двойной откат:** если по ошибке вызвать `increment` дважды для одной и той же неудавшейся операции, остаток в Redis станет завышенным. Нужна идемпотентность компенсации (например, запись в таблицу откатов по `requestId` или флаг в транзакции).
- **Порядок операций:** компенсация должна вызываться только когда списание в Redis уже произошло, а запись в PG — нет. Иначе риск откатить то, что в БД не записывали.
- **Синхронизация PG после Redis:** текущая реализация пишет в PG после каждого резерва. При очень высокой нагрузке можно рассмотреть батчирование или асинхронную запись в очередь с последующей записью в PG — с пониманием, что до момента записи PG будет отставать (читаем из Redis для резерва, из PG — для отчётов после сверки).

---

## План стабилизации (реализации)

Только перечень шагов, без кода.

1. **Компенсирующая транзакция в резерве**  
   В обработчике резерва (стратегия Redis): после успешного списания в Redis и создания записи в `inventory_transactions` выполнить обновление остатка в PG. При любой ошибке на этапе записи в PG вызвать откат в Redis (`increment`). Убедиться, что откат вызывается не более одного раза на одну неудавшуюся операцию (идемпотентность компенсации).

2. **Логирование и мониторинг**  
   Логировать каждый вызов компенсации (sku, quantity, requestId, причина ошибки PG). Метрики: количество компенсаций в единицу времени, ошибки при вызове `increment`. Алерты при росте числа компенсаций или при сбоях `increment`.

3. **Reconciliation (сверка)**  
   Фоновый воркер по расписанию: для каждого продукта (или по списку SKU с активным остатком в Redis) прочитать `stock_quantity` из PG и значение из Redis. Если значение в Redis меньше, чем в PG — выставить в Redis значение из PG. Если больше — по политике: либо выровнять по PG (источник правды), либо залогировать расхождение. Результаты сверки логировать.

4. **Обработка падения `increment`**  
   Если в момент компенсации Redis недоступен: залогировать неудачную компенсацию, записать в таблицу «pending_rollbacks» или в очередь. Отдельный процесс периодически повторяет попытки `increment` по этим записям. Reconciliation со временем исправит остаток в Redis даже если часть откатов не выполнится.

5. **Тесты**  
   Сценарий «Redis списал успешно, запись в PG падает» — проверка, что вызывается откат и остаток в Redis восстанавливается. Сценарий «откат в catch тоже падает» — проверка, что состояние логируется и/или попадает в очередь отложенной компенсации; при следующем запуске reconciliation или воркера откатов консистентность восстанавливается.

6. **Документация и runbook**  
   Описать: когда срабатывает компенсация, как читать логи и метрики, что делать при массовых расхождениях. Краткий runbook на случай длительной недоступности Redis (временный переход на pessimistic/optimistic, полная сверка после восстановления Redis).
=======
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
   If Redis is unavailable at compensation time: log the failed compensation, write to a “pending_rollbacks” table or queue. A separate process periodically retries `increment` for these entries. Reconciliation will eventually fix Redis stock even if some rollbacks never succeed.

5. **Tests**  
   Scenario “Redis decremented successfully, PG write fails” — assert that rollback is called and Redis stock is restored. Scenario “rollback in catch also fails” — assert that state is logged and/or enqueued for deferred compensation; on the next run of reconciliation or the rollback worker, consistency is restored.

6. **Documentation and runbook**  
   Describe: when compensation runs, how to read logs and metrics, what to do on mass discrepancies. Short runbook for prolonged Redis unavailability (temporary switch to pessimistic/optimistic, full reconciliation after Redis is back).
>>>>>>> ab9d472 (feat(01-atomic-inventory): Redis compensating transaction, DI refactor, stabilization doc)
