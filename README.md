# 🔬 Моя Highload Лаборатория: Путь к Senior Backend Engineer

<div align="center">
  
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Node.js](https://img.shields.io/badge/Node.js-20+-339933?style=for-the-badge&logo=nodedotjs)](https://nodejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![Kafka](https://img.shields.io/badge/Kafka-3.5-231F20?style=for-the-badge&logo=apache-kafka)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)

**30 архитектурных паттернов | 8 блоков | Финальный дашборд**

</div>

## 🎯 Посыл и философия

Этот репозиторий — моя лаборатория по погружению в highload-архитектуры. Каждый модуль здесь — не просто "туториал", а полноценный микросервис, который я проектирую и реализую с нуля. 

**Зачем это всё?** 
Чтобы не просто знать паттерны по названиям, а понимать их внутреннее устройство, узкие места и особенности реализации под нагрузкой. 30 проектов — 30 вызовов, каждый из которых закрывает конкретную проблему, с которой я столкнусь в реальном продакшне.

**Подход:** Каждую задачу я реализую на двух языках — Go и Node.js. Это позволяет глубже понять, как конкурентность и архитектура языка влияют на производительность и выбор инструмента под конкретную задачу.

---

## 🗺 Дорожная карта развития

```mermaid
graph LR
    A[Блок I: Основы<br/>Single Instance] --> B[Блок II: Сеть<br/>Distributed System]
    B --> C[Блок III: Данные<br/>Data Engineering]
    C --> D[Блок IV: Real-time<br/>Сложные взаимодействия]
    D --> E[Блок V: Observability<br/>Мониторинг]
    E --> F[Блок VI: Event-Driven<br/>Kafka & Streaming]
    F --> G[Блок VII: Low Latency<br/>Сетевой стек]
    G --> H[Блок VIII: Security<br/>Crypto & Blockchain]
    
    style A fill:#e1f5fe,stroke:#01579b
    style B fill:#fff3e0,stroke:#e65100
    style C fill:#f3e5f5,stroke:#4a148c
    style D fill:#e8f5e8,stroke:#1b5e20
    style E fill:#ffebee,stroke:#b71c1c
    style F fill:#fff8e1,stroke:#ff6f00
    style G fill:#e0f2f1,stroke:#004d40
    style H fill:#fbe9e7,stroke:#bf360c
```

---

## 📦 Блок I: Основы Concurrency и целостности (Single Instance)

*Фундамент любого сервиса. Работа с потоками, блокировками и гарантиями в рамках одного инстанса.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **1** | [Atomic Inventory Counter](./01-atomic-inventory) | Распродажа: списание 1k товаров под 100k заявок без ухода в минус | `SELECT FOR UPDATE` `Optimistic Locking` `Pessimistic Locking` `Mutex` `Atomic` |
| **2** | [Anti-Bruteforce Vault](./02-anti-bruteforce) | Защита входа с прогрессивной задержкой при неверных попытках | `Leaky Bucket` `Rate Limiting` `In-memory storage` |
| **3** | [Heavy Task Worker Pool](./03-heavy-worker) | Очередь тяжелых задач с контролем потребления ресурсов | `Worker Pool` `Semaphore` `Channels` `Worker Threads` |
| **4** | [Idempotency Key Provider](./04-idempotency) | Middleware, гарантирующий ровно одно выполнение операции | `Idempotency-Key` `Exactly-Once` `Redis` |

---

## 🌐 Блок II: Распределенные системы и Сеть (Distributed)

*Как заставить множество серверов работать как единый организм.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **5** | [Distributed Rate Limiter](./05-rate-limiter) | Ограничение нагрузки на весь кластер через общее хранилище | `Fixed Window` `Sliding Window` `Redis Lua` |
| **6** | [Multilayer Cache](./06-multilayer-cache) | L1 (in-memory) + L2 (Redis) с защитой от Cache Stampede | `Cache Patterns` `Pub/Sub` `Single Flight` |
| **7** | [Secure BFF](./07-bff) | Прослойка для Mobile/Web с безопасным обменом токенов | `JWT` `Cookies` `API Composition` |
| **8** | [API Gateway Aggregator](./08-gateway) | Параллельный сбор данных из 5 микросервисов | `Reverse Proxy` `Load Balancing` `Partial Failures` |

---

## 📊 Блок III: Data Engineering под нагрузкой

*Когда данных становится слишком много для одной базы.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **9** | [Terabyte Data Mocker](./09-data-mocker) | Генератор миллионов строк и оптимизация вставки | `Bulk Insert` `pgx Copy Protocol` `Streams` `Index Tuning` |
| **10** | [Read/Write Splitter](./10-readwrite-splitter) | Прослойка, разделяющая потоки на Master и реплики | `Leader/Follower` `Replication Lag` `CQRS` |
| **11** | [Custom Database Sharder](./11-sharder) | Распределение данных по разным БД через Consistent Hashing | `Consistent Hashing` `Sharding` `Virtual Nodes` |
| **12** | [SaaS Multitenancy Isolation](./12-multitenancy) | Изоляция данных разных клиентов в одном кластере | `Row-Level Security` `Schema per Tenant` |

---

## ⚡ Блок IV: Сложные паттерны и Real-time

*Мгновенные ответы и целостность в распределенном мире.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **13** | [High-Load Chat Engine](./13-chat-engine) | 50k+ WebSocket-соединений с рассылкой через Pub/Sub | `WebSockets` `Redis Pub/Sub` `Goroutines` `Event Loop` |
| **14** | [Real-time Leaderboard](./14-leaderboard) | Топ игроков из миллиона записей в реальном времени | `Sorted Sets` `In-Memory` `Atomic Updates` |
| **15** | [Distributed SAGA Orchestrator](./15-saga) | Распределенная транзакция с механизмом компенсации | `SAGA Pattern` `Outbox` `Orchestration` |
| **16** | [Circuit Breaker Service](./16-circuit-breaker) | Предохранитель для защиты от каскадных отказов | `Circuit Breaker` `Retry` `Timeout` `Bulkhead` |

---

## 📈 Блок V: Observability и Финал (The Grand Dashboard)

*Как понять, что происходит в системе под нагрузкой.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **17** | [Dynamic Feature Toggle](./17-feature-toggle) | Управление фичами без перезагрузки инстансов | `Feature Flags` `ETCD/Redis` `Runtime Config` |
| **18** | [Log Aggregator (Mini ELK)](./18-log-aggregator) | Сбор и парсинг логов в реальном времени | `Log Streaming` `ClickHouse` `Fluentd` |
| **19** | [System Metrics Exporter](./19-metrics-exporter) | Прошивка всех сервисов метриками | `Prometheus` `Latency` `RPS` `Error Rate` |
| **20** | [The Grand Dashboard](./20-grand-dashboard) | Визуализация всех 19 сервисов под нагрузкой | `Grafana` `PromQL` `k6` `Benchmarking` |

---

## 📨 Блок VI: Message Brokers & Event-Driven (BigTech Standard)

*Асинхронное взаимодействие и потоковая обработка событий.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **21** | [Kafka Exactly-Once Delivery](./21-kafka-exactly-once) | Пайплайн с гарантией обработки без дублей | `Kafka` `Idempotent Producer` `Transactional API` |
| **22** | [Event Sourcing Engine](./22-event-sourcing) | Состояние вычисляется из истории событий | `Event Store` `Audit Log` `Snapshots` |
| **23** | [Distributed Job Scheduler](./23-job-scheduler) | Планировщик, гарантирующий запуск на одном инстансе | `Distributed Locks` `Cron` `Leader Election` |
| **24** | [Change Data Capture (CDC)](./24-cdc) | Стриминг изменений из Postgres в ElasticSearch | `Debezium` `Kafka Connect` `Real-time Sync` |

---

## ⚙️ Блок VII: High Performance & Networking (Low Latency)

*Когда накладные расходы HTTP и JSON становятся роскошью.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **25** | [Custom TCP/UDP Proxy](./25-tcp-proxy) | Балансировка трафика на транспортном уровне | `Raw Sockets` `L4 Load Balancing` `Go net` |
| **26** | [Zero-Copy File Server](./26-zero-copy) | Отдача файлов минуя буферы приложения | `sendfile` `Stream.pipe` `DMA` |
| **27** | [Binary Protocol Parser](./27-binary-protocol) | Замена JSON на Protobuf/MessagePack | `Protocol Buffers` `MessagePack` `Serialization` |

---

## 🔐 Блок VIII: Cryptography & Decentralization (Crypto/Security)

*Механизмы безопасности и распределенного консенсуса.*

| # | Проект | Суть | Технологии |
|:-:|--------|------|------------|
| **28** | [Distributed Lock Manager (Redlock)](./28-redlock) | Блокировки между независимыми сервисами | `Redlock` `Redis` `Distributed Consensus` |
| **29** | [Merkle Tree Validator](./29-merkle-tree) | Проверка целостности миллионов записей | `Hash Trees` `Blockchain` `Tamper-proof` |
| **30** | [Hot/Cold Wallet Logic](./30-wallet) | Архитектура разделения доступа к активам | `Multi-sig` `Withdrawal Queues` `Transaction Signing` |

---

## 🏗 Как устроен каждый проект

```
проект/
├── go/                   # Реализация на Go
│   ├── cmd/
│   ├── internal/
│   └── README.md         # Особенности Go-версии
├── node/                 # Реализация на Node.js
│   ├── src/
│   ├── tests/
│   └── README.md         # Особенности Node-версии
├── docs/                 # Схемы, диаграммы, результаты тестов
├── docker-compose.yml    # Локальный запуск зависимостей
├── Makefile              # Удобные команды
└── README.md             # Описание проекта
```

---

## 🚀 Инфраструктура (общая для всех проектов)

В корне репозитория подготовлена общая инфраструктура, чтобы не разворачивать сервисы для каждой задачи заново:

```bash
infrastructure/
├── docker-compose.yml    # Postgres, Redis, Kafka, Prometheus, Grafana, ClickHouse
├── prometheus/           # Конфигурация сбора метрик
├── grafana/              # Дашборды (включая The Grand Dashboard)
├── k6/                   # Сценарии нагрузочного тестирования
└── scripts/              # Утилиты для запуска бенчмарков
```

**Быстрый старт:**

```bash
# Поднять всю инфраструктуру
make infra-up

# Запустить конкретный проект (например, atomic-inventory на Go)
cd 01-atomic-inventory
make run-go

# Запустить нагрузочный тест
make load-test

# Посмотреть метрики в Grafana
open http://localhost:3000
```

---

## 📊 Финальный дашборд (Блок V, проект 20)

Когда все 30 проектов будут готовы, запускается единый сценарий:

1. **Benchmark-бот** генерирует нагрузку на все сервисы одновременно
2. **Prometheus** собирает метрики с каждого инстанса
3. **Grafana** отображает:

   - RPS по каждому сервису (Go vs Node.js)
   - Латентность (p95, p99) в сравнении
   - Потребление памяти и CPU
   - Количество ошибок и ретраев
   - Состояние очередей и блокировок
   - Тепловые карты конфликтов (для блокировок)

---

## 🎓 Чему я учусь

- **Concurrency:** Сравнение горутин и event loop, worker threads, атомарные операции
- **Databases:** Транзакции, уровни изоляции, блокировки, шардирование, репликация
- **Architecture:** CQRS, Event Sourcing, SAGA, Circuit Breaker, BFF, API Gateway
- **Message Brokers:** Kafka, RabbitMQ, гарантии доставки, consumer groups
- **Observability:** Метрики, логи, трассировка, профилирование под нагрузкой
- **Networking:** WebSockets, TCP/UDP, бинарные протоколы, zero-copy
- **Security:** JWT, RLS, мультиподпись, распределенные блокировки

---

## 🤝 Обратная связь

Это живой репозиторий, который растёт вместе с моим пониманием highload-систем. Если у вас есть идеи или замечания — буду рад обсудить в Issues или PR.

---

<div align="center">
  
**⭐ Если этот репозиторий поможет и вам — поставьте звезду! ⭐**

</div>
