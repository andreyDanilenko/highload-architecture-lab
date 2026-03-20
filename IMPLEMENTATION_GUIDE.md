# Руководство по отработке задач: уровни сложности и стратегии реализации

> Документ для Senior/Lead инженеров: план реализации, альтернативные решения, оптимизации и чеклисты для каждой из 30 задач.

---

## Легенда уровней

| Уровень | Описание | Критичность |
|---------|----------|-------------|
| **L1** | Базовая реализация, MVP | Низкая |
| **L2** | Production-ready с edge cases | Средняя |
| **L3** | High-load, оптимизации, мониторинг | Высокая |
| **L4** | Enterprise: отказоустойчивость, multi-region | Критическая |

---

## Block I: Concurrency & Consistency

### 01 — Atomic Inventory Counter

**Уровни проблемы:**
- **L1:** Race condition при вычитании, отрицательный остаток
- **L2:** Deadlock при высокой конкуренции, throughput degradation
- **L3:** 100k RPS, latency spikes, connection pool exhaustion
- **L4:** Distributed inventory, eventual consistency across DC

**Стратегии решения:**

| Стратегия | Плюсы | Минусы | Когда использовать |
|-----------|-------|--------|---------------------|
| Pessimistic (SELECT FOR UPDATE) | Гарантия ACID, простота | Блокировки, deadlocks | Низкая конкуренция, критичные данные |
| Optimistic (version column) | Меньше блокировок, лучше throughput | Retry storms при конфликтах | Средняя конкуренция |
| Redis INCR/DECR | O(1), минимальная latency | Доп. система, eventual sync | Высокий RPS, flash sales |
| PostgreSQL advisory locks | Гибкость, именованные блокировки | Сложнее отладка | Специфичные сценарии |

**План реализации:**
1. Реализовать все 4 стратегии с бенчмарками
2. Добавить EXPLAIN ANALYZE для медленных запросов
3. Настроить connection pooling (PgBouncer)
4. Метрики: conflict rate, p99 latency, throughput
5. Load test: k6 с ramp-up до 100k concurrent

**Оптимизации:**
- Batch decrements (резервировать N единиц за один запрос)
- Pre-warming connection pool
- Индекс на `(product_id, version)` для optimistic
- Мониторинг `pg_stat_activity` на long-running transactions

---

### 02 — Anti-Bruteforce Vault

**Уровни проблемы:**
- **L1:** Неатомарный подсчёт попыток
- **L2:** Sliding vs fixed window — граничные случаи
- **L3:** Redis latency под нагрузкой, Lua script timeout
- **L4:** Multi-region, clock drift между инстансами

**Стратегии решения:**

| Стратегия | Реализация | Trade-off |
|-----------|------------|-----------|
| Fixed window | INCR + EXPIRE | Burst на границе окна |
| Sliding window log | ZADD + ZREMRANGEBYSCORE | Память O(n) по попыткам |
| Sliding window counter | Lua: approximate count | Неточность vs память |
| Token bucket | Lua + last_refill | Сложнее, гибче |

**План реализации:**
1. Lua script: атомарный ZADD + ZREMRANGEBYSCORE + ZCARD
2. Progressive delay: exponential backoff в Lua или приложении
3. Rate limit на сам vault endpoint (защита от DDoS)
4. Метрики: failed_attempts, lockout_duration, script_execution_time

**Оптимизации:**
- Pipeline для batch операций
- EVALSHA вместо EVAL (кэш скрипта)
- Redis Cluster: один ключ = один slot (hash tag)

---

### 03 — Heavy Task Worker Pool

**Уровни проблемы:**
- **L1:** Unbounded goroutines/workers — OOM
- **L2:** Graceful shutdown, in-flight task completion
- **L3:** Priority queue, backpressure propagation
- **L4:** Distributed workers, job persistence

**Стратегии решения:**

| Подход | Go | Node.js | Особенности |
|-------|-----|---------|-------------|
| Worker pool | N goroutines + channel | worker_threads / p-queue | Предсказуемое потребление |
| Semaphore | errgroup + semaphore | p-limit | Ограничение параллелизма |
| Priority queue | heap + channel | bull/bullmq | Критичные задачи первыми |

