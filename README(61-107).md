# Phase 3: Research & Specialization (Projects 61-107)

[![Rust](https://img.shields.io/badge/Rust-2021+-DEA584?style=flat-square&logo=rust)](https://www.rust-lang.org/)
[![C++](https://img.shields.io/badge/C++-20-00599C?style=flat-square&logo=cplusplus)](https://isocpp.org/)
[![Assembly](https://img.shields.io/badge/Assembly-x86_64-525252?style=flat-square&logo=assemblyscript)](https://en.wikipedia.org/wiki/Assembly_language)
[![CUDA](https://img.shields.io/badge/CUDA-12-76B900?style=flat-square&logo=nvidia)](https://developer.nvidia.com/cuda-toolkit)
[![FPGA](https://img.shields.io/badge/FPGA-Verilog-5C2D91?style=flat-square&logo=amd)](https://en.wikipedia.org/wiki/Field-programmable_gate_array)
[![Quantum](https://img.shields.io/badge/Quantum-Qiskit-6929C4?style=flat-square&logo=ibm)](https://qiskit.org/)

*A map of research directions for testing ideas, reproducing results, and exploring unfamiliar systems.*

---

## The Mindset Shift

Phase 1 investigates **application behavior**.  
Phase 2 investigates **internal mechanisms**.  
Phase 3 offers directions for **focused research and specialization**.

Choose a question, study existing work, state your assumptions, and run an experiment that could disprove your hypothesis. Reproducing a known result or finding the limits of a design is a useful outcome.

**These projects form an optional research map. They do not assign a career level or require mastery of every field.**

---

## Areas to Explore in This Phase

| Area | Skills |
|------|--------|
| **Hardware/Software Co-design** | FPGA programming, Verilog, RTL, hardware accelerators |
| **Quantum Computing** | Quantum algorithms, Qiskit, quantum cryptography |
| **Bioinformatics** | Genomic data processing, protein folding, DNA storage |
| **Distributed Systems Theory** | Consensus assumptions and variations, Byzantine fault tolerance |
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
├── 98-postgres-core/         # Sprint 25: Open Source Contribution
├── 99-linux-kernel/
├── 100-llvm-contribution/
├── 101-kafka-kip-lead/
│
├── 102-gdb-from-scratch/     # Sprint 26: Systems Integration
├── 103-linux-from-scratch/
├── 104-compiler-from-scratch/
├── 105-database-from-scratch/
├── 106-os-from-scratch/
└── 107-your-own-invention/
```

---

## Sprint 15: Hardware/Software Co-design (Projects 61-64)

*Measure when specialized hardware improves the full workload.*

### 61 — FPGA Accelerator
**What:** Implement a hardware accelerator for a specific algorithm (e.g., compression, encryption) on FPGA.  
**Why:** Compare a specialized accelerator with a software baseline, including data transfer, latency, throughput, and resource cost.  
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
**Why:** FHE has substantial computational costs; profile a chosen operation and investigate whether hardware acceleration improves the full workload.  
**Implementation:** Specialized arithmetic circuits, modular multiplication.  
**What you'll learn:** Hardware acceleration for cryptography, side-channel resistance.

---

## Sprint 16: Quantum Computing (Projects 65-67)

*Explore quantum computation with explicit assumptions and classical baselines.*

### 65 — Quantum Algorithms
**What:** Implement Shor's algorithm or Grover's search on quantum simulators.  
**Why:** Study how quantum algorithms change the complexity of selected problems and distinguish this from changing what is computable.  
**Implementation:** Qiskit, quantum circuits, superposition, entanglement.  
**What you'll learn:** Quantum gates, amplitude amplification, quantum Fourier transform.

### 66 — Quantum Cryptography
**What:** Simulate BB84 quantum key distribution with noise and an explicit attacker model.  
**Why:** BB84 illustrates how measurements can reveal eavesdropping under defined assumptions; authentication, implementation flaws, and side channels remain part of the security model.  
**Implementation:** Qiskit, quantum states, measurement, eavesdropping detection.  
**What you'll learn:** Quantum key distribution, no-cloning theorem, eavesdropping detection.

### 67 — Quantum Machine Learning
**What:** Build a quantum neural network for simple classification.  
**Why:** Compare a small quantum model with a classical baseline, accounting for data encoding, training cost, noise, and accuracy.  
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
**Why:** DNA offers high potential storage density, with encoding constraints, synthesis cost, and read errors that can be modeled experimentally.  
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
**What:** Specify and evaluate a variation of an existing consensus protocol for a defined use case.  
**Why:** Existing protocols make trade-offs you might not need.  
**Implementation:** Explicit assumptions, safety and liveness properties, comparison with an existing protocol; document counterexamples as findings.  
**What you'll learn:** Research methodology, protocol design, evaluation.

### 73 — Distributed Operating System
**What:** Prototype a bounded OS service, such as process management, that presents a cluster through a single-system interface.  
**Why:** Investigate where location transparency helps and where network delays and partial failures must remain visible.  
**Implementation:** Single system image, distributed process management.  
**What you'll learn:** OS design at cluster scale, location transparency.

### 74 — Global Clock Service
**What:** Model a clock service with explicit uncertainty intervals, inspired by TrueTime.  
**Why:** Bounded clock uncertainty can support externally consistent transactions, while other consistency protocols use logical ordering without synchronized physical clocks.  
**Implementation:** Simulated clock drift and synchronization errors, uncertainty intervals, and a transaction-ordering experiment; specialized clock hardware is optional.  
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
**Why:** Prove stated properties against an explicit specification and assumptions, then test the implementation against the model.  
**Implementation:** Pre/post conditions, invariants, proof automation.  
**What you'll learn:** Program verification, automated theorem proving.

### 77 — Effect Systems
**What:** Implement effect tracking in a language (like Koka's effect system).  
**Why:** Track side effects in type system.  
**Implementation:** Effect inference, effect handlers, algebraic effects.  
**What you'll learn:** Effect typing, algebraic effects, handlers.

### 78 — Language from Idea
**What:** Design a small language for a selected paradigm, such as probabilistic or differentiable programming, and compare it with existing approaches.  
**Why:** Explore when dedicated syntax and semantics improve expression of a problem compared with a library.  
**Implementation:** A bounded language specification, compiler or interpreter, small library, and comparison examples.  
**What you'll learn:** Language semantics, implementation choices, and comparison with existing approaches.

---

## Sprint 20: Operating Systems Research (Projects 79-81)

*Rethinking the foundation.*

### 79 — Unikernel
**What:** Build a unikernel — application specialized to run directly on hypervisor.  
**Why:** Measure the effect of a specialized runtime on attack surface, boot time, memory, performance, and observability.  
**Implementation:** Single address space, minimal libc, hypervisor interface.  
**What you'll learn:** OS specialization, minimalism, fast boot.

### 80 — Capability-based OS
**What:** Implement capability-based security in a small kernel.  
**Why:** Capabilities make authority explicit and delegable; compare confinement and revocation with an ACL-based design under the same threat model.  
**Implementation:** Capability passing, revocation, amplification.  
**What you'll learn:** Capability security, object capabilities, confinement.

### 81 — Persistent Memory System
**What:** Build a system that treats persistent memory as first-class.  
**Why:** Persistent memory exposes the difference between a memory write becoming visible and becoming durable; these semantics can be explored in a simulator.  
**Implementation:** Persistent data structures and simulated crashes; use DAX or suitable hardware when available, with explicit persistence-ordering assumptions.  
**What you'll learn:** Persistent memory programming, crash consistency.

---

## Sprint 21: Cryptography Research (Projects 82-85)

*Study cryptographic mechanisms, assumptions, and implementation costs.*

### 82 — Post-Quantum Cryptography
**What:** Study and benchmark standardized post-quantum algorithms such as ML-KEM and ML-DSA using established implementations; implement a bounded component for learning.  
**Why:** A sufficiently capable quantum computer would threaten schemes such as RSA and elliptic-curve cryptography; explore migration costs and compatibility with post-quantum alternatives.  
**Implementation:** Lattice-based cryptography, module learning with errors.  
**What you'll learn:** Lattice crypto, NIST PQC standards.

### 83 — Fully Homomorphic Encryption
**What:** Implement a simple FHE scheme (like BFV or CKKS).  
**Why:** Compute on encrypted data without decrypting.  
**Implementation:** Learning with errors, relinearization, bootstrapping.  
**What you'll learn:** Homomorphic operations, noise management, bootstrapping.

### 84 — MPC at Scale
**What:** Build a multi-party computation system for many parties.  
**Why:** Parties can compute an agreed result while protecting inputs under a specified threat model; account for information disclosed by the result itself.  
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
**Why:** Geographic distance imposes communication delays; measure the trade-offs between latency, availability, and the chosen consistency guarantee.  
**Implementation:** Hierarchical consensus, geographic sharding, latency modeling.  
**What you'll learn:** Geo-distributed systems, speed of light limits.

### 91 — Interplanetary Networking (DTN)
**What:** Implement Delay/Disruption Tolerant Networking protocols.  
**Why:** Space communication has minutes of delay.  
**Implementation:** Bundle protocol, custody transfer, store-and-forward.  
**What you'll learn:** DTN, extreme latency, lossy links.

### 92 — Space Communication Protocols
**What:** Design protocols for satellite constellations (like Starlink).  
**Why:** Moving satellite networks offer a concrete setting for studying changing topology, intermittent links, and handover.  
**Implementation:** Handover, routing in moving networks, laser links.  
**What you'll learn:** Mobile networks, orbital dynamics, laser communication.

---

## Sprint 24: Your Own Research (Projects 93-97)

*Form a focused hypothesis, compare it with prior work, and evaluate the result.*

### 93 — CRDT Research
**What:** Investigate a CRDT design for a selected data type, starting from existing constructions and a stated invariant.  
**Why:** Existing CRDTs don't cover all use cases.  
**Implementation:** Mathematical proof, implementation, evaluation.  
**What you'll learn:** Research methodology, convergence proofs, evaluation, and writing a reproducible research note.

### 94 — Novel Storage Engine
**What:** Investigate a storage-engine design for a defined workload and hardware model.  
**Why:** Different workloads and hardware can change the trade-offs between layouts, indexes, and background maintenance.  
**Implementation:** New data structures, benchmarks, analysis.  
**What you'll learn:** Storage research, hardware trends.

### 95 — New Consensus Protocol
**What:** Continue project 72 by evaluating a protocol variation for a domain such as IoT or edge computing.  
**Why:** Raft/Paxos make assumptions that may not hold.  
**Implementation:** Protocol spec, implementation, evaluation.  
**What you'll learn:** Protocol design, trade-off analysis.

### 96 — Language Design
**What:** Continue project 78 with a focused language-design question in a selected domain.  
**Why:** Compare how alternative language designs express a concrete task and which errors they prevent.  
**Implementation:** Grammar, compiler, examples.  
**What you'll learn:** Language design for novel domains.

### 97 — OS Concept
**What:** Prototype one OS mechanism for a selected hardware model, such as a specialized accelerator.  
**Why:** Identify an OS assumption affected by the selected hardware and evaluate a specific alternative.  
**Implementation:** Kernel prototype, drivers, benchmarks.  
**What you'll learn:** OS design for novel hardware.

---

## Sprint 25: Open Source Contribution (Projects 98-101)

*Give back to the tools you used.*

### 98 — PostgreSQL Core Contribution
**What:** Prepare a focused PostgreSQL patch with a reproducible issue, tests, and a rationale, then submit it for review.  
**Why:** Be part of the database you've used for years.  
**Implementation:** Pick an issue, design, implement, shepherding.  
**What you'll learn:** Large codebase navigation, community process.

### 99 — Linux Kernel Contribution
**What:** Prepare a bounded Linux kernel fix, test, or feature and participate in its review.  
**Why:** Explore the design and review practices of a widely used kernel.  
**Implementation:** Kernel development process, mailing lists, reviews.  
**What you'll learn:** Kernel internals, upstream process.

### 100 — LLVM Contribution
**What:** Add an optimization or feature to LLVM.  
**Why:** LLVM provides a widely used compiler infrastructure with established testing and review practices.  
**Implementation:** LLVM internals, optimization passes, testing.  
**What you'll learn:** Compiler infrastructure, code generation.

### 101 — Kafka Improvement Proposal
**What:** Develop a Kafka improvement proposal supported by a concrete problem, alternatives, and a prototype where useful.  
**Why:** Practice technical proposal writing and evaluating compatibility, operating costs, and community feedback.  
**Implementation:** Proposal draft, prototype or measurements, and community review; record the feedback and resulting revisions.  
**What you'll learn:** Open source design discussions, proposal revision, and consensus building; acceptance and release depend on the community.

---

## Sprint 26: Integrating Systems Knowledge (Projects 102-106)

*Connect mechanisms by building bounded, working systems.*

### 102 — GDB from Scratch
**What:** Build a debugger that can set breakpoints and inspect memory.  
**Why:** Understand how debuggers see into running programs.  
**Implementation:** ptrace, ELF parsing, DWARF debug info.  
**What you'll learn:** Debugger internals, binary formats, debugging info.

### 103 — Linux from Scratch
**What:** Build your own Linux distribution from source.  
**Why:** Trace how a selected Linux system boots and connects its kernel, userspace, and build dependencies.  
**Implementation:** Kernel compilation, bootloader, init system, package management.  
**What you'll learn:** Boot process, build dependencies, and kernel/userspace integration.

### 104 — Compiler from Scratch
**What:** Write a compiler for a real language (C subset) that generates working code.  
**Why:** Understand every phase of compilation.  
**Implementation:** Lexer, parser, semantic analysis, code generation.  
**What you'll learn:** Full compiler pipeline, assembly, linking.

### 105 — Database from Scratch
**What:** Build an experimental database with a defined API, workload, durability model, and failure tests.  
**Why:** Connect storage, query execution, transactions, and recovery, and measure where the design stops meeting its stated requirements.  
**Implementation:** Storage, indexing, query processing, transactions, replication.  
**What you'll learn:** Database architecture, crash recovery, resource limits, and evidence needed to assess operational readiness.

### 106 — OS from Scratch
**What:** Build a working operating system (boot to userspace).  
**Why:** Trace how a small system boots, manages resources, and runs user programs.  
**Implementation:** Bootloader, memory management, processes, filesystem, drivers.  
**What you'll learn:** Core OS mechanisms, their integration, and hardware interaction.

---

## Sprint 27: Independent Investigation (Project 107)

*Choose a question that you want to investigate in depth.*

### 107 — Your Original Contribution
**What:** Investigate a focused question through prior work, a prototype, and a reproducible experiment.  
**Why:** Practice forming and testing an independent argument about a system or mechanism.  
**Implementation:** Research question, literature review, prototype, evaluation, and a research note describing results and limitations.  
**What you'll learn:** Independent investigation and communication of findings; a replication, counterexample, or negative result is a valid outcome.

**Possible directions:**

- A language feature that prevents a specific class of errors
- A storage layout evaluated against an existing baseline
- A protocol variation for a clearly defined failure model
- A hardware/software co-design for a critical problem
- An explanation or counterexample that clarifies the limits of an existing design

---

## How Phase 3 Differs from Previous Phases

| Aspect | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|
| **Focus** | Application experiments | Mechanism experiments | Research questions |
| **Code** | Microservices | Systems | Research prototypes |
| **Depth** | Applications | Components | First principles |
| **Languages** | Go, Node.js | Go, Rust, C | Anything needed |
| **Mindset** | "How do I use?" | "How do I build?" | "What should exist?" |
| **Learning outcome** | Explain application behavior | Explain internal mechanisms | Form and evaluate research hypotheses |

---

## Choosing a Phase 3 Project

Choose prerequisites for the specific experiment, rather than waiting to complete every earlier phase. A small verification model, GPU comparison, or source-code investigation can also support work in earlier projects.

For the chosen question, prepare:

- Enough background to explain the baseline mechanism and its assumptions
- A paper, specification, or established implementation to study
- A bounded prototype and a way to evaluate it
- Time to learn unfamiliar concepts and revise the experiment
- A record of results, limitations, and unanswered questions

No fixed number of years or completed projects is required. Adjust scope to your available time and deepen the branches that remain useful or interesting.

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

## Evidence of Learning

For a chosen direction, aim to demonstrate that you can:

- Read a relevant paper and explain its assumptions and contribution
- Reproduce a result or document why your experiment differs
- Compare a proposed change with a baseline under defined conditions
- Find counterexamples and separate measured results from conjecture
- Navigate a relevant codebase and prepare a reviewable change
- Explain your findings so another person can repeat the experiment

A contribution can be a prototype, test, bug report, comparison, or research note. Publication, upstream acceptance, and invention are possible outcomes, not completion requirements.

---

## Final Note

Projects 61-107 are a **landscape of optional directions**. Explore enough breadth to see connections, then choose where a deeper investigation is worthwhile.

Some experiments fit into a few sessions; others can grow into long-term work. Scope each one around a question, measurable evidence, and explicit limits. Revise or stop a branch when it has answered the question you brought to it.

The laboratory develops engineering judgment through experiments, failure analysis, and work with constraints. Operating real systems, taking responsibility for decisions, and collaborating with people add further experience to that foundation.

---

**Keep the questions meaningful, the experiments reproducible, and the conclusions proportional to the evidence.**
