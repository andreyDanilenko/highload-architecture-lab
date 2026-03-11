# Redis: Полный конспект

## 1. TTL (Time To Live) - Время жизни ключа

### Базовые команды
```bash
# Установка ключа с TTL
SET key "value" EX 60           # Живет 60 секунд
SETEX key 60 "value"            # То же самое

# Управление TTL
TTL key                          # Сколько осталось секунд (-2 нет ключа, -1 бесконечно)
EXPIRE key 30                    # Установить TTL на существующий ключ
PERSIST key                       # Убрать TTL (сделать вечным)
```

### Практическое применение
- Сессии пользователей (30 минут)
- Кеш данных (5-10 минут)
- Одноразовые ссылки (24 часа)
- Rate limiting (счетчик на минуту)

---

## 2. STRING - Простые строки

### Основные операции
```bash
SET key "value"                   # Создать/обновить
GET key                           # Получить значение
DEL key                           # Удалить

# Числовые операции
INCR key                          # Увеличить на 1
INCRBY key 5                      # Увеличить на 5
DECR key                          # Уменьшить на 1
DECRBY key 3                      # Уменьшить на 3

# Пакетные операции
MSET key1 "val1" key2 "val2"      # Множественная запись
MGET key1 key2                     # Множественное чтение
```

### Кейсы
1. **Счетчик просмотров**: `INCR page:home:views`
2. **Rate limiting**: `INCR rate:ip` + `EXPIRE rate:ip 60`
3. **Кеш**: `SET product:123 data EX 300`

---

## 3. LIST - Списки (очереди и стеки)

### Операции
```bash
# Добавление
LPUSH list "a"                     # Слева (в начало)
RPUSH list "b"                     # Справа (в конец)

# Извлечение
LPOP list                           # Слева
RPOP list                           # Справа

# Просмотр
LRANGE list 0 -1                    # Все элементы
LRANGE list 0 4                     # Первые 5
LLEN list                           # Длина списка

# Управление
LTRIM list 0 9                      # Оставить только первые 10
```

### Кейсы
1. **Очередь задач**: `LPUSH tasks job` + `RPOP tasks` (FIFO)
2. **Стек**: `LPUSH stack item` + `LPOP stack` (LIFO)
3. **История действий**: `LPUSH history user:1 action` + `LTRIM history 0 9`

---

## 4. SET - Множества (уникальные значения)

### Операции
```bash
# Добавление/удаление
SADD set "member"                   # Добавить
SREM set "member"                    # Удалить
SMEMBERS set                         # Все элементы
SISMEMBER set "member"               # Проверка наличия (1/0)
SCARD set                            # Количество элементов

# Множественные операции
SUNION set1 set2                     # Объединение
SINTER set1 set2                      # Пересечение
SDIFF set1 set2                       # Разность
```

### Кейсы
1. **Уникальные посетители**: `SADD visitors:2024-01-15 ip1 ip2`
2. **Теги поста**: `SADD post:42:tags "redis" "golang"`
3. **Общие друзья**: `SINTER friends:alice friends:bob`

---

## 5. HASH - Хеш-таблицы (объекты)

### Операции
```bash
# Запись
HSET user:1 name "Alice" age "30"   # Создать/обновить поля
HMSET user:1 email "a@mail.com"      # Множественная запись (устарело, можно HSET)

# Чтение
HGET user:1 name                      # Одно поле
HMGET user:1 name age                  # Несколько полей
HGETALL user:1                         # Все поля
HKEYS user:1                           # Только ключи
HVALS user:1                           # Только значения

# Числовые операции
HINCRBY user:1 login_count 1           # Увеличить счетчик
```

### Кейсы
1. **Профиль пользователя**: `HSET user:1001 name "Alice" email "a@mail.com"`
2. **Статистика**: `HINCRBY page:stats views 1` + `HINCRBY page:stats clicks 1`
3. **Сессия**: `HSET session:xyz user_id 1001 ip "1.2.3.4"` + `EXPIRE session:xyz 3600`

---

## 6. SORTED SET - Сортированные множества