**План реализации:**
1. Semaphore на max concurrent tasks
2. Context с deadline для graceful shutdown
3. Task timeout + cancellation
4. Метрики: queue_depth, processing_time, rejection_rate

**Оптимизации:**
- Prefetch: worker забирает следующую задачу до завершения текущей
- Adaptive pool size по CPU/memory
- Dead letter queue для failed tasks

---

### 04 — Idempotency Key Provider

**Уровни проблемы:**
- **L1:** Duplicate request → duplicate side effect
- **L2:** Key collision, TTL vs business TTL mismatch
- **L3:** Redis failover во время check-and-set
- **L4:** Cross-service idempotency, key namespace

**Стратегии решения:**

| Подход | Хранение | Гарантии |
|--------|----------|----------|
| Redis SET NX | In-memory | Fast, но при потере — дубликаты |
| PostgreSQL UNIQUE | Durable | Медленнее, ACID |
| Hybrid | Redis + async to PG | Best of both |

**План реализации:**
1. Middleware: извлечь X-Idempotency-Key из header
2. Redis: SET key NX EX TTL с payload (request hash)
3. При NX=0: вернуть cached response (если сохранён)
4. Response caching: сохранять результат успешной операции

**Оптимизации:**
- Key format: `{tenant}:{operation}:{hash}` — изоляция
- TTL = max(operation_timeout * 2, 24h)
- Lua script для atomic check-and-set-and-store

---

## Block II: Distributed Systems & Networking

### 05 — Distributed Rate Limiter

**Уровни проблемы:**
- **L1:** Per-instance limit бесполезен за LB
- **L2:** Clock skew между инстансами
- **L3:** Redis single point of failure
- **L4:** Multi-tenant, разные лимиты по ключу

**Стратегии решения:**

| Алгоритм | Точность | Память | Redis round-trips |
|----------|----------|--------|-------------------|
| Fixed window | Низкая (burst) | O(1) | 2 |
| Sliding window log | Высокая | O(n) | 1 (Lua) |
| Token bucket | Средняя | O(1) | 2 |
| Leaky bucket | Высокая | O(1) | 2 |

**План реализации:**
1. Lua: sliding window с ZADD/ZREMRANGEBYSCORE/ZCARD
2. Fallback: при Redis unavailable — allow или deny? (configurable)
3. Rate limit key: IP, user_id, API_key — комбинируемые
4. Headers: X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After

**Оптимизации:**
- Local cache лимита (allow 10% burst без Redis)
- Redis Cluster: hash tag для ключа пользователя
- Метрика: rate_limit_hits по dimension

---

### 06 — Multilayer Cache

**Уровни проблемы:**
- **L1:** Cache stampede при массовом expire
- **L2:** L1/L2 inconsistency, stale reads
- **L3:** Memory pressure, eviction policy
- **L4:** Cache invalidation across instances

**Стратегии решения:**

| Паттерн | Описание | Consistency |
|---------|----------|-------------|
| Cache-aside | App управляет | Best effort |
| Read-through | Cache прозрачно | TTL-based |
| Write-through | Write в cache+DB | Strong |
| Refresh-ahead | Probabilistic early expire | Eventual |

**План реализации:**
1. L1: sync.Map (Go) / LRU (Node) с TTL
2. L2: Redis с тем же key schema
3. Single flight: один запрос на miss — остальные ждут
4. Probabilistic early expiration: `expire * (0.8 + 0.2*rand)` 

**Оптимизации:**
- Медленные запросы: кэшировать дольше, отдельный TTL
- Cache warming при старте
- Метрики: hit_ratio L1/L2, stampede_prevented, load_on_db

---

### 07 — Secure BFF

**Уровни проблемы:**
- **L1:** JWT validation, signature verification
- **L2:** Token refresh, session fixation
- **L3:** Aggregation latency, partial failure
- **L4:** Zero-trust, mTLS, key rotation

**Стратегии решения:**

| Аспект | Решения |
|--------|---------|
| Auth | JWT (RS256), Opaque token + introspection |
| Session | HttpOnly cookie, short-lived access token |
| Aggregation | Parallel fetch, timeout per service |

**План реализации:**
1. JWT middleware: verify signature, exp, iss, aud
2. JWKS endpoint для key rotation
3. BFF aggregates: Promise.allSettled / errgroup
4. Secure headers: CSP, HSTS, X-Frame-Options

