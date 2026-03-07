# 📚 **FAQ: Типичные ошибки и их решения**
# 📚 **FAQ: Common Errors and Their Solutions**

---

## 🐛 **ОШИБКА 1: Redis не подключается (ENOTFOUND redis)**
## 🐛 **ERROR 1: Redis connection fails (ENOTFOUND redis)**

**Симптом / Symptoms:**
```
❌ Redis error: getaddrinfo ENOTFOUND redis
```

**Причина / Cause:**
Код пытается подключиться к Redis по hostname `redis`, но при локальном запуске нужно `localhost`.
The code tries to connect to Redis using hostname `redis`, but locally it should use `localhost`.

**Решение / Solution:**
```typescript
// В config/redis.ts / In config/redis.ts
const redisHost = process.env.REDIS_HOST || 'localhost'; // важно: НЕ 'redis' / important: NOT 'redis'
const redisUrl = `redis://${redisHost}:6379`;
```

**В `.env` / In `.env`:**
```env
REDIS_HOST=localhost  # для локального запуска / for local development
# REDIS_HOST=redis    # для запуска в Docker / for Docker
```

---

## 🐛 **ОШИБКА 2: PostgreSQL не подключается (ENOTFOUND postgres)**
## 🐛 **ERROR 2: PostgreSQL connection fails (ENOTFOUND postgres)**

**Симптом / Symptoms:**
```
❌ Failed to start server: Error: getaddrinfo ENOTFOUND postgres
```

**Причина / Cause:**
Аналогично Redis — код ищет PostgreSQL по hostname `postgres`.
Similar to Redis — code looks for PostgreSQL using hostname `postgres`.

**Решение / Solution:**
```typescript
// В config/database.ts / In config/database.ts
const dbHost = process.env.DB_HOST || 'localhost'; // важно: НЕ 'postgres' / important: NOT 'postgres'
```

**В `.env` / In `.env`:**
```env
DB_HOST=localhost     # для локального запуска / for local development
# DB_HOST=postgres    # для запуска в Docker / for Docker
```

---

## 🐛 **ОШИБКА 3: Cannot find module './config/database'**
## 🐛 **ERROR 3: Cannot find module './config/database'**

**Симптом / Symptoms:**
```
Error: Cannot find module './config/database' imported from './server.ts'
```

**Причина / Cause:**
Проблема с расширениями в ES модулях (`type: "module"` в package.json).
Issue with extensions in ES modules (`type: "module"` in package.json).

**Вариант А (убрать .js из импортов) / Option A (remove .js from imports):**
```typescript
// В server.ts / In server.ts
import { pgPool } from './config/database';        // ✅ правильно / correct
import { pgPool } from './config/database.js';     // ❌ ошибка если файл .ts / error if file is .ts
```

**Вариант Б (использовать tsx вместо ts-node) / Option B (use tsx instead of ts-node):**
```json
// В package.json / In package.json
{
  "scripts": {
    "dev": "tsx watch src/server.ts"  // ✅ tsx понимает ES модули / tsx understands ES modules
  }
}
```

---

## 🐛 **ОШИБКА 4: Called end on pool more than once**
## 🐛 **ERROR 4: Called end on pool more than once**

**Симптом / Symptoms:**
```
❌ Error during shutdown: Error: Called end on pool more than once
```

**Причина / Cause:**
Несколько обработчиков сигналов пытаются закрыть пул соединений повторно.
Multiple signal handlers try to close the connection pool repeatedly.

**Решение / Solution:**
```typescript
// В server.ts / In server.ts — only ONE handler
const gracefulShutdown = async (signal: string) => {
  console.log(`\n🛑 Received ${signal}, shutting down...`);
  await fastify.close();
  await pgPool.end();           // закрыть только один раз / close only once
  if (redisClient.isReady) {
    await redisClient.quit();
  }
  process.exit(0);
};

