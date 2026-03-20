# Phase 3: Master / Distinguished Engineer (Projects 61-107)

[![Rust](https://img.shields.io/badge/Rust-2021+-DEA584?style=flat-square&logo=rust)](https://www.rust-lang.org/)
[![C++](https://img.shields.io/badge/C++-20-00599C?style=flat-square&logo=cplusplus)](https://isocpp.org/)
[![Assembly](https://img.shields.io/badge/Assembly-x86_64-525252?style=flat-square&logo=assemblyscript)](https://en.wikipedia.org/wiki/Assembly_language)
[![CUDA](https://img.shields.io/badge/CUDA-12-76B900?style=flat-square&logo=nvidia)](https://developer.nvidia.com/cuda-toolkit)
[![FPGA](https://img.shields.io/badge/FPGA-Verilog-5C2D91?style=flat-square&logo=amd)](https://en.wikipedia.org/wiki/Field-programmable_gate_array)
[![Quantum](https://img.shields.io/badge/Quantum-Qiskit-6929C4?style=flat-square&logo=ibm)](https://qiskit.org/)

*Beyond engineering. Beyond architecture. Into the realm where new technologies are born.*

---

## The Mindset Shift

In Phase 1, you learned to **use** technologies.  
In Phase 2, you learned to **create** technologies.  
In Phase 3, you learn to **invent** new paradigms.

This is the difference between an architect and a distinguished engineer. Between a mechanic and an inventor. Between someone who builds tools and someone who changes how tools are built.

**At this level, you're not solving problems — you're redefining what problems are worth solving.**

---

## What You'll Master in This Phase

| Area | Skills |
|------|--------|
| **Hardware/Software Co-design** | FPGA programming, Verilog, RTL, hardware accelerators |
| **Quantum Computing** | Quantum algorithms, Qiskit, quantum cryptography |
| **Bioinformatics** | Genomic data processing, protein folding, DNA storage |
| **Distributed Systems Theory** | New consensus algorithms, byzantine fault tolerance |
| **Programming Languages** | Language design, type system research, formal verification |
| **Operating Systems Research** | Unikernels, exokernels, capability-based security |
| **Cryptography Research** | Post-quantum crypto, fully homomorphic encryption |
| **AI/ML Infrastructure** | ML compilers, distributed training, model optimization |
| **Planetary Scale** | Global consensus, interplanetary networking |

---

## Repository Structure

```
phase-3/
├── infrastructure/           # Research environments, simulators
├── papers/                   # Academic papers you'll implement
├── notes/                    # Your research notes
│
├── 61-fpga-accelerator/      # Sprint 15: Hardware/Software
├── 62-riscv-core/
├── 63-gpu-compute-shader/
├── 64-homomorphic-hardware/
│
├── 65-quantum-algorithms/    # Sprint 16: Quantum Computing
├── 66-quantum-crypto/
├── 67-quantum-ml/
│
├── 68-genome-pipeline/       # Sprint 17: Bioinformatics
├── 69-protein-folding/
├── 70-dna-storage/
│
├── 71-bft-consensus/         # Sprint 18: Distributed Systems Research
├── 72-new-consensus/
├── 73-distributed-os/
├── 74-global-clock/
│
├── 75-dependent-types/       # Sprint 19: Programming Languages Research
├── 76-formal-verification/
├── 77-effect-systems/
├── 78-language-from-idea/
│
├── 79-unikernel/             # Sprint 20: OS Research
├── 80-capability-os/
├── 81-persistent-memory/
│
├── 82-post-quantum-crypto/   # Sprint 21: Cryptography Research
├── 83-fully-homomorphic/
├── 84-mpc-at-scale/
├── 85-zero-knowledge-vm/
│
├── 86-ml-compiler/           # Sprint 22: AI/ML Infrastructure
├── 87-distributed-training/
├── 88-model-optimizer/
├── 89-ai-scheduler/
│
├── 90-global-consensus/      # Sprint 23: Planetary Scale
├── 91-interplanetary-dtns/
├── 92-space-protocols/
│
├── 93-crdt-research/         # Sprint 24: Your Own Research
├── 94-novel-storage-engine/
├── 95-new-consensus-protocol/
├── 96-language-design/
├── 97-os-concept/
│
├── 98-postgres-core/         # Sprint 25: Open Source Mastery
├── 99-linux-kernel/
├── 100-llvm-contribution/
├── 101-kafka-kip-lead/
│
├── 102-gdb-from-scratch/     # Sprint 26: Ultimate Understanding
├── 103-linux-from-scratch/
├── 104-compiler-from-scratch/
├── 105-database-from-scratch/
├── 106-os-from-scratch/
└── 107-your-own-invention/
```

---

## Sprint 15: Hardware/Software Co-design (Projects 61-64)

*When software isn't fast enough — build hardware.*

### 61 — FPGA Accelerator
**What:** Implement a hardware accelerator for a specific algorithm (e.g., compression, encryption) on FPGA.  
**Why:** Understand how hardware can outperform software by orders of magnitude.  
**Implementation:** Verilog/VHDL, hardware description, simulation, synthesis.  
**What you'll learn:** Hardware design, pipelining, parallelism at gate level.

### 62 — RISC-V Core
**What:** Implement a simple RISC-V CPU core in Verilog.  
**Why:** Understand how processors execute your code.  
**Implementation:** Instruction fetch/decode/execute, pipeline, hazard handling.  
**What you'll learn:** CPU architecture, pipelining, hazard detection.

### 63 — GPU Compute Shader
**What:** Write CUDA/compute shaders for massive parallel computation.  
**Why:** Leverage GPUs for non-graphics workloads.  
**Implementation:** CUDA/OpenCL, thread hierarchy, memory models, optimization.  
**What you'll learn:** Parallel programming, GPU architecture, warp divergence.

### 64 — Homomorphic Encryption Hardware
**What:** Design hardware acceleration for homomorphic encryption operations.  
**Why:** FHE is too slow in software — hardware is the future.  
**Implementation:** Specialized arithmetic circuits, modular multiplication.  
**What you'll learn:** Hardware acceleration for cryptography, side-channel resistance.

---

## Sprint 16: Quantum Computing (Projects 65-67)

*The next paradigm.*

### 65 — Quantum Algorithms
**What:** Implement Shor's algorithm or Grover's search on quantum simulators.  
**Why:** Understand how quantum computing changes what's computable.  
**Implementation:** Qiskit, quantum circuits, superposition, entanglement.  
**What you'll learn:** Quantum gates, amplitude amplification, quantum Fourier transform.

### 66 — Quantum Cryptography
**What:** Implement BB84 quantum key distribution.  
**Why:** Quantum communication is unhackable by classical means.  
**Implementation:** Qiskit, quantum states, measurement, eavesdropping detection.  
**What you'll learn:** Quantum key distribution, no-cloning theorem, eavesdropping detection.

### 67 — Quantum Machine Learning
**What:** Build a quantum neural network for simple classification.  
**Why:** Quantum ML promises exponential speedups for certain problems.  
**Implementation:** Variational circuits, parameterized quantum circuits.  
**What you'll learn:** Quantum feature maps, variational algorithms, barren plateaus.

---

## Sprint 17: Bioinformatics (Projects 68-70)

*Computing at the intersection with biology.*

### 68 — Genomic Data Pipeline
**What:** Process DNA sequencing data: alignment, variant calling.  
**Why:** Genomics generates massive data requiring specialized algorithms.  
**Implementation:** BWT, FM-index, Smith-Waterman, parallel processing.  
**What you'll learn:** String algorithms for genomics, compression of genetic data.

### 69 — Protein Folding Simulation
**What:** Implement simplified protein folding models.  
**Why:** Understand how structure emerges from sequence.  
**Implementation:** Molecular dynamics, energy minimization, monte carlo methods.  
**What you'll learn:** Computational chemistry, parallel simulation, analysis.

### 70 — DNA Storage System
**What:** Encode/decode binary data into DNA sequences.  
**Why:** DNA is the densest storage medium known.  
**Implementation:** Encoding schemes, error correction for biological media.  
**What you'll learn:** Biological constraints, error correction, dense encoding.

---

## Sprint 18: Distributed Systems Research (Projects 71-74)

*Beyond Paxos and Raft.*

### 71 — Byzantine Fault Tolerance
**What:** Implement PBFT (Practical Byzantine Fault Tolerance).  
**Why:** Handle malicious nodes, not just crashes.  
**Implementation:** View changes, prepare/commit phases, authentication.  
**What you'll learn:** BFT consensus, quorums, fault models.

### 72 — New Consensus Protocol
**What:** Design and implement your own consensus protocol for a specific use case.  
**Why:** Existing protocols make trade-offs you might not need.  
**Implementation:** Novel approach, evaluation against existing protocols.  
**What you'll learn:** Research methodology, protocol design, evaluation.

### 73 — Distributed Operating System
**What:** Build an OS that treats a cluster as a single computer.  
**Why:** Future systems will be distributed by default.  
**Implementation:** Single system image, distributed process management.  
**What you'll learn:** OS design at cluster scale, location transparency.

### 74 — Global Clock Service
**What:** Implement TrueTime-like service (Google Spanner's secret sauce).  
**Why:** Global consistency requires global time.  
**Implementation:** GPS + atomic clocks, clock uncertainty, time synchronization.  
**What you'll learn:** Physical time, clock synchronization, uncertainty intervals.

---

## Sprint 19: Programming Languages Research (Projects 75-78)

*Creating new ways to express computation.*

### 75 — Dependent Types
**What:** Add dependent types to a simple language.  
**Why:** Prove properties of programs at compile time.  
**Implementation:** Type checking with dependent types, theorem proving.  
**What you'll learn:** Type theory, Curry-Howard correspondence, proof assistants.

### 76 — Formal Verification
**What:** Formally verify a small program using Hoare logic or separation logic.  
**Why:** Prove correctness, not just test for it.  
**Implementation:** Pre/post conditions, invariants, proof automation.  
**What you'll learn:** Program verification, automated theorem proving.

### 77 — Effect Systems
**What:** Implement effect tracking in a language (like Koka's effect system).  
**Why:** Track side effects in type system.  
**Implementation:** Effect inference, effect handlers, algebraic effects.  
**What you'll learn:** Effect typing, algebraic effects, handlers.

### 78 — Language from Idea
**What:** Design and implement a language for a novel paradigm (e.g., probabilistic programming, differentiable programming).  
**Why:** New problems need new languages.  
**Implementation:** Full compiler stack, standard library, examples.  
**What you'll learn:** Full-stack language design, novel semantics.

---

## Sprint 20: Operating Systems Research (Projects 79-81)

*Rethinking the foundation.*

### 79 — Unikernel
**What:** Build a unikernel — application specialized to run directly on hypervisor.  
**Why:** Minimal attack surface, fast boot, high performance.  
**Implementation:** Single address space, minimal libc, hypervisor interface.  
**What you'll learn:** OS specialization, minimalism, fast boot.

### 80 — Capability-based OS
**What:** Implement capability-based security in a small kernel.  
**Why:** Capabilities are more secure than ACLs.  
**Implementation:** Capability passing, revocation, amplification.  
**What you'll learn:** Capability security, object capabilities, confinement.

### 81 — Persistent Memory System
**What:** Build a system that treats persistent memory as first-class.  
**Why:** New hardware (Optane) changes the memory/storage hierarchy.  
**Implementation:** DAX, persistent data structures, crash consistency.  
**What you'll learn:** Persistent memory programming, crash consistency.

---

## Sprint 21: Cryptography Research (Projects 82-85)

*Security for the next century.*

### 82 — Post-Quantum Cryptography
**What:** Implement a post-quantum algorithm (e.g., Kyber, Dilithium).  
**Why:** Quantum computers break current crypto.  
**Implementation:** Lattice-based cryptography, module learning with errors.  
**What you'll learn:** Lattice crypto, NIST PQC standards.

### 83 — Fully Homomorphic Encryption
**What:** Implement a simple FHE scheme (like BFV or CKKS).  
**Why:** Compute on encrypted data without decrypting.  
**Implementation:** Learning with errors, relinearization, bootstrapping.  
**What you'll learn:** Homomorphic operations, noise management, bootstrapping.

### 84 — MPC at Scale
**What:** Build a multi-party computation system for many parties.  
**Why:** Multiple parties compute jointly without revealing inputs.  
**Implementation:** Secret sharing, garbled circuits, oblivious transfer.  
**What you'll learn:** Secure computation, malicious security, efficiency.

### 85 — Zero-Knowledge Virtual Machine
**What:** Implement a VM that proves correct execution with ZK proofs.  
**Why:** Verify computation without re-executing.  
**Implementation:** zk-SNARKs/zk-STARKs, arithmetic circuits, proof generation.  
**What you'll learn:** ZK proofs, verifiable computation, zkVM architecture.

---

## Sprint 22: AI/ML Infrastructure (Projects 86-89)

*The systems that run AI.*

### 86 — ML Compiler
**What:** Build a compiler that optimizes ML models (like TVM or XLA).  
**Why:** ML models need hardware-specific optimization.  
**Implementation:** Graph optimization, operator fusion, code generation.  
**What you'll learn:** ML graph optimization, code generation for accelerators.

### 87 — Distributed Training System
**What:** Implement distributed training with data/model parallelism.  
**Why:** Large models don't fit on one GPU.  
**Implementation:** All-reduce, parameter servers, gradient synchronization.  
**What you'll learn:** Distributed ML, synchronization, scaling efficiency.

### 88 — Model Optimizer
**What:** Build a system that compresses models (quantization, pruning, distillation).  
**Why:** Deploy models on edge devices.  
**Implementation:** Quantization-aware training, pruning algorithms, distillation.  
**What you'll learn:** Model compression, deployment constraints.

### 89 — AI Scheduler
**What:** Build a scheduler for ML workloads on clusters.  
**Why:** ML training has different patterns than web services.  
**Implementation:** Gang scheduling, preemption, elasticity.  
**What you'll learn:** ML workload characteristics, specialized scheduling.

---

## Sprint 23: Planetary Scale (Projects 90-92)

*Systems that span the globe — and beyond.*

### 90 — Global Consensus
**What:** Implement consensus across multiple continents.  
**Why:** True global systems require overcoming speed of light.  
**Implementation:** Hierarchical consensus, geographic sharding, latency modeling.  
**What you'll learn:** Geo-distributed systems, speed of light limits.

### 91 — Interplanetary Networking (DTN)
**What:** Implement Delay/Disruption Tolerant Networking protocols.  
**Why:** Space communication has minutes of delay.  
**Implementation:** Bundle protocol, custody transfer, store-and-forward.  
**What you'll learn:** DTN, extreme latency, lossy links.

### 92 — Space Communication Protocols
**What:** Design protocols for satellite constellations (like Starlink).  
**Why:** Low-earth orbit satellites are the new backbone.  
**Implementation:** Handover, routing in moving networks, laser links.  
**What you'll learn:** Mobile networks, orbital dynamics, laser communication.

---

## Sprint 24: Your Own Research (Projects 93-97)

*Now you create what doesn't exist.*

### 93 — CRDT Research
**What:** Invent a new CRDT for a data type that doesn't have one.  
**Why:** Existing CRDTs don't cover all use cases.  
**Implementation:** Mathematical proof, implementation, evaluation.  
**What you'll learn:** Research methodology, publication.

### 94 — Novel Storage Engine
**What:** Design a storage engine for a new hardware trend.  
**Why:** Existing engines optimized for old hardware.  
**Implementation:** New data structures, benchmarks, analysis.  
**What you'll learn:** Storage research, hardware trends.

### 95 — New Consensus Protocol
**What:** Design a consensus protocol optimized for your domain (IoT, edge, etc.).  
**Why:** Raft/Paxos make assumptions that may not hold.  
**Implementation:** Protocol spec, implementation, evaluation.  
**What you'll learn:** Protocol design, trade-off analysis.

### 96 — Language Design
**What:** Design a language for a new paradigm (quantum, biological, etc.).  
**Why:** Existing languages don't capture the paradigm well.  
**Implementation:** Grammar, compiler, examples.  
**What you'll learn:** Language design for novel domains.

### 97 — OS Concept
**What:** Design an OS for a new hardware architecture (quantum, neuromorphic, etc.).  
**Why:** Traditional OS assumptions don't hold.  
**Implementation:** Kernel prototype, drivers, benchmarks.  
**What you'll learn:** OS design for novel hardware.

---

## Sprint 25: Open Source Mastery (Projects 98-101)

*Give back to the tools you used.*

### 98 — PostgreSQL Core Contribution
**What:** Submit a significant patch to PostgreSQL core.  
**Why:** Be part of the database you've used for years.  
**Implementation:** Pick an issue, design, implement, shepherding.  
**What you'll learn:** Large codebase navigation, community process.

### 99 — Linux Kernel Contribution
**What:** Submit a driver or feature to Linux kernel.  
**Why:** The kernel runs the world.  
**Implementation:** Kernel development process, mailing lists, reviews.  
**What you'll learn:** Kernel internals, upstream process.

### 100 — LLVM Contribution
**What:** Add an optimization or feature to LLVM.  
**Why:** LLVM powers most modern compilers.  
**Implementation:** LLVM internals, optimization passes, testing.  
**What you'll learn:** Compiler infrastructure, code generation.

### 101 — Kafka KIP Lead
**What:** Lead a Kafka Improvement Proposal (KIP) to completion.  
**Why:** Shape the future of event streaming.  
**Implementation:** Proposal, community consensus, implementation, release.  
**What you'll learn:** Open source leadership, consensus building.

---

## Sprint 26: Ultimate Understanding (Projects 102-106)

*Build it from scratch. Everything.*

### 102 — GDB from Scratch
**What:** Build a debugger that can set breakpoints and inspect memory.  
**Why:** Understand how debuggers see into running programs.  
**Implementation:** ptrace, ELF parsing, DWARF debug info.  
**What you'll learn:** Debugger internals, binary formats, debugging info.

### 103 — Linux from Scratch
**What:** Build your own Linux distribution from source.  
**Why:** Understand every piece of your OS.  
**Implementation:** Kernel compilation, bootloader, init system, package management.  
**What you'll learn:** Full OS stack, boot process, system integration.

### 104 — Compiler from Scratch
**What:** Write a compiler for a real language (C subset) that generates working code.  
**Why:** Understand every phase of compilation.  
**Implementation:** Lexer, parser, semantic analysis, code generation.  
**What you'll learn:** Full compiler pipeline, assembly, linking.

### 105 — Database from Scratch
**What:** Build a production-worthy database (not just toy).  
**Why:** Understand every layer of data systems.  
**Implementation:** Storage, indexing, query processing, transactions, replication.  
**What you'll learn:** Full database architecture, production concerns.

### 106 — OS from Scratch
**What:** Build a working operating system (boot to userspace).  
**Why:** Ultimate understanding of computing.  
**Implementation:** Bootloader, memory management, processes, filesystem, drivers.  
**What you'll learn:** Full OS architecture, hardware interaction.

---

## Sprint 27: Your Own Invention (Project 107)

*The one that defines your career.*

### 107 — Your Original Contribution
**What:** Invent something new. Not implement something that exists — create what doesn't.  
**Why:** This is the peak — becoming someone who changes the field.  
**Implementation:** Research, prototype, evaluation, publication, community.  
**What you'll learn:** Invention, contribution to human knowledge.

**Ideas (but yours will be better):**
- A new programming paradigm
- A fundamentally new storage engine
- A consensus algorithm for a new domain
- A hardware/software co-design for a critical problem
- A system that enables something previously impossible

---

## How Phase 3 Differs from Previous Phases

| Aspect | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|
| **Focus** | Using | Creating | Inventing |
| **Code** | Microservices | Systems | Research prototypes |
| **Depth** | Applications | Components | First principles |
| **Languages** | Go, Node.js | Go, Rust, C | Anything needed |
| **Mindset** | "How do I use?" | "How do I build?" | "What should exist?" |
| **Outcome** | Senior Engineer | Architect | Distinguished Engineer |

---

## Prerequisites for Phase 3

Before starting Phase 3, you should have:

- Completed Phase 1 and Phase 2 (5+ years of focused work)
- Deep expertise in multiple areas
- Ability to read and understand academic papers
- Willingness to learn completely new fields
- Comfort with being a beginner again
- **Time and space to think** — this isn't about shipping features

---

## Infrastructure for Phase 3

```bash
phase-3/infrastructure/
├── docker/                 # Still useful
├── qemu/                   # For OS development
├── fpga/                   # FPGA toolchains
├── quantum/                # Quantum simulators
├── cluster/                # For distributed research
└── papers/                 # PDFs of academic papers
```

---

## What You'll Become

After Phase 3, you are no longer just an engineer or architect.

**You can:**
- Read academic papers and implement them
- Identify problems that don't have solutions yet
- Invent new algorithms, protocols, and systems
- Contribute to the core of major open source projects
- Teach the next generation of engineers
- Change how technology evolves

**You become the person whose work Phase 1 engineers use and Phase 2 architects study.**

---

## The Real Truth About Phase 3

Most people never get here. And that's okay — the world needs great senior engineers and architects.

Phase 3 is for those who:
- Can't stop asking "why?"
- Need to understand first principles
- Want to leave their mark on the field
- Are willing to spend years on a single problem
- Find joy in the journey, not just the destination

**If Phase 1 makes you a professional, and Phase 2 makes you a master, Phase 3 makes you a pioneer.**

---

## Final Note

Projects 61-107 aren't a checklist. They're a **landscape** — showing what's possible when you've mastered the fundamentals and start exploring the frontiers.

Some of these projects might take years. Some might lead to dead ends. Some might change your career direction entirely.

**That's the point.**

---

**⭐ Phase 1: Senior. Phase 2: Creator. Phase 3: Pioneer. The journey never ends. ⭐**