**Оптимизации:**
- Кэш JWKS (TTL 1h)
- Connection pooling к backend services
- Circuit breaker на каждый backend

---

### 08 — API Gateway Aggregator

**Уровни проблемы:**
- **L1:** Sequential calls — N * latency
- **L2:** One slow service blocks response
- **L3:** Partial failure handling, degradation
- **L4:** Cross-region aggregation, timeout budget

**Стратегии решения:**

| Паттерн | Описание |
|---------|----------|
| Scatter-gather | Parallel fetch, aggregate |
| Fan-out | Fire-and-forget для non-critical |
| Fallback | Default/cached при failure |

**План реализации:**
1. Deadline propagation: `context.WithTimeout` от incoming request
2. Per-service timeout < total deadline
3. Partial response: возвращать что есть + errors array
4. Dependency graph: если B зависит от A — sequential для них

**Оптимизации:**
- Медленные запросы: вынести в отдельный endpoint, async
- Response caching для идемпотентных агрегаций
- Request coalescing для идентичных параллельных запросов

---

## Block III: Data Engineering Under Load

### 09 — Terabyte Data Mocker

**Уровни проблемы:**
- **L1:** Single INSERT — тысячи запросов/сек
- **L2:** Index creation блокирует INSERT
- **L3:** WAL, checkpoint, disk I/O bottleneck
- **L4:** Distributed generation, partitioning

**Стратегии решения:**

| Метод | Throughput | Когда |
|-------|------------|-------|
| Batch INSERT (100-1000 rows) | 10-50k rows/s | Универсально |
| COPY FROM | 100-500k rows/s | Bulk load |
| UNLOGGED table | 2x faster | Temp data, можно потерять |
| Partitioning | Parallel load | Большие таблицы |

**План реализации:**
1. COPY с stdin stream
2. Disable indexes → load → create indexes
3. ANALYZE после load
4. Параллельные workers по partition

**Оптимизации медленных запросов:**
- `EXPLAIN (ANALYZE, BUFFERS)` для каждого типа запроса
- `pg_stat_statements` — топ медленных
- `work_mem`, `maintenance_work_mem` для bulk
- `fillfactor` для append-only таблиц

---

### 10 — Read/Write Splitter

**Уровни проблемы:**
- **L1:** Read your writes — replica lag
- **L2:** Replication lag мониторинг
- **L3:** Sticky session для consistency
- **L4:** Multi-region read replicas

**Стратегии решения:**

| Стратегия | Consistency | Complexity |
|-----------|-------------|------------|
| Random replica | Eventual | Low |
| Lag-aware routing | Better | Medium |
| Session stickiness | Read-your-writes | Medium |
| Sync replica wait | Strong | High latency |

**План реализации:**
1. Parse SQL: SELECT → replica, INSERT/UPDATE/DELETE → master
2. Transaction: весь transaction на master
3. `pg_stat_replication` для lag monitoring
4. При lag > threshold: route to master

**Оптимизации:**
- Connection pool per replica
- Метрика: replica_lag_seconds
- Fallback: replica down → master для reads

---

### 11 — Custom Database Sharder

**Уровни проблемы:**
- **L1:** Hotspot на один shard
- **L2:** Resharding — data migration
- **L3:** Cross-shard queries
- **L4:** Rebalancing без downtime

**Стратегии решения:**

| Стратегия | Равномерность | Reshard cost |
|-----------|---------------|--------------|
| Range | Poor (hotspot) | Low |
| Hash | Good | High (rehash all) |
| Consistent hashing | Good | O(1/n) перемещений |
| Virtual nodes | Better | Same |

**План реализации:**
1. Consistent hashing ring с 100-200 virtual nodes per shard
2. Shard key в каждом запросе
3. Connection pool per shard
4. Dual-write при resharding (transition period)

**Оптимизации:**
- Мониторинг: requests per shard, data distribution
- Slow query log per shard
- Prepared statements per shard connection

---

### 12 — SaaS Multitenancy Isolation

**Уровни проблемы:**
- **L1:** Tenant data leak (SQL injection, bug)
- **L2:** Noisy neighbor — один tenant грузит всех
- **L3:** Connection pool exhaustion
- **L4:** Compliance (GDPR, isolation levels)