// Только эти два обработчика, никаких других / Only these two handlers, no others
process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
```

---

## 🐛 **ОШИБКА 5: Сервер компилируется но не запускается**
## 🐛 **ERROR 5: Server compiles but doesn't start**

**Симптом / Symptoms:**
```
CLI Build success in 62ms
CLI Watching for changes...
# Но сервер не отвечает на curl / But server doesn't respond to curl
```

**Причина / Cause:**
`tsup` только компилирует, но не запускает сервер.
`tsup` only compiles, doesn't run the server.

**Решение / Solution:**
```json
// В package.json / In package.json
{
  "scripts": {
    "dev": "tsx watch src/server.ts",     // ✅ и компилирует и запускает / compiles and runs
    "build": "tsup src/server.ts",         // только для продакшена / production only
  }
}
```

---

## 🐛 **ОШИБКА 6: Порт 3000 уже занят**
## 🐛 **ERROR 6: Port 3000 already in use**

**Симптом / Symptoms:**
```
Error: listen EADDRINUSE: address already in use :::3000
```

**Решение / Solution:**
```bash
# Найти процесс на порту 3000 / Find process on port 3000
lsof -i :3000

# Убить процесс / Kill process
kill -9 <PID>

# Или использовать другой порт в .env / Or use different port in .env
PORT=3001
```

---

## 🐛 **ОШИБКА 7: Docker контейнеры не стартуют**
## 🐛 **ERROR 7: Docker containers won't start**

**Симптом / Symptoms:**
```
ERROR: for postgres  Cannot start service postgres: driver failed programming external connectivity
```

**Решение / Solution:**
```bash
# Остановить всё / Stop everything
docker-compose down

# Удалить конфликтующие контейнеры / Remove conflicting containers
docker rm -f inventory-postgres inventory-redis

# Запустить заново / Start again
docker-compose up -d
```

---

## 🐛 **ОШИБКА 8: PostgreSQL authentication failed**
## 🐛 **ERROR 8: PostgreSQL authentication failed**

**Симптом / Symptoms:**
```
error: password authentication failed for user "postgres"
```

**Решение / Solution:**
```bash
# Проверить .env / Check .env
cat .env | grep DB_PASSWORD

# Должно быть / Should be:
DB_PASSWORD=postgres

# Если пароль другой — сбросить контейнер / If password different — reset container
docker-compose down -v
docker-compose up -d
```

---

## ✅ **ЧЕК-ЛИСТ ПЕРЕД ЗАПУСКОМ / PRE-RUN CHECKLIST**

### 🔍 **Проверить .env / Check .env:**
```env
DB_HOST=localhost     # для локального запуска / for local development
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres

REDIS_HOST=localhost  # для локального запуска / for local development
REDIS_PORT=6379
```

### 🐳 **Проверить Docker / Check Docker:**
```bash
docker ps
# Должны быть оба контейнера: inventory-postgres и inventory-redis
# Both containers should be running: inventory-postgres and inventory-redis
```

### 📁 **Проверить структуру / Check structure:**
```
src/
├── config/
│   ├── database.ts
│   └── redis.ts
└── server.ts
```

### 🚀 **Запустить / Run:**
```bash
npm run dev
```

### 🧪 **Проверить / Verify:**
```bash
curl http://localhost:3000/health
```

---

## 📝 **КОМАНДЫ ДЛЯ БЫСТРОЙ ДИАГНОСТИКИ / QUICK DIAGNOSTIC COMMANDS**

```bash
# Проверить что слушает порт 3000 / Check what's listening on port 3000
lsof -i :3000

# Проверить логи Docker / Check Docker logs
docker logs inventory-postgres
docker logs inventory-redis

# Подключиться к PostgreSQL / Connect to PostgreSQL
docker exec -it inventory-postgres psql -U postgres -d inventory

# Проверить Redis / Check Redis
docker exec -it inventory-redis redis-cli ping

# Перезапустить всё с чистого листа / Clean restart everything
make infra-down
make infra-up
rm -rf node_modules
npm install
npm run dev
```

---

## 💡 **ПОЛЕЗНЫЕ СОВЕТЫ / PRO TIPS**

### **Для разработки / For development:**
- Всегда используйте `tsx watch` для разработки — он понимает ES модули
- Always use `tsx watch` for development — it understands ES modules
- Держите `.env` в `.gitignore` — никогда не коммитьте его
- Keep `.env` in `.gitignore` — never commit it
- Используйте `.env.example` как шаблон
- Use `.env.example` as a template

### **Для Docker / For Docker:**
- В Docker контейнерах hostname = имя сервиса (`postgres`, `redis`)
- In Docker containers, hostname = service name (`postgres`, `redis`)
- Для локальной разработки hostname = `localhost`
- For local development, hostname = `localhost`

### **Для отладки / For debugging:**
```bash
# Включить подробные логи / Enable verbose logging
NODE_ENV=development npm run dev

# Посмотреть все переменные окружения / Check all environment variables
node -e "console.log(process.env)" | grep DB
```
