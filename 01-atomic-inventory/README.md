# Task 01: Atomic Inventory Counter

## 1. Objective
Implement a fault-tolerant inventory deduction system that handles thousands of concurrent requests without race conditions or overselling.

## 2. Core Problem
When multiple threads or application instances access the same database row simultaneously, a race condition occurs:

```
Thread 1: Read stock = 5
Thread 2: Read stock = 5
Thread 1: Write stock = 4
Thread 2: Write stock = 4  ❌ One unit lost
```

**Critical requirement:** `stock_quantity` must never become negative, no matter how many concurrent requests hit the system.

## 3. Implementation Strategies
Implement and compare three approaches in both Go and Node.js:

### Strategy 1: Pessimistic Locking
Use row-level locks in the database.

```sql
BEGIN;
SELECT * FROM products WHERE id = $1 FOR UPDATE;
UPDATE products SET stock = stock - $2 WHERE id = $1 AND stock >= $2;
COMMIT;
```

### Strategy 2: Optimistic Locking
Use version field and retry on conflict.

```sql
UPDATE products 
SET stock = stock - $1, version = version + 1 
WHERE id = $2 AND stock >= $1 AND version = $3;
```

### Strategy 3: Redis Atomic Counter
Use Redis atomic operations with PostgreSQL persistence.

```lua
-- Redis Lua script
local current = redis.call('GET', KEYS[1])
if tonumber(current) >= tonumber(ARGV[1]) then
    redis.call('DECRBY', KEYS[1], ARGV[1])
    return 1
else
    return 0
end
```

## 4. Technical Requirements

### Stack
- Go 1.25+ / Node.js 24.14+
- PostgreSQL 18.2
- Redis 8.2
- Docker + Docker Compose

### Database Schema
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    version INT DEFAULT 0
);

CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    request_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO products (sku, stock_quantity) VALUES ('SKU-TEST-001', 1000);
```

### API Contract
```http
POST /api/v1/inventory/reserve
{
    "sku": "SKU-TEST-001",
    "quantity": 1,
    "requestId": "uuid-123"
}
```

**Responses:**
- `200 OK` - Reserved successfully
- `409 Conflict` - Insufficient stock
- `422 Unprocessable` - Invalid request

## 5. Testing Requirements

### Unit Tests
- Idempotency: same requestId doesn't create duplicate deduction
- Expired accruals excluded from balance
- Concurrent deductions don't cause negative balance
- Queue job idempotency

### Load Tests
- 100,000 requests against 1,000 items
- Measure RPS, p95/p99 latency, error rate
- Compare all three strategies

## 6. Deliverables

- [ ] Go implementation (all 3 strategies)
- [ ] Node.js implementation (all 3 strategies)
- [ ] Unit tests for critical scenarios
- [ ] Load test scripts (k6/wrk)
- [ ] Performance comparison in `/docs`
- [ ] Docker Compose for local development

## 7. Success Criteria
- Final stock = 0 AND successful_requests = 1000
- No negative stock in any test
- All tests passing
- Clear documentation of tradeoffs

---

# Задание 01: Atomic Inventory Counter

## 1. Цель
Реализовать отказоустойчивую систему списания остатков товара, способную корректно обрабатывать тысячи конкурентных запросов без возникновения состояния гонки и ухода в минус.

## 2. Основная проблема
При одновременном доступе нескольких потоков или инстансов приложения к одной строке в базе данных возникает состояние гонки:

```
Поток 1: Читает stock = 5
Поток 2: Читает stock = 5
Поток 1: Записывает stock = 4
Поток 2: Записывает stock = 4  ❌ Потеряна одна единица
```

**Критическое требование:** `stock_quantity` никогда не должен становиться отрицательным, независимо от количества параллельных запросов.

## 3. Стратегии реализации
Реализовать и сравнить три подхода на Go и Node.js:

### Стратегия 1: Пессимистичная блокировка
Использование блокировок строк на уровне базы данных.

```sql
BEGIN;
SELECT * FROM products WHERE id = $1 FOR UPDATE;
UPDATE products SET stock = stock - $2 WHERE id = $1 AND stock >= $2;
COMMIT;
```

### Стратегия 2: Оптимистичная блокировка
Использование поля версии и повторных попыток при конфликте.

```sql
UPDATE products 
SET stock = stock - $1, version = version + 1 
WHERE id = $2 AND stock >= $1 AND version = $3;
```

### Стратегия 3: Атомарный счетчик в Redis
Использование атомарных операций Redis с сохранением в PostgreSQL.

```lua
-- Redis Lua скрипт
local current = redis.call('GET', KEYS[1])
if tonumber(current) >= tonumber(ARGV[1]) then
    redis.call('DECRBY', KEYS[1], ARGV[1])
    return 1
else
    return 0
end
```

## 4. Технические требования

### Стек
- Go 1.21+ / Node.js 20+
- PostgreSQL 16
- Redis 7.2
- Docker + Docker Compose

### Схема базы данных
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    version INT DEFAULT 0
);

CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    request_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO products (sku, stock_quantity) VALUES ('SKU-TEST-001', 1000);
```

### API Контракт
```http
POST /api/v1/inventory/reserve
{
    "sku": "SKU-TEST-001",
    "quantity": 1,
    "requestId": "uuid-123"
}
```

**Ответы:**
- `200 OK` - Успешно зарезервировано
- `409 Conflict` - Недостаточно товара
- `422 Unprocessable` - Невалидный запрос

## 5. Требования к тестированию

### Модульные тесты
- Идемпотентность: одинаковый requestId не создает повторное списание
- Просроченные начисления исключены из баланса
- Конкурентные списания не приводят к отрицательному балансу
- Идемпотентность задач в очереди

### Нагрузочные тесты
- 100,000 запросов на 1,000 единиц товара
- Измерить RPS, p95/p99 задержку, процент ошибок
- Сравнить все три стратегии

## 6. Результаты

- [ ] Go реализация (все 3 стратегии)
- [ ] Node.js реализация (все 3 стратегии)
- [ ] Модульные тесты для критических сценариев
- [ ] Скрипты нагрузочного тестирования (k6/wrk)
- [ ] Сравнение производительности в `/docs`
- [ ] Docker Compose для локальной разработки

## 7. Критерии успеха
- Итоговый остаток = 0 И успешных_запросов = 1000
- Ни одного отрицательного остатка в любом тесте
- Все тесты проходят
- Понятная документация компромиссов
