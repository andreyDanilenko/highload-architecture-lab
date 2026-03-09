# Phase 2: Creator / Architect (Projects 31-60)

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Rust](https://img.shields.io/badge/Rust-2021+-DEA584?style=flat-square&logo=rust)](https://www.rust-lang.org/)
[![C](https://img.shields.io/badge/C-17-A8B9CC?style=flat-square&logo=c)](https://en.wikipedia.org/wiki/C_(programming_language))
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28-326CE5?style=flat-square&logo=kubernetes)](https://kubernetes.io/)
[![LLVM](https://img.shields.io/badge/LLVM-16-262D3A?style=flat-square&logo=llvm)](https://llvm.org/)
[![Linux](https://img.shields.io/badge/Linux-Kernel-FCC624?style=flat-square&logo=linux)](https://kernel.org/)

A structured roadmap of 30 advanced engineering challenges. Each project moves you from using technologies to creating them — databases, compilers, consensus algorithms, and network protocols from scratch.

---

## The Mindset Shift

In Phase 1, you learned to **use** technologies. You built microservices, implemented patterns, and made systems work.

In Phase 2, you learn to **create** technologies. You'll build databases from scratch, implement consensus algorithms, write compilers, and touch the kernel.

**This is the difference between an engineer and an architect. Between a driver and a mechanic. Between someone who uses tools and someone who builds them.**

---

## Repository Structure

```
phase-2/
├── infrastructure/          # Shared dependencies for Phase 2
├── 31-k8s-operator/         # Sprint 10: Infrastructure
├── 32-service-mesh/
├── 33-custom-scheduler/
├── 34-multi-cluster/
├── 35-lsm-tree/             # Sprint 11: Databases from Scratch
├── 36-btree-engine/
├── 37-sql-parser/
├── 38-query-optimizer/
├── 39-mvcc/
├── 40-distributed-txn/
├── 41-vector-db/
├── 42-raft/                 # Sprint 12: Distributed Algorithms
├── 43-paxos/
├── 44-gossip/
├── 45-swim/
├── 46-dht-chord/
├── 47-crdt/
├── 48-vector-clocks/
├── 49-tcp-stack/            # Sprint 13: Networking Deep Dive
├── 50-quic/
├── 51-l7-load-balancer/
├── 52-sidecar-proxy/
├── 53-sdn-controller/
├── 54-vpn-protocol/
├── 55-interpreter/          # Sprint 14: Languages & Compilers
├── 56-garbage-collector/
├── 57-jit-compiler/
├── 58-llvm-frontend/
├── 59-go-runtime-hack/
└── 60-virtual-machine/
```

---

## Sprint 10: Infrastructure & Orchestration (Projects 31-34)

*Moving from deploying on Kubernetes to extending Kubernetes itself.*

### 31 — Kubernetes Operator
**What:** Build a custom operator that manages a complex application (e.g., PostgreSQL cluster, Redis sentinel).  
**Why:** Operators are the standard for running stateful workloads on K8s.  
**Implementation:** Controller-runtime, CRDs, reconciliation loop, status handling.  
**What you'll learn:** Kubernetes internals, custom resources, controller patterns, operator SDK.

### 32 — Service Mesh (Istio/Linkerd)
**What:** Deploy and configure a service mesh, then implement a custom traffic policy.  
**Why:** Service meshes provide observability, security, and control for microservices.  
**Implementation:** Sidecar injection, traffic routing, mTLS, telemetry.  
**What you'll learn:** Sidecar pattern, Envoy proxy, mTLS, distributed tracing.

### 33 — Custom Kubernetes Scheduler
**What:** Write your own scheduler that places pods based on custom logic (e.g., data locality, cost optimization).  
**Why:** Default scheduler doesn't fit every workload.  
**Implementation:** Scheduler framework, plugin architecture, binding cycles.  
**What you'll learn:** Scheduling algorithms, plugin system, Kubernetes architecture.

### 34 — Multi-Cluster Management
**What:** Deploy an application across multiple Kubernetes clusters with failover.  
**Why:** Global scale requires multiple clusters.  
**Implementation:** Cluster registration, cross-cluster service discovery, failover logic.  
**What you'll learn:** Federation, cluster registration, cross-cluster networking, disaster recovery.

---

## Sprint 11: Databases from Scratch (Projects 35-41)

*Stop using databases. Start understanding them.*

### 35 — LSM-tree Storage Engine
**What:** Build a LevelDB/RocksDB-like engine with memtables, SSTables, and compaction.  
**Why:** LSM-trees power most modern write-optimized databases.  
**Implementation:** Skip lists for memtables, sorted files, compaction strategies, bloom filters.  
**What you'll learn:** Write amplification, read amplification, compaction, bloom filters.

### 36 — B+Tree Storage Engine
**What:** Implement a B+Tree with buffer pool management and WAL.  
**Why:** B+Trees power most traditional databases (PostgreSQL, MySQL).  
**Implementation:** Node splitting, merging, buffer pool, write-ahead logging, crash recovery.  
**What you'll learn:** Page organization, cache management, crash recovery.

### 37 — SQL Parser
**What:** Write a parser that converts SQL into an AST (Abstract Syntax Tree).  
**Why:** Every database needs to understand SQL.  
**Implementation:** Lexer, parser (recursive descent or parser combinator), AST nodes.  
**What you'll learn:** Grammar definition, parsing techniques, AST representation.

### 38 — Query Optimizer
**What:** Build a cost-based optimizer that chooses the best execution plan.  
**Why:** The optimizer makes or breaks database performance.  
**Implementation:** Statistics collection, cost models, plan enumeration, join ordering.  
**What you'll learn:** Cardinality estimation, join algorithms, plan selection.

### 39 — MVCC (Multi-Version Concurrency Control)
**What:** Implement MVCC for a toy database to allow reads without blocking writes.  
**Why:** MVCC is how modern databases achieve high concurrency.  
**Implementation:** Version chains, visibility rules, garbage collection, snapshot isolation.  
**What you'll learn:** Snapshot isolation, read consistency, version storage.

### 40 — Distributed Transaction Coordinator
**What:** Build a 2PC or Percolator (Google Spanner-style) transaction coordinator.  
**Why:** Distributed transactions are hard but necessary for consistency.  
**Implementation:** Transaction manager, participant coordination, recovery protocols.  
**What you'll learn:** Two-phase commit, Percolator model, failure handling.

### 41 — Vector Database Engine
**What:** Implement approximate nearest neighbor search with HNSW indexes.  
**Why:** Vector databases are essential for AI applications.  
**Implementation:** Embeddings storage, HNSW graph construction, similarity search.  
**What you'll learn:** Vector similarity, ANN algorithms, HNSW internals.

---

## Sprint 12: Distributed Algorithms (Projects 42-48)

*The math behind distributed systems.*

### 42 — Raft Consensus
**What:** Full Raft implementation: leader election, log replication, snapshotting.  
**Why:** Raft powers etcd, Consul, and many other distributed systems.  
**Implementation:** Follower/candidate/leader states, heartbeats, log matching, safety.  
**What you'll learn:** Consensus, quorum, log replication, cluster membership changes.

### 43 — Paxos
**What:** Implement classic Paxos (simplified, but working).  
**Why:** Paxos is the foundation of distributed consensus theory.  
**Implementation:** Proposers, acceptors, learners, phases 1 & 2.  
**What you'll learn:** The elegance and complexity of Paxos, why Raft exists.

### 44 — Gossip Protocol
**What:** Build a gossip-based membership and broadcast system.  
**Why:** Gossip powers Cassandra, Redis Cluster, and many decentralized systems.  
**Implementation:** Period propagation, infection-style dissemination, failure detection.  
**What you'll learn:** Epidemic broadcast, convergence, load balancing.

### 45 — SWIM Membership Protocol
**What:** Implement the SWIM protocol (used in Consul, Serf).  
**Why:** SWIM adds scalable failure detection to gossip.  
**Implementation:** Ping/ack, indirect probing, suspicion mechanism.  
**What you'll learn:** Scalable failure detection, suspicion, network partitioning.

### 46 — DHT (Chord)
**What:** Build a Chord Distributed Hash Table.  
**Why:** DHTs power many P2P systems and distributed caches.  
**Implementation:** Consistent hashing, finger tables, node joins/leaves, stabilization.  
**What you'll learn:** Ring-based routing, lookup efficiency, stabilization.

### 47 — CRDTs (Conflict-free Replicated Data Types)
**What:** Implement G-Counter, PN-Counter, OR-Set with sync without coordination.  
**Why:** CRDTs enable multi-master replication without conflicts.  
**Implementation:** State-based and operation-based CRDTs, merge functions.  
**What you'll learn:** Commutative operations, eventual consistency without coordination.

### 48 — Vector Clocks
**What:** Implement vector clocks to track causality in distributed systems.  
**Why:** Understanding causality is key to debugging distributed systems.  
**Implementation:** Clock updates, comparison, concurrent vs happened-before detection.  
**What you'll learn:** Causality, partial ordering, version vectors.

---

## Sprint 13: Networking Deep Dive (Projects 49-54)

*Below HTTP. Below TCP. To the wire.*

### 49 — TCP/IP Stack from Scratch
**What:** Implement a minimal TCP/IP stack on raw sockets (or TUN device).  
**Why:** Understand what happens between `socket()` and `write()`.  
**Implementation:** IP packet handling, TCP state machine, congestion control basics.  
**What you'll learn:** SYN/ACK, sequence numbers, windowing, retransmission.

### 50 — QUIC Protocol
**What:** Build a basic QUIC implementation over UDP.  
**Why:** QUIC is the future (HTTP/3).  
**Implementation:** Connection establishment, streams, packet encryption, 0-RTT.  
**What you'll learn:** Stream multiplexing, 0-RTT, connection migration.

### 51 — L7 Load Balancer
**What:** Build an HTTP load balancer with caching, health checks, and sticky sessions.  
**Why:** Understand how your traffic gets distributed.  
**Implementation:** Reverse proxy, round-robin/least-connections, health checking.  
**What you'll learn:** Load balancing algorithms, connection pooling, timeouts.

### 52 — Sidecar Proxy
**What:** Implement a simple sidecar proxy that intercepts traffic and adds telemetry.  
**Why:** Sidecars power service meshes.  
**Implementation:** Transparent interception, metrics collection, retry logic.  
**What you'll learn:** Proxy patterns, interception techniques, observability.

### 53 — SDN Controller (OpenFlow)
**What:** Build a basic SDN controller that programs switch flow tables.  
**Why:** Software-defined networking separates control from data plane.  
**Implementation:** OpenFlow protocol, topology discovery, flow programming.  
**What you'll learn:** Control plane vs data plane, flow tables, network virtualization.

### 54 — VPN Protocol
**What:** Implement a simple VPN tunnel (like WireGuard, but simpler).  
**Why:** Understand how secure tunnels work.  
**Implementation:** TUN device, encryption, packet forwarding, MTU handling.  
**What you'll learn:** Tunneling, encryption in transit, MTU issues.

---

## Sprint 14: Languages & Compilers (Projects 55-60)

*Stop using languages. Start creating them.*

### 55 — Interpreter
**What:** Write an interpreter for a simple language (lexer, parser, AST, eval).  
**Why:** Understand how your code runs.  
**Implementation:** Tokenization, recursive descent parsing, environment, evaluation.  
**What you'll learn:** Syntax vs semantics, scope, closures.

### 56 — Garbage Collector
**What:** Implement a mark-and-sweep or generational GC for your interpreter.  
**Why:** Automatic memory management is magic until you build it.  
**Implementation:** Root scanning, mark phase, sweep phase, object headers.  
**What you'll learn:** Reachability, GC algorithms, stop-the-world.

### 57 — JIT Compiler
**What:** Add a simple JIT to your interpreter (compile hot paths to machine code).  
**Why:** JITs make languages fast.  
**Implementation:** IR generation, machine code emission, calling convention.  
**What you'll learn:** Runtime compilation, profiling, optimization.

### 58 — LLVM Frontend
**What:** Write a frontend that compiles your language to LLVM IR.  
**Why:** LLVM lets you focus on the language, not the backend.  
**Implementation:** AST → LLVM IR, type system lowering, standard library.  
**What you'll learn:** SSA form, LLVM infrastructure, backend independence.

### 59 — Go Runtime Hack
**What:** Modify the Go scheduler or add a custom garbage collector hook.  
**Why:** Understand the runtime your Go code runs on.  
**Implementation:** Fork Go, modify runtime (M/P/G model), build custom compiler.  
**What you'll learn:** M/P/G model, stack management, runtime internals.

### 60 — Virtual Machine
**What:** Build a stack-based VM (like JVM or CPython bytecode).  
**Why:** VMs power most languages.  
**Implementation:** Bytecode design, interpreter loop, stack operations, function calls.  
**What you'll learn:** Bytecode, operand stack, frame management.

---

## How Phase 2 Differs from Phase 1

| Aspect | Phase 1 (01-30) | Phase 2 (31-60) |
|--------|-----------------|-----------------|
| **Focus** | Using technologies | Creating technologies |
| **Code** | Microservices, APIs | Databases, compilers, kernels |
| **Depth** | Application layer | Systems layer |
| **Languages** | Go, Node.js | Go, Rust, C, Assembly |
| **Mindset** | "How do I use X?" | "How would I build X?" |
| **Outcome** | Senior Engineer | Architect / Creator |

---

## Infrastructure for Phase 2

```bash
phase-2/infrastructure/
├── docker-compose.yml    # Kafka, Redis, Postgres (still useful)
├── kind/                 # Local Kubernetes for operator testing
├── qemu/                 # For OS development
├── llvm/                 # LLVM toolchain
└── scripts/              # Build and test utilities
```

**Quick start:**

```bash
# Clone the repo
git clone https://github.com/yourname/107-challenges
cd phase-2

# Start infrastructure
make infra-up

# Pick a project (e.g., 35-lsm-tree)
cd 35-lsm-tree
make run
make test
```

---

## What You'll Become

After Phase 2, you are no longer just a backend engineer.

**You can:**
- Read the source code of PostgreSQL, Redis, or Kubernetes and understand why they're built that way
- Contribute to open source projects at the core level
- Build your own database if existing ones don't fit your needs
- Design a new protocol or algorithm
- Teach others not just how to use tools, but how to create them

**You become the person who creates the technologies that Phase 1 engineers use.**

---

## What You'll Master

- **Infrastructure:** Kubernetes operators, service mesh, custom schedulers, multi-cluster
- **Databases:** LSM-trees, B+Trees, SQL parsers, query optimizers, MVCC, vector databases
- **Distributed Systems:** Raft, Paxos, gossip, SWIM, CRDTs, vector clocks
- **Networking:** TCP stacks, QUIC, load balancers, sidecar proxies, SDN, VPN
- **Languages:** Interpreters, garbage collectors, JIT compilers, LLVM, virtual machines
- **Operating Systems:** Microkernels, filesystems, device drivers, hypervisors, containers

---

## Next Steps

When you complete Phase 2, you're ready for:

### Phase 3: Master / Distinguished Engineer (Projects 61-107)
- Global-scale systems
- AI-powered infrastructure
- Quantum-resistant cryptography
- Research-level systems design
- Open source contribution to major projects
- Building tools from scratch (Git, Docker, grep, curl)

---

**⭐ Phase 1 made you a senior. Phase 2 makes you a creator. The journey continues. ⭐**
