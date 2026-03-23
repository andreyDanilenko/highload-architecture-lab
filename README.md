# 30 Highload Engineering Challenges

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Node.js](https://img.shields.io/badge/Node.js-20+-339933?style=flat-square&logo=nodedotjs)](https://nodejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=flat-square&logo=redis)](https://redis.io/)
[![Kafka](https://img.shields.io/badge/Kafka-3.5-231F20?style=flat-square&logo=apache-kafka)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)

### Hi, my name is Andrey 
and I have been a frontend developer for 4 years. To be honest, I always thought that an IT position was really easy for most tasks. But once I got a good position at a good company, I relaxed and forgot that a good developer must always learn something new.

Today, when my company lost the project actually, I decided to continue my journey. This guide is for developers who share my ideas and want to become better. This guide will be helpful for other positions, not only backend, because it touches a lot of topics about building systems.

Sometimes people believe that AI is our enemy, but fundamental knowledge will always be relevant. This guide was created with help from AI, but I think that's the least benefit you can get from AI.

IMPORTANT!
When you use AI, you can make one small mistake. At the moment when you get an answer to your question, you need to carefully read and update your code. Otherwise, you don't learn anything and continued education will not be healthy. It creates the impression that you know a lot. That's right in part, but...

A structured roadmap of 30 hands-on engineering challenges. Each project solves a real distributed systems problem and builds upon the previous ones. No theory without practice — you write code, break things, and learn why systems are designed the way they are.

---

**Progress:** ▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 6.7% (2/30)

---

## The Philosophy

These challenges exist because building production systems requires more than knowing syntax. You'll learn to:

- Choose between competing strategies (pessimistic vs optimistic locks, fixed vs sliding windows)
- Design for failure (retries, circuit breakers, timeouts)
- Make conscious trade-offs (consistency vs availability, latency vs throughput)
- See the whole system, not just your service

Each challenge is implemented in both Go and Node.js to understand how language concurrency models affect architecture decisions.

---

## Repository Structure

```
/
├── infrastructure/        # Shared dependencies (Postgres, Redis, Kafka, Prometheus, Grafana)
├── 01-atomic-inventory/  # ✅ Foundation: concurrency strategies
├── 02-anti-bruteforce/   # ✅ Rate limiting with Redis
├── 03-heavy-worker/      # Worker pools and semaphores
├── 04-idempotency/       # Request deduplication
├── 05-rate-limiter/      # Cluster-wide Redis limiter
├── 06-multilayer-cache/  # L1 (memory) + L2 (Redis)
├── 07-bff/               # Backend for frontend with JWT
├── 08-gateway/           # Router + load balancer
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
├── 21-kafka-exactly-once/# Exactly-once delivery
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
**What:** Flash sale: deduct 1000 items under 100k concurrent requests without going negative.  
**Why:** Understand race conditions and locking strategies.  
**Implementation:** Compare pessimistic locks (`SELECT FOR UPDATE`), optimistic locks (version column), and Redis atomic operations.  
**What you'll learn:** Transaction isolation levels, deadlocks, and when to use each locking strategy.

### 02 — Anti-Bruteforce Vault ✅
**What:** PIN-protected vault with progressive delays on failed attempts.  
**Why:** Rate limiting is everywhere (login, password reset, 2FA).  
**Implementation:** Sliding window log in Redis with Lua scripts for atomicity.  
**What you'll learn:** Atomic Redis operations, sliding window vs fixed window, and writing Lua scripts.

### 03 — Heavy Task Worker Pool
**What:** Process background jobs with controlled resource consumption.  
**Why:** Prevent system overload from unpredictable workloads.  
**Implementation:** Worker pool with semaphores, task queue, graceful shutdown.  
**What you'll learn:** Goroutines vs event loop, backpressure, and task prioritization.

### 04 — Idempotency Key Provider
**What:** Middleware guaranteeing exactly-one execution of operations.  
**Why:** Payment retries shouldn't charge twice.  
**Implementation:** Store request IDs in Redis with TTL, atomic check-and-set.  
**What you'll learn:** Idempotency patterns, exactly-once semantics, and idempotency key lifecycle.

---

## Block II: Distributed Systems & Networking

*Making multiple servers work as one organism.*

### 05 — Distributed Rate Limiter
**What:** Rate limit across a cluster, not per instance.  
**Why:** With load balancing, per-instance limits are useless.  
**Implementation:** Redis-based sliding window with Lua.  
**What you'll learn:** Distributed state management, clock synchronization issues, and atomic scripts.

### 06 — Multilayer Cache
**What:** L1 (in-memory) + L2 (Redis) with cache stampede protection.  
**Why:** Cache misses at high load can kill databases.  
**Implementation:** Single flight pattern, probabilistic early expiration.  
**What you'll learn:** Cache strategies (cache-aside, read-through), stampede prevention, and consistency trade-offs.

### 07 — Secure BFF (Backend for Frontend)
**What:** Middleware for mobile/web clients handling auth and API composition.  
**Why:** Frontends shouldn't talk directly to microservices.  
**Implementation:** JWT validation, secure cookies, aggregate multiple backend calls.  
**What you'll learn:** Token security, session management, and API composition patterns.

### 08 — API Gateway Aggregator
**What:** Collect data from 5 microservices in parallel for a single response.  
**Why:** Client-side aggregation is slow and chatty.  
**Implementation:** Fan-out requests, handle partial failures, set deadlines.  
**What you'll learn:** Scatter-gather pattern, deadline propagation, and partial failure handling.

---

## Block III: Data Engineering Under Load

*When one database isn't enough.*

### 09 — Terabyte Data Mocker
**What:** Generate millions of rows and optimize insertion.  
**Why:** Test data pipelines without real production data.  
**Implementation:** Bulk insert, streaming, index tuning.  
**What you'll learn:** Batch processing, COPY protocol, and indexing strategies.

### 10 — Read/Write Splitter
**What:** Route reads to replicas, writes to master.  
**Why:** Scale reads without affecting write performance.  
**Implementation:** Detect statement type, handle replication lag.  
**What you'll learn:** Leader/follower architecture, eventual consistency, and lag monitoring.

### 11 — Custom Database Sharder
**What:** Distribute data across multiple databases using consistent hashing.  
**Why:** Horizontal scaling requires data distribution.  
**Implementation:** Consistent hashing with virtual nodes, rebalancing logic.  
**What you'll learn:** Sharding strategies, hotspot prevention, and resharding challenges.

### 12 — SaaS Multitenancy Isolation
**What:** Isolate customer data in a shared cluster.  
**Why:** SaaS platforms need both isolation and efficiency.  
**Implementation:** Row-Level Security, schema-per-tenant, connection pooling.  
**What you'll learn:** Multi-tenancy patterns, security boundaries, and noisy neighbor prevention.

---

## Block IV: Complex Patterns & Real-time

*Instant responses and integrity in distributed systems.*

### 13 — High-Load Chat Engine
**What:** Handle 50k+ WebSocket connections with message broadcasting.  
**Why:** Real-time features are expected in modern apps.  
**Implementation:** Redis Pub/Sub for cross-instance broadcast, connection management.  
**What you'll learn:** WebSocket scaling, connection limits, and Pub/Sub patterns.

### 14 — Real-time Leaderboard
**What:** Top players from millions of scores updated in real-time.  
**Why:** Gaming and competitive features need instant updates.  
**Implementation:** Redis Sorted Sets, atomic updates, pagination.  
**What you'll learn:** Sorted set operations, real-time aggregation, and memory efficiency.

### 15 — Distributed SAGA Orchestrator
**What:** Distributed transaction with compensation mechanisms.  
**Why:** Two-phase commit doesn't scale.  
**Implementation:** Orchestration-based SAGA with compensation steps.  
**What you'll learn:** Distributed transaction patterns, compensating transactions, and failure recovery.

### 16 — Circuit Breaker Service
**What:** Protect system from cascading failures.  
**Why:** One slow service shouldn't bring down everything.  
**Implementation:** Circuit states (closed/open/half-open), retry with backoff, bulkhead.  
**What you'll learn:** Failure detection, graceful degradation, and resilience patterns.

---

## Block V: Observability & The Grand Dashboard

*Understanding what happens under load.*

### 17 — Dynamic Feature Toggle
**What:** Enable/disable features without redeployment.  
**Why:** Safe rollouts and instant kill switches.  
**Implementation:** Configuration service with Redis/ETCD, runtime updates.  
**What you'll learn:** Feature flag management, A/B testing infrastructure, and configuration distribution.

### 18 — Log Aggregator (Mini ELK)
**What:** Collect and parse logs from all services in real-time.  
**Why:** Debugging distributed systems requires centralized logs.  
**Implementation:** Structured logging, log shipping, searchable storage.  
**What you'll learn:** Log formats, correlation IDs, and efficient log storage.

### 19 — System Metrics Exporter
**What:** Instrument all services with Prometheus metrics.  
**Why:** You can't improve what you don't measure.  
**Implementation:** RPS, latency (p95/p99), error rates, resource usage.  
**What you'll learn:** Metrics types (counters, gauges, histograms), instrumentation, and cardinality.

### 20 — The Grand Dashboard
**What:** Visualize all 19 services under load in Grafana.  
**Why:** See the system as a whole, not individual parts.  
**Implementation:** Load testing with k6, real-time dashboards, comparison views.  
**What you'll learn:** Performance visualization, bottleneck identification, and system-wide observability.

---

## Block VI: Message Brokers & Event-Driven Architecture

*Asynchronous communication and stream processing.*

### 21 — Kafka Exactly-Once Delivery
**What:** Pipeline guaranteeing no duplicate message processing.  
**Why:** Financial and inventory systems can't tolerate duplicates.  
**Implementation:** Idempotent producer, transactional API, consumer idempotency.  
**What you'll learn:** Kafka semantics, exactly-once trade-offs, and transactional boundaries.

### 22 — Event Sourcing Engine
**What:** System state derived from event history.  
**Why:** Audit trails and time travel queries.  
**Implementation:** Event store, snapshots, event replay.  
**What you'll learn:** Event sourcing vs CRUD, snapshot strategies, and CQRS integration.

### 23 — Distributed Job Scheduler
**What:** Schedule jobs that run exactly once across a cluster.  
**Why:** Cron on every instance runs jobs N times.  
**Implementation:** Distributed locks, leader election, job coordination.  
**What you'll learn:** Distributed cron, lease management, and failover handling.

### 24 — Change Data Capture (CDC)
**What:** Stream database changes to other systems in real-time.  
**Why:** Keep caches, search indexes, and analytics in sync.  
**Implementation:** Debezium, Kafka Connect, transformation logic.  
**What you'll learn:** CDC internals, log-based change capture, and synchronization patterns.

---

## Block VII: High Performance & Networking

*When HTTP and JSON overhead is too much.*

### 25 — Custom TCP/UDP Proxy
**What:** Load balance traffic at the transport layer.  
**Why:** L4 proxies are faster than L7.  
**Implementation:** Raw sockets, connection forwarding, health checks.  
**What you'll learn:** TCP handshake, connection pooling, and L4 load balancing algorithms.

### 26 — Zero-Copy File Server
**What:** Serve files without copying through application buffers.  
**Why:** Minimize CPU and memory for static content.  
**Implementation:** sendfile syscall, DMA, efficient streaming.  
**What you'll learn:** Kernel bypass, memory-mapped files, and I/O optimization.

### 27 — Binary Protocol Parser
**What:** Replace JSON with Protobuf/MessagePack.  
**Why:** Smaller payloads, faster serialization.  
**Implementation:** Schema definition, code generation, performance comparison.  
**What you'll learn:** Serialization formats, binary wire protocols, and schema evolution.

---

## Block VIII: Security & Decentralization

*Security mechanisms and distributed consensus.*

### 28 — Distributed Lock Manager (Redlock)
**What:** Distributed locks across independent services.  
**Why:** Coordinate access to shared resources.  
**Implementation:** Redlock algorithm with Redis, lock renewal, fencing tokens.  
**What you'll learn:** Distributed locking pitfalls, clock drift, and the Redlock debate.

### 29 — Merkle Tree Validator
**What:** Verify integrity of millions of records.  
**Why:** Detect tampering and sync large datasets efficiently.  
**Implementation:** Hash tree construction, proof generation, verification.  
**What you'll learn:** Hash-based verification, tree synchronization, and blockchain fundamentals.

### 30 — Hot/Cold Wallet Logic
**What:** Secure asset storage with withdrawal limits.  
**Why:** Protect funds in production systems.  
**Implementation:** Multi-signature, withdrawal queues, approval workflows.  
**What you'll learn:** Security architecture, transaction signing, and operational security.

---

## Infrastructure

Shared dependencies for all projects:

```bash
infrastructure/
├── docker-compose.yml    # Postgres, Redis, Kafka, Prometheus, Grafana, ClickHouse
├── prometheus/           # Metrics collection configuration
├── grafana/              # Pre-built dashboards
├── k6/                   # Load testing scenarios
└── scripts/              # Benchmark utilities
```

**Quick start:**

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

## What You'll Master

- **Concurrency:** Goroutines vs event loop, worker threads, atomic operations
- **Databases:** Transactions, isolation levels, locks, sharding, replication
- **Architecture:** CQRS, Event Sourcing, SAGA, Circuit Breaker, BFF, API Gateway
- **Message Brokers:** Kafka, guarantees, consumer groups
- **Observability:** Metrics, logs, tracing, profiling under load
- **Networking:** WebSockets, TCP/UDP, binary protocols, zero-copy
- **Security:** JWT, RLS, multi-signature, distributed locks

---

**⭐ If this roadmap helps you, star the repository! ⭐**
