## 📊 **Детальный анализ всех 30 проектов**

---

## **Блок I: Основы Concurrency (Single Instance)**

### **01-atomic-inventory-counter**
| | |
|---|---|
| **Реальная фича** | Списание товаров, бонусов, бронирование мест |
| **Где используется** | E-commerce (Ozon/WB), авиабилеты, банковские транзакции |
| **Best practices** | Optimistic/Pessimistic locks, transaction isolation levels, retry logic |
| **Интеграция** | Ядро любого сервиса с конкурентным доступом к ресурсу |

### **02-anti-bruteforce-vault**
| | |
|---|---|
| **Реальная фича** | Защита от перебора паролей, API ключей |
| **Где используется** | Auth-сервисы (Google/Auth0), платежные шлюзы |
| **Best practices** | Progressive delay, leaky bucket, Redis storage |
| **Интеграция** | Middleware для login/register endpoints |

### **03-heavy-task-worker**
| | |
|---|---|
| **Реальная фича** | Обработка видео, генерация отчетов, рассылка email |
| **Где используется** | YouTube (обработка видео), CRM (рассылки) |
| **Best practices** | Worker pools, backpressure, graceful shutdown |
| **Интеграция** | Отдельный worker-сервис + очередь задач |

### **04-idempotency-key-provider**
| | |
|---|---|
| **Реальная фича** | Защита от повторных платежей, дублей заказов |
| **Где используется** | Stripe, PayPal, любые платежные системы |
| **Best practices** | Exactly-once delivery, idempotency keys, Redis TTL |
| **Интеграция** | Middleware для POST-эндпоинтов |

---

## **Блок II: Распределенные системы**

### **05-distributed-rate-limiter**
| | |
|---|---|
| **Реальная фича** | Ограничение запросов от одного IP/пользователя |
| **Где используется** | Twitter API, GitHub API, любые публичные API |
| **Best practices** | Sliding window, Redis Lua, cluster sync |
| **Интеграция** | API Gateway / отдельный rate-limiter сервис |

### **06-multilayer-cache**
| | |
|---|---|
| **Реальная фича** | Кэширование данных с защитой от cache stampede |
| **Где используется** | Netflix (рекомендации), любые read-heavy сервисы |
| **Best practices** | L1/L2 cache, single flight, write-through |
| **Интеграция** | Кэширующий слой перед БД |

### **07-secure-bff**
| | |
|---|---|
| **Реальная фича** | Агрегация API для фронтенда с безопасностью |
| **Где используется** | Микросервисные архитектуры, мобильные приложения |
| **Best practices** | JWT validation, API composition, cookies |
| **Интеграция** | Входная точка для фронтенда |

### **08-api-gateway-proxy**
| | |
|---|---|
| **Реальная фича** | Единый вход для всех микросервисов |
| **Где используется** | Kubernetes (Ingress), Netflix Zuul, Kong |
| **Best practices** | Load balancing, circuit breaking, routing |
| **Интеграция** | Центральный шлюз для всей системы |

---

## **Блок III: Data Engineering**

### **09-terabyte-data-mocker**
| | |
|---|---|
| **Реальная фича** | Генерация тестовых данных для нагрузочного тестирования |
| **Где используется** | Тестирование БД перед продакшеном |
| **Best practices** | Bulk insert, batch processing, index optimization |
| **Интеграция** | CI/CD пайплайн, тестовые окружения |

### **10-readwrite-splitter**
| | |
|---|---|
| **Реальная фича** | Автоматическое разделение чтения/записи |
| **Где используется** | Instagram, любые системы с master-slave репликацией |
| **Best practices** | Replication lag handling, connection pooling |
| **Интеграция** | Слой доступа к данным (DAO/repository) |

### **11-custom-database-sharder**
| | |
|---|---|
| **Реальная фича** | Горизонтальное масштабирование БД |
| **Где используется** | Discord, Uber, любые hyper-growth проекты |
| **Best practices** | Consistent hashing, rebalancing, virtual nodes |
| **Интеграция** | Прокси-слой перед шардированными БД |

### **12-saas-multitenancy**
| | |
|---|---|
| **Реальная фича** | Изоляция данных разных клиентов |
| **Где используется** | Salesforce, Slack, любые B2B SaaS |
| **Best practices** | Row-level security, schema per tenant, connection pooling |
| **Интеграция** | База данных для SaaS-продукта |

---

## **Блок IV: Real-time & Complex**

### **13-highload-chat-engine**
| | |
|---|---|
| **Реальная фича** | Real-time сообщения для миллионов пользователей |
| **Где используется** | Discord, Telegram, WhatsApp |
| **Best practices** | WebSocket scaling, Pub/Sub, presence |
| **Интеграция** | Отдельный chat-сервис |

### **14-real-time-leaderboard**
| | |
|---|---|
| **Реальная фича** | Топ игроков в реальном времени |
| **Где используется** | Игры (PUBG, Fortnite), геймификация |
| **Best practices** | Redis sorted sets, in-memory aggregation |
| **Интеграция** | Геймификационный движок |

### **15-distributed-saga**
| | |
|---|---|
| **Реальная фича** | Распределенные транзакции между микросервисами |
| **Где используется** | Бронирование отелей (отель + авиабилет) |
| **Best practices** | Saga pattern (orchestration/choreography), outbox |
| **Интеграция** | Оркестратор для бизнес-процессов |

### **16-circuit-breaker**
| | |
|---|---|
| **Реальная фича** | Защита от каскадных отказов |
| **Где используется** | Netflix Hystrix, Amazon AWS |
| **Best practices** | Failure detection, half-open state, metrics |
| **Интеграция** | Обертка вокруг внешних вызовов |

---

## **Блок V: Observability**

