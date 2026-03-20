# Как работает Anti-Bruteforce: пошаговое объяснение

Документ объясняет реализацию L1, L2 и L3 по шагам, включая основы Redis.

---

## Часть 1: Что такое Sliding Window (скользящее окно)

**Цель:** Ограничить количество попыток входа (например, 5 попыток за 60 секунд).

**Проблема Fixed Window (фиксированное окно):**
- Окно 00:00–01:00: 5 попыток
- Окно 01:00–02:00: ещё 5 попыток
- Атакующий может сделать 5 попыток в 00:59 и 5 в 01:01 → 10 попыток за 2 секунды

**Решение Sliding Window:**
- В любой момент считаем попытки за **последние** 60 секунд
- В 01:01:30 окно = 00:00:30 – 01:01:30
- Попытки из 00:59 уже «выпали» из окна

**Как хранить:** Список меток времени (timestamp) каждой попытки. Окно = «сейчас минус 60 сек».

---

## Часть 2: Redis — что нужно знать

### Redis = ключ-значение в памяти

```
SET user:123 "John"     → ключ "user:123", значение "John"
GET user:123           → "John"
DEL user:123           → удалить
```

### Sorted Set (ZSET) — упорядоченное множество

Структура: **ключ** → множество пар **(member, score)**. Элементы отсортированы по score.

```
ZADD rate:192.168.1.1  1700000001  "req_001"   # score=timestamp, member=уникальный id
ZADD rate:192.168.1.1  1700000005  "req_002"
ZADD rate:192.168.1.1  1700000010  "req_003"
```

**Команды для нашего кейса:**

| Команда | Что делает |
|---------|------------|
| `ZADD key score member` | Добавить элемент. Score = timestamp (секунды), member = уникальный id попытки |
| `ZREMRANGEBYSCORE key min max` | Удалить элементы с score от min до max (старые попытки) |
| `ZCOUNT key min max` | Посчитать элементы с score в диапазоне |
| `ZRANGE key 0 -1 WITHSCORES` | Получить все элементы с их score |

**Пример:**
```
Ключ: rate:192.168.1.1
Элементы: (1700000000, "a"), (1700000010, "b"), (1700000055, "c"), (1700000065, "d")
Сейчас: 1700000070, окно 60 сек → cutoff = 1700000010

ZREMRANGEBYSCORE rate:192.168.1.1 -inf 1700000009  → удаляем "a" (старше 60 сек)
ZCOUNT rate:192.168.1.1 1700000010 1700000070     → 3 (b, c, d в окне)
```

---

## Часть 3: L1 — Naive (пошагово)

**Хранилище:** `map[string][]time.Time` в памяти процесса. Ключ = IP, значение = список времён попыток.

### Шаги при каждом запросе:

```
Запрос: POST /login от IP 192.168.1.1
Лимит: 5 попыток за 60 секунд
```

1. **Блокировка:** `s.mu.Lock()` — только один поток может выполнять логику
2. **Текущее время:** `now = time.Now()`
3. **Граница окна:** `cutoff = now - 60 сек` (всё старше — вне окна)
4. **Взять попытки IP:** `times = data["192.168.1.1"]`
5. **Отфильтровать:** оставить только `t > cutoff`
6. **Проверка:** если `len(kept) >= 5` → **429 Too Many Requests**
7. **Добавить текущую попытку:** `kept = append(kept, now)`
8. **Сохранить:** `data["192.168.1.1"] = kept`
9. **Разблокировать:** `s.mu.Unlock()`
10. **Ответ:** 200 OK

**Проблемы:** Не масштабируется (2 инстанса = 2× лимит), блокировка под нагрузкой.

---

## Часть 4: L2 — Pessimistic (пошагово)

**Хранилище:** Redis. Ключ `rate:{ip}` = Sorted Set с timestamp как score.

**Идея:** Перед работой с данными IP — взять **блокировку** (lock). Только один запрос в момент времени может читать/писать данные этого IP.

### Шаги при каждом запросе:

```
Запрос: POST /resource/pessimistic от IP 192.168.1.1
```

#### Фаза 1: Взять блокировку

1. **Ключ блокировки:** `lock:rate:192.168.1.1`
2. **Попытка:** `SET lock:rate:192.168.1.1 <uuid> NX EX 1`
   - `NX` = установить только если ключа нет (если уже занят — ошибка)
   - `EX 1` = TTL 1 секунда (автоосвобождение при зависании)
3. **Результат:**
   - `OK` → блокировка получена, идём дальше
   - не OK → ждём 50ms, повторяем (макс 3 раза). Если не получилось → 500

#### Фаза 2: Работа с данными (под блокировкой)

4. **Удалить старые записи:**
   ```
   ZREMRANGEBYSCORE rate:192.168.1.1 -inf (cutoff-1)
   ```
   Удаляем попытки старше 60 секунд.

