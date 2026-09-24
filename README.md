# 30 Highload Engineering Challenges

[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Node.js](https://img.shields.io/badge/Node.js-20+-339933?style=flat-square&logo=nodedotjs)](https://nodejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=flat-square&logo=redis)](https://redis.io/)
[![Kafka](https://img.shields.io/badge/Kafka-3.5-231F20?style=flat-square&logo=apache-kafka)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)

### Hi, my name is Andrey 
and I have been a frontend developer for 4 years. To be honest, I always thought that an IT position was really easy for most tasks. But once I got a good position at a good company, I relaxed and forgot that a good developer must always learn something new.

Today, when my company lost the project actually, I decided to continue my journey. This guide is for developers who share my ideas and want to become better. This guide will be helpful for other positions, not only backend, because it touches a lot of topics about building systems.

Sometimes people believe that AI is our enemy, but fundamental knowledge will always be relevant. This guide was created with help from AI, but I think that's the least benefit you can get from AI.

AI can help write the code, prepare experiments, and explain mechanisms. As a learner, I want to predict behavior, trace execution, debug failures, and explain what the evidence supports. Completed code and demonstrated understanding are recorded separately.

A roadmap of 30 hands-on engineering challenges, followed by optional systems and research tracks in [projects 31–60](README(31-60).md) and [projects 61–107](README(61-107).md). Each challenge explores a system behavior through implementations, failures, and measurements. The lab develops engineering thinking through experiments, failure analysis, and work within constraints. Experience operating real systems, taking responsibility for decisions, and collaborating with people complements this foundation.

---

**Initial implementations:** tasks 01–04, with different levels of completeness. [Task 04](04-idempotency/README.md) includes four implemented strategies and a [recorded experiment](04-idempotency/docs/experiment-notes.md); code availability and learner understanding are tracked separately.

---

## Explore the challenges

- [Articles, documentation and real projects for 01–30](docs/reading-map-01-30.md) — English and Russian sources with pointers to what to inspect.
- [Quality review of the first 30 challenge descriptions](docs/quality-review-01-30.md) — useful scope, overlaps and suggested smaller experiments.
- [EventLab and connected playgrounds for 01–107](docs/project-playground-01-107.md) — an optional way to grow an events-and-tickets app from task 05, or combine experiments into separate projects.

The main activity is exploring and changing working examples. These maps are browsing aids, not a requirement to complete every challenge or put every technology into one application. Sources and a detailed review currently cover 01–30 only.

---

## The Philosophy

These challenges provide practice in:

- Choosing between competing strategies (pessimistic vs optimistic locks, fixed vs sliding windows)
- Designing for failure (retries, circuit breakers, timeouts)
- Making conscious trade-offs (consistency vs availability, latency vs throughput)
- Seeing how components affect the whole system

The roadmap proposes Go and Node.js implementations to compare concurrency models. Availability differs by task; for example, task 01 currently has a Node.js implementation and a Go placeholder. Task READMEs distinguish current code from planned extensions.

For each experiment, record the hypothesis, invariant, workload, failure model, and observed limits. Track **planned**, **implemented**, and **experimentally verified** separately. A successful demonstration supports the tested conditions; operational readiness requires validation against the intended workload and failure model.

---

## Learning and contributing

Start with one question and a small experiment. Keep a dated notebook at each task's `docs/experiment-notes.md`: hypothesis, setup, actual results, guarantee boundaries, counterexamples, and your own reflection. AI drafts, measured facts, and learner reflections stay distinct; an empty reflection is better than invented understanding.

- [Contributor and agent workflow](CONTRIBUTING.md) — roles, decomposition, shared code, and reproducible experiments (Russian).
- [Experiment notebook template](docs/experiment-template.md) — copy into a task and add one entry per experiment (Russian).
- [Go learning map](docs/go-learning-map.md) — connect language mechanisms to experiments and evidence of understanding (Russian).

The same practice supports other learners: choose the depth and pace that answer your question. Structured notes can later support search or a RAG experiment; no retrieval system is assumed to exist yet.

---

## Planned Repository Structure

```
/
├── infrastructure/        # Shared dependencies (Postgres, Redis, Kafka, Prometheus, Grafana)
├── 01-atomic-inventory/  # ✅ Foundation: concurrency strategies
├── 02-anti-bruteforce/   # ✅ Rate limiting with Redis
├── 03-worker-pool/       # ✅ Worker pools and semaphores
├── 04-idempotency/       # Request deduplication
├── 05-rate-limiter/      # Cluster-wide Redis limiter
├── 06-multilayer-cache/  # L1 (memory) + L2 (Redis)
├── 07-bff/               # Backend for frontend with JWT
├── 08-gateway/           # Fan-out API aggregation
├── 09-data-mocker/       # Bulk insert and indexing
├── 10-readwrite-splitter/# Master + replica patterns
├── 11-sharder/           # Consistent hashing
├── 12-multitenancy/      # Row-Level Security
├── 13-chat-engine/       # WebSockets + Pub/Sub
├── 14-leaderboard/       # Redis Sorted Sets
├── 15-saga/              # Distributed transactions
├── 16-circuit-breaker/   # Failure protection
├── 17-feature-toggle/    # Runtime configuration
├── 18-log-aggregator/    # Structured logging
├── 19-metrics-exporter/  # Prometheus metrics
├── 20-grand-dashboard/   # Grafana visualization
├── 21-kafka-exactly-once/# Kafka transaction boundaries
├── 22-event-sourcing/    # Event store
├── 23-job-scheduler/     # Distributed locks
├── 24-cdc/               # Change Data Capture
├── 25-tcp-proxy/         # L4 load balancing
├── 26-zero-copy/         # sendfile, DMA
├── 27-binary-protocol/   # Protobuf/MessagePack
├── 28-redlock/           # Distributed locks
├── 29-merkle-tree/       # Hash trees
└── 30-wallet/            # Multi-signature
```

---

## Block I: Concurrency & Consistency (Single Instance)

*Foundation of any service. Working with threads, locks, and guarantees within a single instance.*

### 01 — Atomic Inventory Counter ✅
**What:** Compare stock-deduction strategies under increasing contention while checking that stock never goes negative; record achieved load and hardware limits.  
**Why:** Understand race conditions and locking strategies.  
**Implementation:** Compare pessimistic locks (`SELECT FOR UPDATE`), optimistic locks (version column), and Redis atomic operations.  
**What you'll learn:** Transaction isolation levels, deadlocks, and when to use each locking strategy.

**Explore:** [EN: theory/article](https://www.postgresql.org/docs/current/explicit-locking.html) · [RU: material](https://postgrespro.ru/docs/postgresql/current/explicit-locking) · [Project/code](https://github.com/postgres/postgres/tree/master/src/test/isolation) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-01).

### 02 — Anti-Bruteforce Vault ✅
**What:** PIN-protected vault with progressive delays on failed attempts.  
**Why:** Rate limiting is everywhere (login, password reset, 2FA).  
**Implementation:** Sliding window log in Redis with Lua scripts for atomicity.  
**What you'll learn:** Atomic Redis operations, sliding window vs fixed window, and writing Lua scripts.

**Explore:** [EN: theory/article](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html) · [RU: material](https://nginx.org/ru/docs/http/ngx_http_limit_req_module.html) · [Project/code](https://github.com/nginx/nginx/blob/master/src/http/modules/ngx_http_limit_req_module.c) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-02).

### 03 — Heavy Task Worker Pool
**What:** Process background jobs with controlled resource consumption.  
**Why:** Prevent system overload from unpredictable workloads.  
**Implementation:** Worker pool with semaphores, task queue, graceful shutdown.  
**What you'll learn:** Goroutines vs event loop, backpressure, and task prioritization.

**Explore:** [EN: theory/article](https://go.dev/blog/pipelines) · [RU: material](https://learn.microsoft.com/ru-ru/dotnet/core/extensions/queue-service) · [Project/code](https://github.com/golang/sync/blob/master/semaphore/semaphore_example_test.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-03).

### 04 — Idempotency Key Provider
**What:** Compare request-deduplication strategies and the conditions under which they prevent repeated business effects.  
**Why:** Payment retries shouldn't charge twice.  
**Implementation:** Store scoped request keys and fingerprints, replay completed results, and test failures between the business effect and result persistence. Redis atomicity alone does not cover an external effect.  
**What you'll learn:** Idempotency patterns, exactly-once semantics, and idempotency key lifecycle.

**Explore:** [EN: theory/article](https://docs.stripe.com/api/idempotent_requests) · [RU: material](https://yookassa.ru/developers/using-api/interaction-format) · [Project/code](https://github.com/stripe/stripe-go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-04).

---

## Block II: Distributed Systems & Networking

*Making multiple servers work as one organism.*

### 05 — Distributed Rate Limiter
**What:** Rate limit across a cluster, not per instance.  
**Why:** Per-instance limits protect each node; enforcing a shared quota requires coordination and an explicit policy for partitions.  
**Implementation:** Redis-based sliding window with Lua.  
**What you'll learn:** Distributed state management, clock synchronization issues, and atomic scripts.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/rate-limiting-pattern) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/rate-limiting-pattern) · [Project/code](https://github.com/go-redis/redis_rate) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-05).

### 06 — Multilayer Cache
**What:** L1 (in-memory) + L2 (Redis) with cache stampede protection.  
**Why:** Cache misses at high load can kill databases.  
**Implementation:** Single flight pattern, probabilistic early expiration.  
**What you'll learn:** Cache strategies (cache-aside, read-through), stampede prevention, and consistency trade-offs.

**Explore:** [EN: theory/article](https://blog.cloudflare.com/sometimes-i-cache/) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/cache-aside) · [Project/code](https://github.com/golang/sync/blob/master/singleflight/singleflight.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-06).

### 07 — Secure BFF (Backend for Frontend)
**What:** Middleware for mobile/web clients handling auth and API composition.  
**Why:** A BFF can centralize client-specific composition and session handling; compare its cost with direct API access.  
**Implementation:** JWT validation, secure cookies, aggregate multiple backend calls.  
**What you'll learn:** Token security, session management, and API composition patterns.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/backends-for-frontends) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/backends-for-frontends) · [Project/code](https://docs.duendesoftware.com/bff/samples/) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-07).

### 08 — API Gateway Aggregator
**What:** Collect data from 5 microservices in parallel for a single response.  
**Why:** Client-side aggregation is slow and chatty.  
**Implementation:** Fan-out requests, handle partial failures, set deadlines.  
**What you'll learn:** Scatter-gather pattern, deadline propagation, and partial failure handling.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/gateway-aggregation) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/gateway-aggregation) · [Project/code](https://github.com/krakend/playground-community) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-08).

---

## Block III: Data Engineering Under Load

*When one database isn't enough.*

### 09 — Terabyte Data Mocker
**What:** Generate millions of rows and optimize insertion.  
**Why:** Test data pipelines without real production data.  
**Implementation:** Bulk insert, streaming, index tuning.  
**What you'll learn:** Batch processing, COPY protocol, and indexing strategies.

**Explore:** [EN: theory/article](https://www.postgresql.org/docs/current/populate.html) · [RU: material](https://postgrespro.ru/docs/postgresql/current/sql-copy) · [Project/code](https://github.com/jackc/pgx/blob/master/copy_from.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-09).

### 10 — Read/Write Splitter
**What:** Route reads to replicas, writes to master.  
**Why:** Replicas can offload reads, with replication overhead and freshness trade-offs to measure.  
**Implementation:** Route by the operation’s consistency requirements, keep transactions on the appropriate node, and handle replication lag and read-your-writes.  
**What you'll learn:** Leader/follower architecture, eventual consistency, and lag monitoring.

**Explore:** [EN: theory/article](https://www.postgresql.org/docs/current/hot-standby.html) · [RU: material](https://postgrespro.ru/docs/postgresql/current/hot-standby) · [Project/code](https://github.com/postgresml/pgcat) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-10).

### 11 — Custom Database Sharder
**What:** Distribute data across multiple databases using consistent hashing.  
**Why:** Horizontal scaling requires data distribution.  
**Implementation:** Consistent hashing with virtual nodes, rebalancing logic.  
**What you'll learn:** Sharding strategies, hotspot prevention, and resharding challenges.

**Explore:** [EN: theory/article](https://cdn.amazon.science/ac/1d/eb50c4064c538c8ac440ce6a1d91/dynamo-amazons-highly-available-key-value-store.pdf) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/sharding) · [Project/code](https://github.com/buraksezer/consistent) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-11).

### 12 — SaaS Multitenancy Isolation
**What:** Isolate customer data in a shared cluster.  
**Why:** SaaS platforms need both isolation and efficiency.  
**Implementation:** Row-Level Security, schema-per-tenant, connection pooling.  
**What you'll learn:** Multi-tenancy patterns, security boundaries, and noisy neighbor prevention.

**Explore:** [EN: theory/article](https://www.postgresql.org/docs/16/ddl-rowsecurity.html) · [RU: material](https://postgrespro.ru/docs/postgresql/16/ddl-rowsecurity) · [Project/code](https://github.com/postgres/postgres/blob/master/src/backend/rewrite/rowsecurity.c) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-12).

---

## Block IV: Complex Patterns & Real-time

*Instant responses and integrity in distributed systems.*

### 13 — High-Load Chat Engine
**What:** Measure sustainable WebSocket connections and message fan-out under a specified resource budget.  
**Why:** Real-time features are expected in modern apps.  
**Implementation:** Redis Pub/Sub for transient cross-instance broadcast, connection management, and explicit reconnect behavior; durable delivery requires retained messages and replay.  
**What you'll learn:** WebSocket scaling, connection limits, and Pub/Sub patterns.

**Explore:** [EN: theory/article](https://redis.io/docs/latest/develop/pubsub/) · [RU: material](https://developer.mozilla.org/ru/docs/Web/API/WebSocket) · [Project/code](https://github.com/gorilla/websocket/blob/main/examples/chat/README.md) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-13).

### 14 — Real-time Leaderboard
**What:** Top players from millions of scores updated in real-time.  
**Why:** Gaming and competitive features need instant updates.  
**Implementation:** Redis Sorted Sets, atomic updates, pagination.  
**What you'll learn:** Sorted set operations, real-time aggregation, and memory efficiency.

**Explore:** [EN: theory/article](https://redis.io/docs/latest/develop/data-types/sorted-sets/) · [RU: material](https://aws.amazon.com/ru/elasticache/what-is-redis/) · [Project/code](https://raw.githubusercontent.com/redis/redis/unstable/src/t_zset.c) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-14).

### 15 — Distributed SAGA Orchestrator
**What:** Distributed transaction with compensation mechanisms.  
**Why:** Compare compensation-based workflows with coordinated transactions, including latency, blocking, and acceptable intermediate states.  
**Implementation:** Orchestration-based SAGA with compensation steps.  
**What you'll learn:** Distributed transaction patterns, compensating transactions, and failure recovery.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/saga) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/saga) · [Project/code](https://github.com/temporalio/samples-go/blob/main/saga/workflow.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-15).

### 16 — Circuit Breaker Service
**What:** Protect system from cascading failures.  
**Why:** One slow service shouldn't bring down everything.  
**Implementation:** Circuit states (closed/open/half-open), retry with backoff, bulkhead.  
**What you'll learn:** Failure detection, graceful degradation, and resilience patterns.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/circuit-breaker) · [Project/code](https://github.com/sony/gobreaker) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-16).

---

## Block V: Observability & The Grand Dashboard

*Understanding what happens under load.*

### 17 — Dynamic Feature Toggle
**What:** Enable/disable features without redeployment.  
**Why:** Controlled rollouts and kill switches, with measured propagation delay and fallback behavior.  
**Implementation:** Configuration service with Redis/ETCD, runtime updates.  
**What you'll learn:** Feature flag management, A/B testing infrastructure, and configuration distribution.

**Explore:** [EN: theory/article](https://docs.getunleash.io/get-started/unleash-overview) · [RU: material](https://learn.microsoft.com/ru-ru/dotnet/architecture/cloud-native/feature-flags) · [Project/code](https://github.com/Unleash/unleash) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-17).

### 18 — Log Aggregator (Mini ELK)
**What:** Collect and parse logs from all services in real-time.  
**Why:** Correlated, searchable logs help investigate behavior across components; evaluate delivery and storage costs.  
**Implementation:** Structured logging, log shipping, searchable storage.  
**What you'll learn:** Log formats, correlation IDs, and efficient log storage.

**Explore:** [EN: theory/article](https://go.dev/blog/slog) · [RU: material](https://learn.microsoft.com/ru-ru/dotnet/core/extensions/logging/overview) · [Project/code](https://github.com/grafana/alloy-scenarios/tree/main/app-instrumentation/logging/popular-logging-frameworks) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-18).

### 19 — System Metrics Exporter
**What:** Instrument all services with Prometheus metrics.  
**Why:** You can't improve what you don't measure.  
**Implementation:** RPS, latency (p95/p99), error rates, resource usage.  
**What you'll learn:** Metrics types (counters, gauges, histograms), instrumentation, and cardinality.

**Explore:** [EN: theory/article](https://prometheus.io/docs/practices/instrumentation/) · [RU: material](https://learn.microsoft.com/ru-ru/dotnet/core/diagnostics/metrics-instrumentation) · [Project/code](https://github.com/prometheus/client_golang) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-19).

### 20 — The Grand Dashboard
**What:** Visualize a selected integration scenario under load in Grafana, comparing baseline behavior and injected failures.  
**Why:** See the system as a whole, not individual parts.  
**Implementation:** Load testing with k6, real-time dashboards, comparison views.  
**What you'll learn:** Performance visualization, bottleneck identification, and system-wide observability.

**Explore:** [EN: theory/article](https://grafana.com/docs/grafana/latest/visualizations/dashboards/build-dashboards/best-practices/) · [RU: material](https://yandex.cloud/ru/docs/monitoring/operations/prometheus/querying/grafana) · [Project/code](https://github.com/grafana/quickpizza) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-20).

---

## Block VI: Message Brokers & Event-Driven Architecture

*Asynchronous communication and stream processing.*

### 21 — Kafka Exactly-Once Delivery
**What:** Compare Kafka transactional processing with deduplication of effects in external systems; identify the boundary of each guarantee.  
**Why:** Financial and inventory systems can't tolerate duplicates.  
**Implementation:** Idempotent producer, transactional output plus offsets, read_committed consumers, and an explicit inbox/outbox or destination-side idempotency design for external effects.  
**What you'll learn:** Kafka semantics, exactly-once trade-offs, and transactional boundaries.

**Explore:** [EN: theory/article](https://kafka.apache.org/41/design/design/#message-delivery-semantics) · [RU: material](https://habr.com/ru/companies/lamoda/articles/678932/) · [Project/code](https://github.com/apache/kafka/tree/trunk/examples) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-21).

### 22 — Event Sourcing Engine
**What:** System state derived from event history.  
**Why:** Audit trails and time travel queries.  
**Implementation:** Event store, snapshots, event replay.  
**What you'll learn:** Event sourcing vs CRUD, snapshot strategies, and CQRS integration.

**Explore:** [EN: theory/article](https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing) · [RU: material](https://learn.microsoft.com/ru-ru/azure/architecture/patterns/event-sourcing) · [Project/code](https://github.com/oskardudycz/EventSourcing.NetCore) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-22).

### 23 — Distributed Job Scheduler
**What:** Coordinate scheduled jobs across a cluster using leases, with explicit handling of retries and overlapping attempts.  
**Why:** Cron on every instance runs jobs N times.  
**Implementation:** Leases, leader election, durable job state, and destination-enforced fencing or idempotent effects; an expired lease does not stop the previous worker.  
**What you'll learn:** Distributed cron, lease management, and failover handling.

**Explore:** [EN: theory/article](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) · [RU: material](https://kubernetes.io/ru/docs/concepts/workloads/controllers/cron-jobs/) · [Project/code](https://github.com/go-co-op/gocron) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-23).

### 24 — Change Data Capture (CDC)
**What:** Stream database changes to other systems in real-time.  
**Why:** Keep caches, search indexes, and analytics in sync.  
**Implementation:** Debezium, Kafka Connect, transformation logic.  
**What you'll learn:** CDC internals, log-based change capture, and synchronization patterns.

**Explore:** [EN: theory/article](https://debezium.io/documentation/reference/stable/connectors/postgresql.html) · [RU: material](https://yandex.cloud/ru/docs/data-transfer/concepts/cdc) · [Project/code](https://raw.githubusercontent.com/debezium/debezium/main/debezium-connector-postgres/src/main/java/io/debezium/connector/postgresql/PostgresStreamingChangeEventSource.java) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-24).

---

## Block VII: High Performance & Networking

*When HTTP and JSON overhead is too much.*

### 25 — Custom TCP/UDP Proxy
**What:** Load balance traffic at the transport layer.  
**Why:** Transport-layer forwarding can avoid application-layer processing; compare measured overhead and required routing features.  
**Implementation:** Raw sockets, connection forwarding, health checks.  
**What you'll learn:** TCP handshake, connection pooling, and L4 load balancing algorithms.

**Explore:** [EN: theory/article](https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html) · [RU: material](https://nginx.org/ru/docs/stream/ngx_stream_proxy_module.html) · [Project/code](https://github.com/inetaf/tcpproxy) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-25).

### 26 — Zero-Copy File Server
**What:** Serve files without copying through application buffers.  
**Why:** Minimize CPU and memory for static content.  
**Implementation:** Compare buffered I/O and sendfile where supported; document the OS, TLS path, and page-cache state.  
**What you'll learn:** Kernel-assisted transfer, memory-mapped files, and I/O costs; sendfile still uses the kernel and is distinct from kernel bypass.

**Explore:** [EN: theory/article](https://man7.org/linux/man-pages/man2/sendfile.2.html) · [RU: material](https://nginx.org/ru/docs/http/ngx_http_core_module.html#sendfile) · [Project/code](https://go.dev/src/internal/poll/sendfile_unix.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-26).

### 27 — Binary Protocol Parser
**What:** Replace JSON with Protobuf/MessagePack.  
**Why:** Compare size and serialization cost for the same payloads, implementations, and compression settings.  
**Implementation:** Schema definition, code generation, performance comparison.  
**What you'll learn:** Serialization formats, binary wire protocols, and schema evolution.

**Explore:** [EN: theory/article](https://protobuf.dev/programming-guides/encoding/) · [RU: material](https://learn.microsoft.com/ru-ru/aspnet/core/grpc/versioning?view=aspnetcore-10.0) · [Project/code](https://github.com/protocolbuffers/protoscope) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-27).

---

## Block VIII: Security & Decentralization

*Security mechanisms and distributed consensus.*

### 28 — Distributed Lock Manager (Redlock)
**What:** Distributed locks across independent services.  
**Why:** Coordinate access to shared resources.  
**Implementation:** Compare Redis leases and Redlock assumptions; test renewal failure and stale owners. Fencing requires a separate reliable token source and enforcement by the protected resource.  
**What you'll learn:** Distributed locking pitfalls, clock drift, and the Redlock debate.

**Explore:** [EN: theory/article](https://redis.io/docs/latest/develop/clients/patterns/distributed-locks/) · [RU: material](https://habr.com/ru/companies/dododev/articles/709598/) · [Project/code](https://github.com/go-redsync/redsync/blob/master/mutex.go) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-28).

### 29 — Merkle Tree Validator
**What:** Verify integrity of millions of records.  
**Why:** Detect tampering and sync large datasets efficiently.  
**Implementation:** Hash tree construction, proof generation, verification.  
**What you'll learn:** Hash-based verification, tree synchronization, and blockchain fundamentals.

**Explore:** [EN: theory/article](https://www.rfc-editor.org/rfc/rfc9162.html) · [RU: material](https://ethereum.org/ru/developers/docs/data-structures-and-encoding/patricia-merkle-trie/) · [Project/code](https://github.com/transparency-dev/merkle) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-29).

### 30 — Hot/Cold Wallet Logic
**What:** Model asset custody, withdrawal limits, and approval workflows using simulated assets.  
**Why:** Explore signing boundaries, separation of duties, and auditability under a defined threat model.  
**Implementation:** Multi-signature, withdrawal queues, approval workflows.  
**What you'll learn:** Security architecture, transaction signing, and operational security.

**Explore:** [EN: theory/article](https://github.com/bitcoin/bitcoin/blob/master/doc/psbt.md) · [RU: material](https://bitcoin.org/ru/secure-your-wallet) · [Project/code](https://github.com/bitcoin/bitcoin/blob/master/doc/psbt.md) · [What to inspect and scope notes](docs/reading-map-01-30.md#task-30).

---

## Infrastructure

Proposed shared dependencies for integration experiments; use each implemented task’s README for current run commands:

```bash
infrastructure/
├── docker-compose.yml    # Postgres, Redis, Kafka, Prometheus, Grafana, ClickHouse
├── prometheus/           # Metrics collection configuration
├── grafana/              # Pre-built dashboards
├── k6/                   # Load testing scenarios
└── scripts/              # Benchmark utilities
```

**Target workflow for the planned shared infrastructure (not yet a runnable quick start):**

```bash
# Start all dependencies
make infra-up

# Run specific project (e.g., atomic-inventory in Go)
cd 01-atomic-inventory
make run-go

# Run load test
make load-test

# View metrics
open http://localhost:3000
```

---

## Areas to Explore

- **Concurrency:** Goroutines vs event loop, worker threads, atomic operations
- **Databases:** Transactions, isolation levels, locks, sharding, replication
- **Architecture:** CQRS, Event Sourcing, SAGA, Circuit Breaker, BFF, API Gateway
- **Message Brokers:** Kafka, guarantees, consumer groups
- **Observability:** Metrics, logs, tracing, profiling under load
- **Networking:** WebSockets, TCP/UDP, binary protocols, zero-copy
- **Security:** JWT, RLS, multi-signature, distributed locks

---

**⭐ If this roadmap helps you, star the repository! ⭐**