### **17-dynamic-feature-toggle**
| | |
|---|---|
| **Реальная фича** | Включение фич без деплоя |
| **Где используется** | Facebook, Google, canary deployments |
| **Best practices** | Centralized config, gradual rollout, A/B testing |
| **Интеграция** | Config-сервис + SDK |

### **18-log-aggregator**
| | |
|---|---|
| **Реальная фича** | Централизованный сбор логов |
| **Где используется** | ELK Stack, Grafana Loki |
| **Best practices** | Structured logging, log streaming |
| **Интеграция** | Все сервисы пишут сюда |

### **19-metrics-exporter**
| | |
|---|---|
| **Реальная фича** | Сбор метрик для мониторинга |
| **Где используется** | Prometheus + Grafana в каждом проекте |
| **Best practices** | RED метод, histograms, exemplars |
| **Интеграция** | Встроен в каждый микросервис |

### **20-grand-dashboard**
| | |
|---|---|
| **Реальная фича** | Единый дашборд для всей системы |
| **Где используется** | Мониторинг продакшена |
| **Best practices** | SLA/SLO отслеживание, алерты |
| **Интеграция** | Верхнеуровневый мониторинг |

---

## **Блок VI: Event-Driven**

### **21-kafka-exactly-once**
| | |
|---|---|
| **Реальная фича** | Гарантированная обработка событий без дублей |
| **Где используется** | Финансовые системы, аудит |
| **Best practices** | Idempotent producer, transactional API |
| **Интеграция** | Продюсер/консьюмер для критичных данных |

### **22-event-sourcing**
| | |
|---|---|
| **Реальная фича** | Хранение истории всех изменений |
| **Где используется** | Банки, аудиторские системы |
| **Best practices** | Event store, snapshots, replay |
| **Интеграция** | Хранение транзакций в банке |

### **23-distributed-job-scheduler**
| | |
|---|---|
| **Реальная фича** | Планировщик задач без единой точки отказа |
| **Где используется** | Рассылки, бекапы, ретеншен данные |
| **Best practices** | Leader election, distributed locks |
| **Интеграция** | Отдельный scheduler-сервис |

### **24-change-data-capture**
| | |
|---|---|
| **Реальная фича** | Стриминг изменений из БД в другие системы |
| **Где используется** | Индексация в Elastic, обновление кэшей |
| **Best practices** | Debezium, Kafka Connect |
| **Интеграция** | Синхронизация между сервисами |

---

## **Блок VII: High Performance**

### **25-custom-tcp-proxy**
| | |
|---|---|
| **Реальная фича** | Балансировка на транспортном уровне |
| **Где используется** | Балансировщики нагрузки (HAProxy, Nginx) |
| **Best practices** | Zero-copy, event loop, buffer management |
| **Интеграция** | Входной слой для всего трафика |

### **26-zero-copy-file-server**
| | |
|---|---|
| **Реальная фича** | Отдача статики с минимальными затратами |
| **Где используется** | CDN, файловые хранилища (S3) |
| **Best practices** | sendfile, DMA, kernel bypass |
| **Интеграция** | Отдельный статик-сервер |

### **27-binary-protocol-parser**
| | |
|---|---|
| **Реальная фича** | Эффективная сериализация данных |
| **Где используется** | gRPC, Avro, Thrift |
| **Best practices** | Schema evolution, code generation |
| **Интеграция** | Внутренняя коммуникация сервисов |

---

## **Блок VIII: Security & Crypto**

### **28-distributed-lock-manager**
| | |
|---|---|
| **Реальная фича** | Блокировки между сервисами |
| **Где используется** | Координация распределенных задач |
| **Best practices** | Redlock algorithm, fencing tokens |
| **Интеграция** | Координация критических секций |

### **29-merkle-tree-validator**
| | |
|---|---|
| **Реальная фича** | Проверка целостности больших данных |
| **Где используется** | Блокчейн (Bitcoin), бэкапы |
| **Best practices** | Hash trees, consistency proofs |
| **Интеграция** | Валидация данных при синхронизации |

### **30-hot-cold-wallet**
| | |
|---|---|
| **Реальная фича** | Безопасное хранение криптоактивов |
| **Где используется** | Криптобиржи (Binance, Coinbase) |
| **Best practices** | Multi-sig, approval workflows, cold storage |
| **Интеграция** | Wallet-сервис для крипто |

---

## 🔥 **Как объединить в реальные сервисы**

### **Сервис 1: API Gateway** (проекты 5, 7, 8, 16)
- Rate limiting (5)
- BFF логика (7)
- Роутинг (8)
- Circuit breaker (16)

### **Сервис 2: Auth Service** (проект 2)
- Anti-bruteforce
- JWT management

### **Сервис 3: Business Core** (проекты 1, 4, 15)
- Inventory/бронирования
- Idempotency
- Saga orchestration

### **Сервис 4: Data Layer** (проекты 10, 11, 12)
- Read/write splitting
- Sharding
- Multitenancy

### **Сервис 5: Worker Service** (проекты 3, 21, 23)
- Тяжелые задачи
- Kafka consumers
- Job scheduling

### **Сервис 6: Real-time Service** (проекты 13, 14)
- WebSockets
- Leaderboards

### **Сервис 7: Storage Service** (проекты 6, 9, 24, 26)
- Caching
- File serving
- CDC

### **Сервис 8: Observability Stack** (проекты 17, 18, 19, 20)
- Feature toggles
- Logs
- Metrics
- Dashboard

### **Сервис 9: Crypto Service** (проекты 28, 29, 30)
- Distributed locks
- Merkle trees
- Wallet logic

### **Сервис 10: Network Proxy** (проекты 25, 27)
- TCP proxy
- Binary protocols