### Операции
```bash
# Добавление
ZADD leaderboard 100 "player1"       # score = 100
ZADD leaderboard 200 "player2"

# Получение
ZRANGE leaderboard 0 -1               # По возрастанию (без scores)
ZRANGE leaderboard 0 -1 WITHSCORES    # С очками
ZREVRANGE leaderboard 0 -1             # По убыванию

# По диапазону score
ZRANGEBYSCORE leaderboard 100 200      # С score от 100 до 200

# Статистика
ZSCORE leaderboard "player1"           # Score игрока
ZCARD leaderboard                       # Количество
ZCOUNT leaderboard 100 200               # Количество в диапазоне
ZRANK leaderboard "player1"              # Место (по возрастанию)
ZREVRANK leaderboard "player1"           # Место (по убыванию)
```

### Кейсы
1. **Таблица лидеров**: `ZADD game:scores 1500 "user:42"`
2. **Рейтинг постов**: `ZADD posts:rating 100 "post:1"`
3. **Онлайн пользователи**: `ZADD online:users timestamp "user:42"`

---

## 7. Транзакции (MULTI/EXEC/WATCH)

### Базовые транзакции
```bash
MULTI                              # Начать транзакцию
    INCR key1
    DECR key2
EXEC                               # Выполнить все
DISCARD                            # Отменить транзакцию
```

### Оптимистичная блокировка
```bash
WATCH key                          # Следить за ключом
val = GET key                       # Прочитать
MULTI
    SET key new_val
EXEC                               # Если key изменился после WATCH - не выполнится
UNWATCH                            # Снять слежение
```

### Кейсы
1. **Атомарный перевод**: `DECR from` + `INCR to` в MULTI/EXEC
2. **Обновление остатков**: WATCH + проверка + MULTI/EXEC

---

## 8. Pub/Sub - Публикация/Подписка

### Команды
```bash
# Подписка
SUBSCRIBE channel                   # На один канал
PSUBSCRIBE news:*                    # На шаблон (все начинающиеся с news:)

# Публикация
PUBLISH channel "message"            # Отправить сообщение

# Управление
UNSUBSCRIBE channel                  # Отписаться
PUBSUB CHANNELS                       # Список активных каналов
```

### Особенности
- Сообщения не сохраняются
- Получают только активные подписчики
- 1 издатель → много подписчиков

---

## 9. Lua скрипты

### Основы
```bash
# Загрузка скрипта
SCRIPT LOAD "return redis.call('GET', KEYS[1])"

# Выполнение
EVALSHA sha 1 key arg1 arg2          # По SHA
EVAL "script" 1 key arg1 arg2         # Прямое выполнение

# Управление
SCRIPT EXISTS sha                     # Проверить существование
SCRIPT FLUSH                          # Удалить все скрипты
SCRIPT KILL                            # Убить выполняющийся скрипт
```

### Примеры
```lua
-- CAS операция
local current = redis.call('GET', KEYS[1])
if current == ARGV[1] then
    redis.call('SET', KEYS[1], ARGV[2])
    return 1
end
return 0

-- Rate limiter
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[2])
end
return current
```

---

## 10. Streams - Потоки данных

### Основные команды
```bash
# Добавление
XADD stream * field1 value1 field2 value2    # * = авто-ID
XADD stream MAXLEN ~ 1000 * field value      # Ограничение размера

# Чтение
XRANGE stream - +                             # Все с начала
XREVRANGE stream + - COUNT 10                  # Последние 10
XREAD COUNT 10 STREAMS stream 0                # Прочитать с ID=0

# Группы потребителей
XGROUP CREATE stream group $                   # Создать группу
XREADGROUP GROUP group consumer COUNT 1 STREAMS stream >  # Чтение
XACK stream group id                            # Подтверждение
XPENDING stream group                           # Неподтвержденные
```

### Кейсы
1. **Логи действий**: XADD user:actions * user_id 42 action "login"
2. **Очередь задач**: Группы потребителей для распределенной обработки
3. **Аудит событий**: Хранение истории изменений

---

## 11. Продвинутые структуры