5. **Посчитать текущие:**
   ```
   ZCOUNT rate:192.168.1.1 cutoff now
   ```

6. **Проверка:** если `count >= 5` → **429**, иначе продолжаем

7. **Добавить попытку:**
   ```
   ZADD rate:192.168.1.1 <now_unix> <unique_id>
   ```
   Score = timestamp в секундах, member = уникальный id (например, nanosecond)

#### Фаза 3: Освободить блокировку

8. **DEL lock:rate:192.168.1.1** — освободить lock (в defer)

9. **Ответ:** 200 OK

**Схема потоков:**
```
Запрос A (IP 1.1.1.1)     Запрос B (IP 1.1.1.1)     Запрос C (IP 2.2.2.2)
        |                          |                          |
   SET lock NX → OK                 |                          |
        |                    SET lock NX → fail                 |
   ZREM, ZCOUNT, ZADD               |                    SET lock NX → OK
        |                    sleep 50ms                         |
   DEL lock                         |                    ZREM, ZCOUNT, ZADD
        |                    SET lock NX → OK                   |
        |                    ZREM, ZCOUNT, ZADD                 |
        |                    DEL lock                    DEL lock
```

A и B сериализуются (один ждёт другого). C идёт параллельно — другой IP, другой lock.

---

## Часть 5: L3 — Optimistic (как это работает)

**Идея:** Без блокировки. Читаем данные, проверяем, пишем. Если между чтением и записью кто-то изменил ключ — Redis отменит нашу запись, и мы **повторим** попытку.

### Механизм Redis: WATCH + MULTI + EXEC

- **WATCH key** — Redis запоминает версию ключа
- **MULTI** — начать транзакцию (команды буферизуются)
- **EXEC** — выполнить все команды из буфера
- **Важно:** если между WATCH и EXEC ключ изменился → EXEC возвращает `nil` (ничего не выполнилось)

### Шаги L3 (псевдокод):

```go
func Allow(ip string, limit int, windowSec int64) (bool, error) {
    key := "rate:" + ip
    for attempt := 0; attempt < maxRetries; attempt++ {
        // 1. Начать "наблюдение" за ключом
        pipe := client.Pipeline()
        pipe.Watch(ctx, key)
        
        // 2. Прочитать текущие попытки
        entries, _ := client.ZRangeWithScores(ctx, key, 0, -1).Result()
        
        // 3. Отфильтровать старые (client-side)
        now := time.Now().Unix()
        cutoff := now - windowSec
        var kept int
        for _, z := range entries {
            if z.Score >= float64(cutoff) {
                kept++
            }
        }
        
        // 4. Уже превышен лимит?
        if kept >= limit {
            pipe.Unwatch(ctx)
            return false, nil  // 429
        }
        
        // 5. Запланировать добавление (пока не выполнено!)
        pipe.Multi()
        pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: uniqueId()})
        
        // 6. Выполнить. Если ключ изменился — вернёт nil
        results, err := pipe.Exec(ctx)
        if err == redis.TxFailedErr || results == nil {
            // Конфликт! Кто-то изменил key между WATCH и EXEC
            continue  // retry
        }
        
        return true, nil  // 200 OK
    }
    return false, errTooManyRetries
}
```

### Визуально: два параллельных запроса

```
Запрос A (IP 1.1.1.1)              Запрос B (IP 1.1.1.1)
        |                                   |
   WATCH rate:1.1.1.1                       |
        |                            WATCH rate:1.1.1.1
   ZRANGE → 3 попытки                      |
        |                            ZRANGE → 3 попытки
   MULTI + ZADD + EXEC                     |
        |  → OK, 4 попытки                  |
        |                            MULTI + ZADD + EXEC
        |                              → nil! (key изменился)
        |                            RETRY: WATCH, ZRANGE...
        |                              → 4 попытки
        |                            MULTI + ZADD + EXEC
        |                              → OK, 5 попыток
```

A успешно с первого раза. B получил конфликт (A изменил ключ), сделал retry и успешно.

### Когда L3 плохо работает

Если 100 запросов от одного IP одновременно:
- Все делают WATCH → читают одни и те же 4 попытки
- Первый EXEC успешен → 5 попыток
- Остальные 99 получают nil → retry
- На retry снова конкуренция → много retry, нагрузка на Redis

Поэтому для bruteforce (один IP бьёт много раз) L3 не идеален. L4 (Lua) — один атомарный скрипт, без retry.

---

## L2 vs L3: в чём разница

