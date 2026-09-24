# Implementation Plan: Building Scalable Backend Systems (Go + Node.js)

*A catalogue of possible experiments. The sprint groups below are browsing aids, not a schedule or a required sequence. Check task READMEs for current implementation status.*

[Articles and real projects for 01–30](docs/reading-map-01-30.md) · [Quality and scope review](docs/quality-review-01-30.md) · [EventLab and connected playgrounds for all 107 topics](docs/project-playground-01-107.md)

From task 05, the default application is **EventLab: events, bookings and tickets**. It can grow through selected challenges; other experiments can become independent projects and connect later through a small interface or shared dataset. The 01–04 examples remain useful standalone baselines. The scenarios below are plans, not implemented integration.

---

## Before You Start

**Tech Stack:** Go 1.27+, Node.js 20+, PostgreSQL 16, Redis 7.2, Kafka 3.5, Docker, Prometheus, Grafana

**Prerequisites:** Basic knowledge of at least one backend language, SQL, and REST APIs.

---

## Sprint 1: Foundation — Concurrency & Consistency

**Goal:** Investigate race conditions, locks, and data integrity through reproducible experiments.

| Project | What You Build |
|---------|----------------|
| 01 Atomic Inventory Counter | Three concurrency strategies (pessimistic, optimistic, Redis) |
| 02 Anti-Bruteforce Vault | Sliding window rate limiter with Redis Lua |
| 03 Heavy Task Worker | Worker pool with semaphores and graceful shutdown |
| 04 Idempotency Key Provider | Request deduplication with explicit effect, persistence, and retry boundaries |

**Known Challenges:**
- Deadlocks in pessimistic locking (Project 1)
- Races between separate Redis operations without an atomic protocol (Project 2)
- Goroutine leaks in worker pools (Project 3)
- TTL vs permanent storage for idempotency keys (Project 4)

**Integration idea — EventLab foundations:**
- Reserve a seat using a transaction strategy from 01
- Protect the organizer login using the experiments from 02
- Generate a ticket or report through the worker pool from 03
- Compare repeated booking/payment commands using 04; the payment stays simulated

---

## Sprint 2: Distributed Systems — From One to Many

**Goal:** Scale from single instance to cluster-aware services.

| Project | What You Build |
|---------|----------------|
| 05 Distributed Rate Limiter | Cluster-wide rate limiting with Redis |
| 06 Multilayer Cache | L1 (memory) + L2 (Redis) with stampede protection |
| 07 Secure BFF | JWT handling and API composition |

**Known Challenges:**
- Clock skew in distributed rate limiting (Project 5)
- Cache stampede under high load (Project 6)
- Secure token storage in cookies vs localStorage (Project 7)

**Integration idea — EventLab API and client sessions:**
- Run two instances of the same small application with a shared quota
- Cache the event catalogue and observe stale data and concurrent misses
- Build the visitor/organizer BFF around one concrete client flow

---

## Sprint 3: API Gateway & Aggregation

**Goal:** Build the entry point for all microservices.

| Project | What You Build |
|---------|----------------|
| 08 API Gateway | Reverse proxy with load balancing and aggregation |

**Known Challenges:**
- Connection pooling and timeouts
- Partial failure handling (when one backend dies)
- Request/response size limits

**Integration idea — EventLab event page aggregation:**
- Compose event details, availability and auxiliary data
- Give the combined request a deadline and explicit partial-result policy
- Keep client-specific session handling in the BFF; study transport balancing separately

---

## Sprint 4: Data Scaling — Beyond One Database

**Goal:** Handle data volumes that exceed single database capacity.

| Project | What You Build |
|---------|----------------|
| 09 Terabyte Data Mocker | Bulk data generation with optimized inserts |
| 10 Read/Write Splitter | Master-replica routing with lag handling |
| 11 Custom Database Sharder | Consistent hashing across databases |
| 12 SaaS Multitenancy | Row-level security and tenant isolation |

**Known Challenges:**
- Replication lag breaking read-your-writes (Project 10)
- Resharding without downtime (Project 11)
- Connection pool exhaustion with many tenants (Project 12)

**Integration idea — EventLab organizer data and reports:**
- Generate reproducible event/booking datasets
- Route reports and fresh booking reads according to their consistency needs
- Experiment with sharding one dataset and isolating two organizers using RLS

---

## Sprint 5: Real-Time Systems

**Goal:** Build low-latency, event-driven services.

| Project | What You Build |
|---------|----------------|
| 13 High-Load Chat Engine | WebSocket server with Redis Pub/Sub |
| 14 Real-Time Leaderboard | Redis sorted sets for instant rankings |
| 15 Distributed SAGA | Orchestrated transactions with compensation |
| 16 Circuit Breaker | Failure protection with retry and bulkhead |

**Known Challenges:**
- WebSocket connection limits and scaling (Project 13)
- SAGA compensation failures (Project 15)
- Circuit breaker state transitions and recovery (Project 16)