**Стратегии решения:**

| Модель | Изоляция | Эффективность |
|--------|----------|---------------|
| Schema per tenant | High | Medium |
| Row-Level Security | Medium | High |
| Database per tenant | Highest | Low |

**План реализации:**
1. RLS policies: `tenant_id = current_setting('app.tenant_id')`
2. SET app.tenant_id в connection
3. Connection pool: отдельный или shared с validation
4. Quota per tenant: rate limit, connection limit

**Оптимизации:**
- Индекс на (tenant_id, ...) для всех таблиц
- `pg_stat_user_tables` по tenant для выявления тяжёлых
- Partitioning по tenant_id для больших tenants

---

## Block IV: Complex Patterns & Real-time

### 13 — High-Load Chat Engine

**Уровни проблемы:**
- **L1:** 10k connections — file descriptor limit
- **L2:** Message ordering, delivery guarantee
- **L3:** 50k+ connections, broadcast storm
- **L4:** Multi-region, message persistence

**Стратегии решения:**

| Аспект | Решения |
|--------|---------|
| Scaling | Redis Pub/Sub, sticky sessions |
| Persistence | Write to DB async, read from cache |
| Ordering | Sequence numbers, idempotent delivery |

**План реализации:**
1. WebSocket server с connection registry
2. Redis PUBLISH на broadcast
3. SUBSCRIBE в каждом instance
4. Connection limits: ulimit, graceful reject

**Оптимизации:**
- Медленные клиенты: message queue per connection
- Batch messages при высокой частоте
- Метрики: connections_per_instance, messages_per_sec

---

### 14 — Real-time Leaderboard

**Уровни проблемы:**
- **L1:** ZADD при каждом score update
- **L2:** Pagination, ranking consistency
- **L3:** Millions of members, memory
- **L4:** Historical leaderboards, time windows

**Стратегии решения:**

| Операция | Redis команда | Complexity |
|----------|---------------|------------|
| Update score | ZADD | O(log N) |
| Get rank | ZREVRANK | O(log N) |
| Get top K | ZREVRANGE | O(log N + K) |
| Range by rank | ZREVRANGE | O(log N + K) |

**План реализации:**
1. Key: `leaderboard:{game_id}:{season}`
2. Score = points (или composite: points + timestamp)
3. Pagination: ZREVRANGE start stop
4. TTL для сезонных leaderboards

**Оптимизации:**
- Pipeline для batch updates
- Отдельный sorted set для top 1000 (materialized)
- Redis memory: ziplist для маленьких sets

---

### 15 — Distributed SAGA Orchestrator

**Уровни проблемы:**
- **L1:** Partial failure — inconsistent state
- **L2:** Compensation logic, idempotency
- **L3:** Long-running saga, timeout
- **L4:** Cross-service, cross-region

**Стратегии решения:**

| Тип | Координация | Когда |
|-----|-------------|-------|
| Choreography | Events | Простые flows |
| Orchestration | Central coordinator | Сложные, нужен контроль |

**План реализации:**
1. Saga state machine в DB или Redis
2. Each step: execute + record
3. On failure: execute compensations в reverse order
4. Idempotency key per step

**Оптимизации:**
- Timeout per step
- Retry с exponential backoff
- Saga log для debugging

---

### 16 — Circuit Breaker Service

**Уровни проблемы:**
- **L1:** Cascading failure
- **L2:** Threshold tuning
- **L3:** Half-open storm
- **L4:** Distributed state

**Стратегии решения:**

| State | Действие |
|-------|----------|
| Closed | Normal operation |
| Open | Fail fast, no calls |
| Half-open | Probing, limited calls |

**План реализации:**
1. Sliding window: failures / total (или time window)
2. Threshold: 50% failure rate или 5 consecutive
3. Half-open: 1-3 probe requests
4. Bulkhead: отдельный pool для каждого dependency

**Оптимизации:**
- Метрики: circuit_state, failure_rate
- Per-instance vs distributed (Redis) state
- Fallback response при open

---

## Block V: Observability

### 17 — Dynamic Feature Toggle

**Уровни проблемы:**
- **L1:** Config в коде — нужен redeploy
- **L2:** Stale config при cache
- **L3:** High availability config store
- **L4:** A/B testing, gradual rollout по сегментам