### GEO - Геоданные
```bash
GEOADD cities 37.62 55.75 "Moscow"     # Добавить
GEODIST cities "Moscow" "SPb" km        # Расстояние
GEORADIUS cities 37.62 55.75 100 km     # Поиск вокруг точки
GEOPOS cities "Moscow"                   # Координаты
```

### HyperLogLog - Уникальные счетчики
```bash
PFADD visitors:day1 user1 user2 user1    # Добавить
PFCOUNT visitors:day1                     # ≈ уникальных
PFMERGE visitors:week visitors:day1 visitors:day2  # Объединить
```

### Bitmaps - Побитовые операции
```bash
SETBIT user:active 100 1                  # Установить бит 100 в 1
GETBIT user:active 100                     # Получить бит
BITCOUNT user:active                        # Количество 1
BITOP AND result user:active1 user:active2  # Побитовые операции
```

---

## 12. Администрирование

### Информация о сервере
```bash
INFO [section]                         # Статистика (server, memory, stats...)
CLIENT LIST                             # Список подключений
CLIENT KILL ip:port                     # Убить соединение
CONFIG GET *                            # Все настройки
CONFIG SET parameter value               # Изменить настройку
```

### Мониторинг
```bash
MONITOR                                 # Все команды в реальном времени
SLOWLOG GET 10                          # Последние медленные запросы
SLOWLOG LEN                              # Количество в логе
SLOWLOG RESET                            # Очистить лог
```

### Безопасность
```bash
AUTH password                           # Аутентификация
CONFIG SET requirepass "password"        # Установить пароль
```

### Бэкапы и восстановление
```bash
SAVE                                    # Синхронное сохранение
BGSAVE                                  # Асинхронное сохранение
LASTSAVE                                # Время последнего сохранения
FLUSHALL                                # Очистить всё
FLUSHDB                                 # Очистить текущую БД
```

---

## 13. Важные моменты и best practices

### Соглашения об именах
```
object:id:field          # user:1001:name
action:object:id         # likes:post:789
category:subcategory     # rate:1.2.3.4
```

### Паттерны проектирования
1. **Cache Aside**: Читать из кеша → если нет → читать из БД → положить в кеш
2. **Write Through**: Писать сразу и в кеш, и в БД
3. **Write Behind**: Писать в кеш, асинхронно сохранять в БД

### Ошибки и подводные камни
- **KEYS** на проде → используйте SCAN
- **Большие ключи** → могут заблокировать сервер
- **Отсутствие TTL** → утечка памяти
- **Неверный тип данных** → например, SET вместо HASH для объектов

---

## 14. Шпаргалка по возвращаемым значениям

| Команда | Ключ есть | Ключа нет |
|---------|-----------|-----------|
| GET | значение | (nil) |
| TTL | секунды / -1 | -2 |
| EXISTS | 1 | 0 |
| DEL | 1 | 0 |
| INCR | новое значение | 1 |
| SADD | 1 (добавлен) / 0 (уже был) | 1 (создан) |

---

## 15. Полезные комбинации (продвинутые кейсы)

### Сессия с автоматическим продлением
```bash
# При каждом действии
EXPIRE session:user 1800              # Продлить на 30 минут
```

### Рейтинг с лайками и дизлайками
```bash
ZADD post:rating 0 "post:1"            # Инициализация
ZINCRBY post:rating 1 "post:1"          # Лайк
ZINCRBY post:rating -1 "post:1"         # Дизлайк
```

### Очередь с приоритетом
```bash
# Высокий приоритет
LPUSH tasks:priority '{"task":"urgent"}'
# Низкий приоритет
RPUSH tasks:priority '{"task":"background"}'
# Забираем всегда слева
LPOP tasks:priority
```

### Уникальные посетители за период
```bash
PFADD visitors:2024-01-15 "ip1" "ip2"
PFADD visitors:2024-01-16 "ip2" "ip3"
PFMERGE visitors:weekend visitors:2024-01-15 visitors:2024-01-16
PFCOUNT visitors:weekend  # ≈ 3
```