**Integration idea — EventLab live experience:**
- Chat between participants and organizers with explicit reconnect behavior
- Rank popular events from a stream of synthetic sales
- Try reservation, simulated payment and ticket issuance as a small saga
- Measure the response to a slow or unavailable dependency

---

## Sprint 6: Observability — Making Invisible Visible

**Goal:** Understand what happens under load.

| Project | What You Build |
|---------|----------------|
| 17 Dynamic Feature Toggle | Runtime config without redeploy |
| 18 Log Aggregator | Centralized structured logging |
| 19 System Metrics Exporter | Prometheus instrumentation for all services |
| 20 Grand Dashboard | Grafana visualization under load |

**Known Challenges:**
- Feature flag propagation delay (Project 17)
- Log volume and storage costs (Project 18)
- High-cardinality metrics breaking Prometheus (Project 19)

**Integration idea — Observe one EventLab scenario:**
- Use flags to switch one behavior during an experiment
- Correlate logs and metrics for one booking flow
- Compare baseline and injected failure under the same load; add only panels that answer the question

---

## Sprint 7: Event-Driven Architecture

**Goal:** Compare asynchronous delivery and processing guarantees under load and failures.

| Project | What You Build |
|---------|----------------|
| 21 Kafka Exactly-Once | Kafka transaction boundaries and coordination with external destinations |
| 22 Event Sourcing | State from event history with snapshots |
| 23 Distributed Scheduler | Cluster-wide cron with leader election |
| 24 Change Data Capture | Stream database changes to Kafka |

**Known Challenges:**
- Exactly-once semantics vs performance (Project 21)
- Event schema evolution (Project 22)
- Split-brain in leader election (Project 23)
- CDC initial load vs continuous streaming (Project 24)

**Integration idea — EventLab history and projections:**
- Build a projection from booking events and test Kafka transaction boundaries
- Try event sourcing for one aggregate rather than every module
- Schedule expiration/report work with an explicit duplicate-effect policy
- Use CDC to update a search copy and inspect snapshot, lag and replay

---

## Sprint 8: High Performance

**Goal:** Optimize for latency and throughput.

| Project | What You Build |
|---------|----------------|
| 25 TCP/UDP Proxy | L4 load balancing with raw sockets |
| 26 Zero-Copy Server | Compare buffered I/O and sendfile on the chosen OS and transport |
| 27 Binary Protocol | Protobuf/MessagePack instead of JSON |

**Known Challenges:**
- TCP connection state management (Project 25)
- File descriptor limits (Project 26)
- Schema versioning in binary protocols (Project 27)

**Integration idea — NetLab serving EventLab files and messages:**
- Build a TCP proxy; treat UDP as a separate experiment
- Compare buffered I/O and a supported kernel-assisted path for the same archive
- Compare message formats and framing using the same data and compatibility tests

---

## Sprint 9: Security & Consensus

**Goal:** Investigate security and coordination mechanisms under an explicit threat and failure model.

| Project | What You Build |
|---------|----------------|
| 28 Distributed Lock (Redlock) | Cross-service coordination |
| 29 Merkle Tree | Integrity verification for large datasets |
| 30 Hot/Cold Wallet | Multi-signature asset protection |

**Known Challenges:**
- Clock dependency in Redlock (Project 28)
- Tree rebuilding performance (Project 29)
- Private key management (Project 30)

**Integration idea — EventLab asset and integrity experiments:**
- Study leases and stale owners on report generation before considering financial-like actions
- Check archive records against a trusted Merkle root
- Keep wallet approvals and signatures in a separate simulated-asset example

---

## Combine a selected scenario

Build an integration scenario from selected components and run it under load. A challenge can remain a library or an in-process module; use a separate service when the experiment needs a network boundary:

```
Load Generator (k6) → API Gateway (08) → Selected Components → Metrics (19) → Grafana (20)
```

**Questions to investigate:**
- How rate limiting protects auth endpoints
- Which failure propagation paths circuit breakers limit, and which require timeouts or concurrency limits
- How sharding distributes database load
- How CDC propagates changes, and how lag and lost retention affect cache freshness
- Under which I/O and transport conditions kernel-assisted transfer improves efficiency
- What happens to coordinated jobs when a lease expires while its worker is still running

---

## Quick Reference

| Sprint | Focus | Projects |
|--------|-------|----------|
| 1 | Concurrency | 01-04 |
| 2 | Distributed Systems | 05-07 |
| 3 | API Gateway | 08 |
| 4 | Data Scaling | 09-12 |
| 5 | Real-Time | 13-16 |
| 6 | Observability | 17-20 |
| 7 | Event-Driven | 21-24 |
| 8 | Performance | 25-27 |
| 9 | Security | 28-30 |

---

Choose a connection when it makes the mechanism easier to observe. A useful standalone experiment does not have to join the application.