**Стратегии решения:**

| Backend | Latency | Consistency | Когда |
|---------|---------|-------------|-------|
| Redis | Low | Eventual | Большинство случаев |
| ETCD | Low | Strong | K8s ecosystem |
| DB | Medium | Strong | Уже есть, простой вариант |

**План реализации:**
1. Redis/ETCD как config store, key: `feature:{name}`
2. Watch/poll для real-time updates
3. Local cache TTL 10-60s, invalidate on update
4. Fallback: default при unavailability
5. Segment support: user_id hash % 100 для gradual rollout

**Оптимизации:**
- Batch fetch всех flags за один запрос
- Метрики: config_fetch_latency, cache_hit_ratio

---

### 18 — Log Aggregator (Mini ELK)

**Уровни проблемы:**
- **L1:** Логи в файлах на каждой ноде
- **L2:** Correlation ID сквозь сервисы
- **L3:** High volume, disk I/O
- **L4:** Retention, compliance, search performance

**Стратегии решения:**

| Компонент | Варианты |
|-----------|----------|
| Format | JSON structured, logfmt |
| Shipper | Filebeat, Fluentd, Vector |
| Storage | Elasticsearch, ClickHouse, Loki |
| Search | Kibana, Grafana Explore |

**План реализации:**
1. Structured JSON: `{timestamp, level, msg, correlation_id, service, ...}`
2. Correlation ID: X-Request-ID в HTTP, propagate
3. Log shipper: tail files → batch send
4. Storage: индекс по timestamp, correlation_id, level, service

**Оптимизации медленных запросов:**
- Индексы на correlation_id, service, level, timestamp
- Sampling: 100% error, 10% info при нагрузке
- Retention policy: hot/warm/cold tiers
- ClickHouse: partitioning по дате, ORDER BY (timestamp, service)

---

### 19 — System Metrics Exporter

**Уровни проблемы:**
- **L1:** Нет метрик — слепая зона
- **L2:** Cardinality explosion (user_id в labels)
- **L3:** Scrape timeout при многих targets
- **L4:** Long-term storage, aggregation

**Стратегии решения:**

| Metric type | Использование |
|-------------|---------------|
| Counter | RPS, errors, total requests |
| Gauge | Queue depth, connections, memory |
| Histogram | Latency p50/p95/p99 |

**План реализации:**
1. Instrument: middleware для HTTP, custom для business
2. Labels: endpoint, method, status (ограниченный набор)
3. Histogram buckets: 5ms, 25ms, 100ms, 500ms, 1s, 2.5s, 5s
4. /metrics endpoint, Prometheus scrape

**Оптимизации:**
- Не лейблить user_id, request_id — cardinality
- Метрики медленных запросов: `http_request_duration_seconds` с endpoint
- Recording rules для тяжёлых агрегаций

---

### 20 — The Grand Dashboard

**Уровни проблемы:**
- **L1:** Метрики есть, визуализации нет
- **L2:** Связь метрик между сервисами
- **L3:** Load test + dashboard correlation
- **L4:** SLO-based alerting, runbooks

**План реализации:**
1. k6: ramp-up, steady, ramp-down сценарии
2. Grafana: panels для каждого сервиса (RPS, latency, errors)
3. Comparison: baseline vs current run
4. Alerts: latency p99 > threshold, error rate > 1%

**Оптимизации:**
- Dashboard variables для выбора сервиса/времени
- Annotations: mark load test start/end
- Link to traces/logs from metric spike

---

## Block VI: Message Brokers & Event-Driven

### 21 — Kafka Exactly-Once Delivery

**Уровни проблемы:**
- **L1:** At-least-once — дубликаты при retry
- **L2:** Producer idempotence
- **L3:** Transactional read-process-write
- **L4:** Cross-partition exactly-once

**Стратегии решения:**

| Уровень | Механизм |
|---------|----------|
| Producer | enable.idempotence=true, acks=all |
| Consumer | Commit после обработки, idempotent handler |
| Transactional | Producer init transactions, consumer read_committed |

**План реализации:**
1. Idempotent producer: PID + sequence
2. Consumer: idempotency key в бизнес-логике (как в задаче 04)
3. Transactional: для exactly-once stream processing
4. Мониторинг: consumer lag, duplicate_detected metric