| Аспект | L2 Pessimistic | L3 Optimistic |
|--------|----------------|---------------|
| **Синхронизация** | Блокировка (lock) — «никто не входит, пока я не выйду» | Без блокировки — «попробую, если не вышло — retry» |
| **Первый шаг** | SET lock NX — ждать, пока lock свободен | WATCH key — просто «наблюдаю» |
| **Поведение при конкуренции** | Запрос B **ждёт** освобождения lock (sleep 50ms, retry) | Запрос B **работает параллельно**, читает те же данные, пишет — EXEC fail → retry |
| **Количество round-trips** | Lock (1) + ZREM + ZCOUNT + ZADD (3) + DEL (1) = 5+ | WATCH + ZRANGE (2) + MULTI+ZREM+ZADD+EXEC (1) = 3–4, но при retry — ещё раз |
| **Латентность при низкой конкуренции** | Выше: всегда lock + unlock | Ниже: нет ожидания lock |
| **Латентность при высокой конкуренции** | Один работает, остальные ждут (предсказуемо) | Все конкурируют, много retry (непредсказуемо) |
| **Redis команды** | SET NX, DEL, ZREMRANGEBYSCORE, ZCOUNT, ZADD | WATCH, ZRANGE, MULTI, ZREMRANGEBYSCORE, ZADD, EXEC |

**Ключевая идея:**
- **L2:** «Пессимист» — предполагает конфликт, блокирует доступ заранее.
- **L3:** «Оптимист» — предполагает, что конфликт редок, работает без блокировки и откатывается при конфликте.

---

## Методика реализации L3: как пришёл к решению

### 1. Исходные данные (spec)

- `docs/subtask-3-optimistic.md`: WATCH → ZRANGE → filter → MULTI + ZADD → EXEC; при nil от EXEC — retry.
- `docs/strategies-overview.md`: «Optimistic locking; one of two concurrent updaters retries».

### 2. Выбор API go-redis

- Нужны: WATCH, чтение, MULTI/EXEC.
- В go-redis: `client.Watch(ctx, fn, key)` — callback выполняется на «watched» connection.
- Внутри callback: `tx.TxPipelined(ctx, fn)` — MULTI + команды + EXEC.
- При изменении key между WATCH и EXEC: `TxPipelined` возвращает `redis.TxFailedErr`.

### 3. Структура алгоритма

```
loop (max 3 retries):
  err = Watch(key, func(tx):
    entries = tx.ZRangeWithScores(key)
    count = filter(entries, cutoff)
    if count >= limit → return errOverLimit  // 429, не retry
    tx.TxPipelined(ZRemRangeByScore, ZAdd)
  )
  if err == nil → 200
  if err == errOverLimit → 429
  if err == TxFailedErr → continue (retry)
  else → 500
```

### 4. Решения по реализации

| Вопрос | Решение | Причина |
|--------|---------|---------|
| ZRemRangeByScore в транзакции? | Да | Иначе старые записи накапливаются; trim и add должны быть атомарны |
| errOverLimit — отдельный тип? | Да, sentinel error | Отличать «лимит превышен» (429) от «retry» и «ошибка Redis» (500) |
| maxRetries = 3? | Да | Как в spec; при bruteforce 3 retry достаточно, иначе 500 |
| Фильтр client-side? | Да | ZRANGE возвращает все; окно применяем в коде (cutoff) |

### 5. Соответствие L2

- Тот же ключ `rate:{ip}`.
- Тот же формат ZSET: score = timestamp, member = unique id.
- Тот же sliding window: cutoff = now - windowSec.
- Отличие: нет `lock:rate:{ip}` — L3 не использует блокировки.

### 6. Где код

Реализация: `go/internal/adapter/outbound/redis/optimistic/optimistic.go`

- `OptimisticLimiter.Allow()` — основной метод
- `client.Watch(ctx, fn, key)` — обёртка над WATCH
- `tx.TxPipelined(ctx, fn)` — MULTI + ZRemRangeByScore + ZAdd + EXEC
- Endpoint: `POST /resource/optimistic`

---

## Сводная таблица

| Уровень | Где храним | Как избегаем race | Стоимость |
|---------|------------|-------------------|-----------|
| L1 Naive | Память (map) | Mutex (только один поток) | Не масштабируется |
| L2 Pessimistic | Redis ZSET | Lock (SET NX) перед работой | Блокирование, задержки |
| L3 Optimistic | Redis ZSET | WATCH + retry при конфликте | Много retry при конкуренции |
| L4 Atomic | Redis ZSET | Lua-скрипт (всё в одном вызове) | Минимальная, production |

---

## Полезные команды Redis для отладки

```bash
# Подключиться к Redis
redis-cli

# Посмотреть ключи по паттерну
KEYS rate:*

# Посмотреть содержимое Sorted Set
ZRANGE rate:192.168.1.1 0 -1 WITHSCORES

# Посмотреть TTL блокировки
TTL lock:rate:192.168.1.1

# Удалить ключ (очистить для теста)
DEL rate:192.168.1.1
```