**Оптимизации:**
- Batch processing с сохранением order
- Consumer group tuning: max.poll.records, session.timeout

---

### 22 — Event Sourcing Engine

**Уровни проблемы:**
- **L1:** Append-only event store
- **L2:** Snapshot для быстрого restore
- **L3:** Replay performance на больших логах
- **L4:** Schema evolution, event versioning

**Стратегии решения:**

| Аспект | Решения |
|--------|---------|
| Store | Kafka, PostgreSQL (JSONB), dedicated (EventStoreDB) |
| Snapshot | Every N events или по времени |
| Replay | Parallel by aggregate ID |

**План реализации:**
1. Event schema: type, aggregate_id, version, payload, timestamp
2. Append: optimistic concurrency (version check)
3. Snapshot: materialized state каждые 100 events
4. Replay: load snapshot + apply events since

**Оптимизации:**
- Индекс на (aggregate_id, version)
- Partitioning по aggregate_id для параллельного replay
- Event compaction (snapshot + tail)

---

### 23 — Distributed Job Scheduler

**Уровни проблемы:**
- **L1:** Cron на каждой ноде — N копий job
- **L2:** Distributed lock — один исполнитель
- **L3:** Leader election, failover
- **L4:** Long-running jobs, checkpointing

**Стратегии решения:**

| Подход | Координация | Когда |
|--------|-------------|-------|
| Lock | Redis SET NX | Простые cron jobs |
| Leader election | etcd, ZooKeeper | Нужен один leader |
| Queue | Kafka, Redis Queue | Job как message |

**План реализации:**
1. Redis lock: key=job_name, value=instance_id, TTL=job_timeout*2
2. Lock renewal: background goroutine, extend TTL каждые TTL/3
3. Fencing token: increment при каждом acquire
4. Job metadata: next_run, last_run, status

**Оптимизации:**
- Мониторинг: job_duration, lock_contention
- Stagger: не все jobs в 00:00
- Dead job detection: lock expired без completion

---

### 24 — Change Data Capture (CDC)

**Уровни проблемы:**
- **L1:** Debezium + Kafka Connect setup
- **L2:** Schema mapping, transformation
- **L3:** Backpressure при slow consumer
- **L4:** Schema evolution, new columns

**Стратегии решения:**

| Компонент | Варианты |
|-----------|----------|
| Capture | Debezium, pg_logical |
| Transport | Kafka |
| Transform | Kafka Connect SMT, Flink |

**План реализации:**
1. PostgreSQL: logical replication, publication
2. Debezium connector: table include list
3. Kafka topic: schema registry для Avro
4. Consumer: idempotent apply to target

**Оптимизации:**
- Мониторинг: connector lag, consumer lag
- Медленные consumers: увеличить partitions, parallel consumers
- Snapshot mode: initial vs incremental

---

## Block VII: High Performance & Networking

### 25 — Custom TCP/UDP Proxy

**Уровни проблемы:**
- **L1:** Basic forward
- **L2:** Health checks, failover
- **L3:** Connection pooling, keepalive
- **L4:** Zero-copy forwarding

**Стратегии решения:**

| Алгоритм | Описание |
|----------|----------|
| Round-robin | По очереди |
| Least connections | Меньше всего активных |
| IP hash | Sticky по client IP |

**План реализации:**
1. Listen → accept → connect to backend
2. io.Copy bidirectional (или splice на Linux)
3. Health check: TCP probe или HTTP
4. Backend selection по алгоритму

**Оптимизации:**
- Connection pooling к backends
- SO_REUSEPORT для multi-thread accept
- Метрики: connections_per_backend, latency

---

### 26 — Zero-Copy File Server

**Уровни проблемы:**
- **L1:** read() + write() — 4 copy operations
- **L2:** sendfile() — kernel copy
- **L3:** mmap + write — 2 copies
- **L4:** DMA, kernel bypass (DPDK) — advanced

**Стратегии решения:**

| Метод | Copies | Когда |
|-------|--------|-------|
| read/write | 4 | Baseline |
| sendfile | 2 | Static files, Linux |
| mmap | 2 | Random access |

**План реализации:**
1. sendfile(out_fd, in_fd, offset, count)
2. Content-Length, Range support
3. Benchmark: wrk/ab, compare throughput
4. Go: io.Copy with ReaderFrom optimization

**Оптимизации:**
- Preload small files in memory
- Async I/O для больших файлов
- Метрики: bytes_served, cache_hit

---

### 27 — Binary Protocol Parser

**Уровни проблемы:**
- **L1:** JSON — просто, но медленно
- **L2:** Schema evolution
- **L3:** Cross-language compatibility
- **L4:** Streaming, partial parse

**Стратегии решения:**

| Формат | Size | Speed | Schema |
|--------|------|-------|--------|
| JSON | Large | Slow | No |
| MessagePack | Medium | Fast | No |
| Protobuf | Small | Fast | Yes |
| FlatBuffers | Small | Zero-copy | Yes |

**План реализации:**
1. Protobuf schema definition
2. Code generation для Go/Node
3. Benchmark: serialize/deserialize, payload size
4. Backward compatibility: optional fields, reserved

**Оптимизации:**
- Метрики: serialization_duration
- Pool для message objects (reduce GC)
- Schema registry для versioning

---

## Block VIII: Security & Decentralization

### 28 — Distributed Lock Manager (Redlock)

**Уровни проблемы:**
- **L1:** Single Redis — SPOF
- **L2:** Clock drift, false release
- **L3:** Fencing tokens при split-brain
- **L4:** Multi-datacenter

**Стратегии решения:**

| Аспект | Решение |
|--------|---------|
| Quorum | N/2+1 Redis instances |
| Value | Random, verify on release |
| TTL | Больше max execution time |
| Fencing | Monotonic token в protected resource |

**План реализации:**
1. 5 Redis instances (или 3 minimum)
2. Lock: SET NX PX на всех, quorum = 3
3. Unlock: verify value, DEL
4. Renewal: extend TTL до завершения
5. Fencing: при записи в DB — check token > last

**Оптимизации:**
- Jitter при retry
- Метрики: lock_acquisition_time, contention_count
- Fallback: при недоступности — fail safe (не брать lock)

---

### 29 — Merkle Tree Validator

**Уровни проблемы:**
- **L1:** Verify single record
- **L2:** Verify subset — Merkle proof
- **L3:** Sync two trees — diff by subtree
- **L4:** Incremental update, persistence

**Стратегии решения:**

| Операция | Complexity |
|----------|------------|
| Build | O(n) |
| Proof | O(log n) |
| Verify | O(log n) |
| Diff | O(k log n) where k = differences |

**План реализации:**
1. Leaf = hash(record)
2. Parent = hash(left + right)
3. Proof = sibling path to root
4. Verify: recompute root from leaf + proof

**Оптимизации:**
- Batch verification
- Tree persistence для incremental
- Parallel build по уровням

---

### 30 — Hot/Cold Wallet Logic

**Уровни проблемы:**
- **L1:** Single key — compromise = total loss
- **L2:** Withdrawal limits, approval
- **L3:** Multi-sig, threshold
- **L4:** Audit, compliance, key rotation

**Стратегии решения:**

| Аспект | Решение |
|--------|---------|
| Hot | Small balance, fast access |
| Cold | Bulk, offline, multi-sig |
| Withdrawal | Queue, approval workflow |
| Audit | Immutable log всех операций |

**План реализации:**
1. Hot wallet: API, daily limit
2. Cold: offline, M-of-N signatures
3. Withdrawal: create → approve (N) → execute
4. Audit log: append-only, hash chain

**Оптимизации:**
- Rate limit на withdrawal requests
- Alert на аномальные суммы
- Key ceremony documentation

---

## Общий чеклист Senior/Lead

Для каждой задачи:

- [ ] **Медленные запросы:** EXPLAIN ANALYZE, pg_stat_statements, индексы
- [ ] **Метрики:** latency p50/p95/p99, error rate, throughput
- [ ] **Failure modes:** что если Redis/DB/network down?
- [ ] **Load test:** k6 с realistic scenario
- [ ] **Documentation:** ADR для выбранных решений
- [ ] **Observability:** logs, traces, dashboards

---

*Документ создан для систематической отработки 30 Highload Engineering Challenges.*
