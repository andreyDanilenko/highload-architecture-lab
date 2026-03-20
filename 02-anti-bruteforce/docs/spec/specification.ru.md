# Best Practices для реализации Rate Limiting
## Универсальное руководство по защите от брутфорса и DoS-атак

---

## 1. Идентификация клиента (Client Identification)

### 1.1. Корректное извлечение IP-адреса
Самая частая ошибка — использование `req.remoteAddr` напрямую. В продакшене приложение всегда стоит за прокси (nginx, ingress, Cloudflare).

**Правильная реализация:**

```go
type ClientIdentifier struct {
    // Список доверенных прокси (CIDR)
    TrustedProxies []*net.IPNet
    // Заголовки для извлечения IP в порядке приоритета
    Headers []string
}

func (ci *ClientIdentifier) GetClientIP(r *http.Request) string {
    // 1. Проверяем заголовки от доверенных прокси
    for _, header := range ci.Headers {
        if headerValue := r.Header.Get(header); headerValue != "" {
            // Проверяем, что запрос пришел от доверенного прокси
            if ci.isFromTrustedProxy(r.RemoteAddr) {
                return ci.extractFirstIP(headerValue)
            }
        }
    }
    
    // 2. Fallback на RemoteAddr (обрезаем порт)
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    return host
}

// Нормализация IPv6
func normalizeIP(ip string) string {
    parsed := net.ParseIP(ip)
    if parsed == nil {
        return ip
    }
    return parsed.String() // Приводит к каноничному виду
}
```

**Конфигурация по умолчанию:**
```yaml
client_identification:
  trusted_proxies:
    - "10.0.0.0/8"      # Внутренние сети
    - "172.16.0.0/12"
    - "192.168.0.0/16"
  headers:
    - "X-Forwarded-For"  # Стандартный прокси-заголовок
    - "X-Real-IP"        # Альтернативный заголовок
  strip_port: true       # Всегда убираем порт
```

### 1.2. Формирование ключей для rate limiting

**Никогда не используйте IP с портом** — это позволяет обойти защиту простым открытием множества соединений.

**Правильные композитные ключи:**

```go
type RateLimitKey struct {
    Components []string
    Separator  string
}

func (k *RateLimitKey) Build() string {
    return strings.Join(k.Components, k.Separator)
}

// Примеры ключей для разных сценариев
var keys = struct {
    // Для защиты от брутфорса логина
    LoginByIP func(ip, endpoint string) *RateLimitKey
    LoginByUser func(username string) *RateLimitKey
    LoginCombined func(ip, username string) *RateLimitKey
    
    // Для API по ключам
    APIKey func(apiKey, endpoint string) *RateLimitKey
    
    // Для авторизованных пользователей
    User func(userID, endpoint string) *RateLimitKey
}{
    LoginByIP: func(ip, endpoint string) *RateLimitKey {
        return &RateLimitKey{
            Components: []string{"ratelimit", "login", "ip", normalizeIP(ip), endpoint},
            Separator:  ":",
        }
    },
    LoginByUser: func(username string) *RateLimitKey {
        return &RateLimitKey{
            Components: []string{"ratelimit", "login", "user", username},
            Separator:  ":",
        }
    },
    // ... остальные комбинации
}
```

**Правило формирования ключей:**
- Всегда включайте тип лимита (login, api, general)
- Всегда нормализуйте IP (без порта, каноничный формат)
- Для пользовательских идентификаторов используйте хеши (PII protection)
- В продакшене — добавляйте префикс для изоляции окружений (dev/staging/prod)

---

## 2. Многослойная архитектура защиты

### 2.1. Схема распределения ответственности

```mermaid
graph TD
    A[Client] --> B[CDN/WAF Layer]
    B --> C[API Gateway/Ingress]
    C --> D[Application Layer]
    
    B1[Грубые IP-лимиты<br/>CAPTCHA/JS challenge<br/>Блокировка стран] --> B
    C1[Token Bucket по IP/API-key<br/>Сохранение ресурсов backend] --> C
    D1[Бизнес-логика<br/>Антибрутфорс по login<br/>Временные блокировки] --> D
```

### 2.2. Конфигурация для каждого слоя

**Layer 1: CDN/WAF (Cloudflare, AWS WAF, Cloud Armor)**
```yaml
# Глобальная защита инфраструктуры
rate_limits:
  - name: "global_ip_limit"
    key: "${ip}"
    matcher: 
      path: "/api/*"
    limits:
      - count: 1000
        period: "1m"      # Не больше 1000 запросов в минуту с IP
      - count: 10000
        period: "1h"       # Не больше 10000 в час
        
  - name: "login_endpoint_protection"
    key: "${ip}"
    matcher:
      path: "/api/v1/login"
    limits:
      - count: 10
        period: "1m"       # 10 попыток в минуту на /login
    action: "challenge"    # При превышении - показать CAPTCHA
```

**Layer 2: API Gateway / Ingress (NGINX, Envoy, Kong)**
```lua
-- Пример для Kong/OpenResty
local limits = {
    -- Базовый лимит на все запросы
    {
        key = ngx.var.binary_remote_addr,
        window = "60s",
        limit = 1000,
    },
    -- Лимит на авторизационные endpoints
    {
        key = ngx.var.binary_remote_addr,
        window = "60s",
        limit = 30,
        path = {"/login", "/register", "/reset-password"}
    }
}
```

**Layer 3: Application Layer (ваш код)**
```go
type RateLimitMiddleware struct {
    // Слайсинг по времени
    slidingWindow   *SlidingWindowLog
    // Защита от гонок
    pessimisticLock *RedisLock
    // Разные лимиты для разных стратегий
    limits          map[string]*LimitRule
}

type LimitRule struct {
    Key        string // Шаблон ключа
    Window     time.Duration
    MaxRequests int64
    Strategy   string // "sliding_window", "token_bucket", "leaky_bucket"
    Burst      int64  // Для token bucket
}
```

---

## 3. Алгоритмы и структуры данных

### 3.1. Сравнение алгоритмов для разных сценариев

| Алгоритм | Use Case | Плюсы | Минусы | Prod-ready |
|----------|----------|-------|--------|------------|
| **Fixed Window** | Общие лимиты, мониторинг | Простой, минимальный overhead | Проблема с границами окон | ✅ Да |
| **Sliding Window Log** | Антибрутфорс, точные лимиты | Максимально точный | Много памяти в Redis | ✅ Да (с Lua) |
| **Sliding Window Counter** | API limits, высокая нагрузка | Компромисс точность/память | Меньше точности чем Log | ✅ Да |
| **Token Bucket** | Traffic shaping, burst | Поддерживает burst, плавный | Сложнее отладка | ✅ Да |
| **Leaky Bucket** | Стабильный outflow | Предсказуемый outflow | Не гибкий | ⚠️ Редко |

### 3.2. Production-ready Lua скрипт для Redis (Sliding Window Log)

```lua
-- KEYS[1] = ключ (например "ratelimit:login:ip:192.168.1.1")
-- ARGV[1] = текущее время (unix timestamp)
-- ARGV[2] = размер окна в секундах
-- ARGV[3] = максимальное количество запросов
-- ARGV[4] = вес запроса (обычно 1)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local max = tonumber(ARGV[3])
local weight = tonumber(ARGV[4]) or 1

-- Удаляем устаревшие записи
redis.call('ZREMRANGEBYSCORE', key, 0, now - window * 1000)

-- Считаем текущее количество
local current = redis.call('ZCARD', key)

if current + weight > max then
    -- Получаем время до истечения самого старого элемента
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local retry_after = 0
    
    if oldest[2] then
        retry_after = math.ceil((oldest[2] + window * 1000 - now) / 1000)
    end
    
    return {
        0,                      -- allowed
        current,                -- current_count
        max,                    -- limit
        retry_after,            -- retry_after_seconds
        now / 1000              -- current_time_seconds
    }
end

-- Добавляем новый запрос
redis.call('ZADD', key, now, now .. ':' .. math.random())
redis.call('EXPIRE', key, window * 2)  -- TTL с запасом

return {
    1,                          -- allowed
    current + weight,           -- current_count
    max,                        -- limit
    0,                          -- retry_after_seconds
    now / 1000                  -- current_time_seconds
}
```

---

## 4. Бизнес-логика и реакции

### 4.1. Многоуровневые лимиты для защиты логина

```go
type LoginProtection struct {
    // Лимиты для разных уровней защиты
    levels []struct {
        attempts    int
        window      time.Duration
        action      func(*http.Request, string)
        blockPeriod time.Duration
    }
}

func (lp *LoginProtection) Check(ctx context.Context, ip, login string) (*Result, error) {
    // 1. Проверяем по IP
    ipResult := lp.checkIP(ctx, ip)
    
    // 2. Проверяем по логину (защита от распределенных атак)
    loginResult := lp.checkLogin(ctx, login)
    
    // 3. Проверяем комбинацию IP+login
    combinedResult := lp.checkCombined(ctx, ip, login)
    
    // 4. Определяем severity и действие
    severity := lp.calculateSeverity(ipResult, loginResult, combinedResult)
    
    switch severity {
    case SeverityLow:
        // Просто логируем
        return &Result{Allowed: true}, nil
        
    case SeverityMedium:
        // Возвращаем 429 с Retry-After
        return &Result{
            Allowed:     false,
            StatusCode:  429,
            RetryAfter:  30,
            Message:     "Too many attempts, please try again later",
        }, nil
        
    case SeverityHigh:
        // Временная блокировка аккаунта
        lp.tempBlockAccount(ctx, login, 24*time.Hour)
        return &Result{
            Allowed:     false,
            StatusCode:  403,
            Message:     "Account temporarily locked. Contact support.",
        }, nil
        
    case SeverityCritical:
        // Требуем дополнительную верификацию
        return &Result{
            Allowed:     false,
            StatusCode:  401,
            Require2FA:  true,
            Message:     "Additional verification required",
        }, nil
    }
}
```

### 4.2. Retry-After и информирование клиента

```go
type RateLimitResponse struct {
    StatusCode    int           `json:"-"`
    Message       string        `json:"message,omitempty"`
    RetryAfter    int           `json:"retry_after,omitempty"` // секунды
    Limit         int64         `json:"limit"`
    Remaining     int64         `json:"remaining"`
    Reset         int64         `json:"reset"` // unix timestamp
}

func (rl *RateLimitMiddleware) writeRateLimitHeaders(w http.ResponseWriter, r *RateLimitResponse) {
    w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(r.Limit, 10))
    w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(r.Remaining, 10))
    w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(r.Reset, 10))
    
    if r.RetryAfter > 0 {
        w.Header().Set("Retry-After", strconv.Itoa(r.RetryAfter))
        // RFC 7231 формат даты
        retryTime := time.Now().Add(time.Duration(r.RetryAfter) * time.Second)
        w.Header().Set("Retry-After", retryTime.Format(time.RFC1123))
    }
    
    w.WriteHeader(r.StatusCode)
    json.NewEncoder(w).Encode(r)
}
```

---

## 5. Observability и мониторинг

### 5.1. Метрики для Prometheus

```go
type RateLimitMetrics struct {
    // Счетчики превышений
    exceededTotal *prometheus.CounterVec
    
    // Текущее состояние лимитов
    currentLoad *prometheus.GaugeVec
    
    // Время обработки
    checkDuration *prometheus.HistogramVec
    
    // Активные блокировки
    activeBlocks *prometheus.GaugeVec
}

func InitMetrics() *RateLimitMetrics {
    m := &RateLimitMetrics{
        exceededTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "ratelimit_exceeded_total",
                Help: "Total number of rate limit exceeded events",
            },
            []string{"limiter_type", "endpoint", "severity"},
        ),
        
        currentLoad: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "ratelimit_current_load",
                Help: "Current load per limiter",
            },
            []string{"limiter_type", "key"},
        ),
        
        checkDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "ratelimit_check_duration_seconds",
                Help:    "Time spent checking rate limits",
                Buckets: prometheus.DefBuckets,
            },
            []string{"limiter_type"},
        ),
    }
    
    prometheus.MustRegister(m.exceededTotal, m.currentLoad, m.checkDuration)
    return m
}
```

### 5.2. Структурированное логирование

```go
type RateLimitLog struct {
    Level       string                 `json:"level"`
    Timestamp   time.Time              `json:"@timestamp"`
    Type        string                 `json:"type"` // "rate_limit_exceeded", "account_blocked"
    
    // Контекст запроса
    IP          string                 `json:"client_ip"`
    UserID      string                 `json:"user_id,omitempty"`
    Login       string                 `json:"login,omitempty"`
    Endpoint    string                 `json:"endpoint"`
    Method      string                 `json:"method"`
    
    // Лимиты
    LimitType   string                 `json:"limit_type"` // "ip", "login", "combined"
    Limit       int64                   `json:"limit"`
    Current     int64                   `json:"current"`
    Window      int                     `json:"window_seconds"`
    
    // Результат
    Allowed     bool                    `json:"allowed"`
    StatusCode  int                     `json:"status_code"`
    RetryAfter  int                     `json:"retry_after,omitempty"`
    
    // Метаданные
    RequestID   string                 `json:"request_id"`
    Environment string                 `json:"environment"`
}
```

### 5.3. Алерты

```yaml
alerts:
  - name: "High Rate Limit Exceeded Rate"
    condition: "rate(ratelimit_exceeded_total[5m]) > 100"
    severity: "warning"
    description: "High number of rate limit violations"
    
  - name: "Multiple Accounts Blocked"
    condition: "increase(ratelimit_exceeded_total{severity='high'}[10m]) > 10"
    severity: "critical"
    description: "Possible distributed attack in progress"
    
  - name: "Redis Latency"
    condition: "histogram_quantile(0.95, rate(ratelimit_check_duration_bucket[5m])) > 0.1"
    severity: "warning"
    description: "Rate limiter is slowing down requests"
```

---

## 6. Конфигурируемость и feature flags

### 6.1. Гибкая конфигурация (YAML + env)

```yaml
rate_limiter:
  enabled: true
  
  # Глобальные настройки
  default_strategy: "sliding_window"
  key_prefix: "prod:ratelimit"
  
  # Источник IP
  ip_source:
    trust_proxies: true
    proxies_cidr:
      - "10.0.0.0/8"
      - "172.16.0.0/12"
      - "192.168.0.0/16"
    headers_priority:
      - "X-Forwarded-For"
      - "X-Real-IP"
  
  # Правила для разных эндпоинтов
  rules:
    - name: "login_ip"
      endpoint: "/api/v1/login"
      key_pattern: "{ip}"
      window: "1m"
      limit: 5
      strategy: "sliding_window_log"
      severity: "medium"
      
    - name: "login_user"
      endpoint: "/api/v1/login"
      key_pattern: "user:{login}"
      window: "15m"
      limit: 10
      strategy: "sliding_window_log"
      severity: "high"
      on_exceed:
        - action: "block_account"
          duration: "1h"
        - action: "require_captcha"
          
    - name: "api_general"
      endpoint: "/api/v1/*"
      key_pattern: "{api_key}"
      window: "1h"
      limit: 1000
      burst: 100
      strategy: "token_bucket"
      
  # Redis настройки
  redis:
    addresses: ["localhost:6379"]
    cluster_mode: false
    max_retries: 3
    pool_size: 100
    
  # Мониторинг
  metrics:
    enabled: true
    namespace: "myapp"
    
  # Fallback при недоступности Redis
  circuit_breaker:
    enabled: true
    failure_threshold: 10
    timeout: "5s"
```

### 6.2. Feature flags для постепенного включения

```go
type RateLimitConfig struct {
    // Feature flags
    EnableLoginRateLimit  bool `feature:"login-rate-limit"`
    EnableStrictMode      bool `feature:"strict-rate-limit"`
    EnableAccountBlocking bool `feature:"account-blocking"`
    
    // Canary deployment
    CanaryPercent    int      `feature:"rate-limit-canary"`
    CanaryIPs        []string `feature:"rate-limit-whitelist"`
    
    // Динамические лимиты
    DynamicLimits     map[string]int64 `config:"dynamic-limits"`
}
```

---

## 7. Тестирование и валидация

### 7.1. Нагрузочное тестирование

```go
func TestRateLimiterUnderLoad(t *testing.T) {
    // Тест на burst-нагрузку
    t.Run("burst traffic", func(t *testing.T) {
        limiter := NewRateLimiter(config)
        
        // 1000 запросов одновременно
        var wg sync.WaitGroup
        for i := 0; i < 1000; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                _, err := limiter.Check(context.Background(), "192.168.1.1", "/api/test")
                assert.NoError(t, err)
            }()
        }
        wg.Wait()
    })
    
    // Тест на race conditions
    t.Run("race conditions", func(t *testing.T) {
        // Запускаем много горутин с одним IP
        // Проверяем, что лимит не превышен больше чем на допустимую погрешность
    })
    
    // Тест с распределенными клиентами
    t.Run("distributed clients", func(t *testing.T) {
        // Разные IP атакуют одного пользователя
    })
}
```

### 7.2. Тесты безопасности

```python
def test_rate_limit_bypass_attempts():
    """Тестирование попыток обхода rate limit"""
    
    # Попытка с разными портами
    for port in [8080, 8081, 8082]:
        ip = f"192.168.1.1:{port}"
        response = client.post("/login", 
                             json={"username": "admin", "password": "wrong"},
                             origin_ip=ip)
        # Должен лимитироваться одинаково
        assert response.status_code == 429
        
    # Попытка с подменой X-Forwarded-For
    response = client.post("/login",
                          headers={"X-Forwarded-For": "1.2.3.4"},
                          json={"username": "admin", "password": "wrong"})
    # Не должен доверять заголовку от клиента
    assert response.status_code != 429  # Если реальный IP не превысил лимит
```

---

## 8. Чек-лист для внедрения

### ✅ Must have (критично для продакшена)
- [ ] Корректное извлечение IP с учетом доверенных прокси
- [ ] Ключи rate limit **без порта**
- [ ] Композитные ключи (IP + login для защиты от распределенных атак)
- [ ] Атомарные операции через Redis Lua
- [ ] Защита от race conditions
- [ ] HTTP заголовки (RateLimit-*, Retry-After)
- [ ] Метрики в Prometheus
- [ ] Логирование всех превышений
- [ ] Graceful degradation при недоступности Redis

### ⚠️ Should have (важно для production readiness)
- [ ] Feature flags для постепенного включения
- [ ] Динамическая конфигурация без деплоя
- [ ] Алерты на аномалии
- [ ] Документация для разработчиков
- [ ] Интеграционные тесты
- [ ] Нагрузочное тестирование

### 🚀 Nice to have (для enterprise уровня)
- [ ] Distributed rate limiting (consistent hashing)
- [ ] Машинное обучение для обнаружения аномалий
- [ ] Автоматическая блокировка ботнетов
- [ ] Интеграция с SIEM системами
- [ ] A/B тестирование разных стратегий
- [ ] Self-tuning лимитов

---

## Заключение

Rate limiting — это не просто техническая реализация, а комплексная система безопасности. Ключевые принципы:

1. **Trust but verify** — никогда не доверяйте входным данным клиента
2. **Defense in depth** — несколько слоев защиты
3. **Observability** — если не измеряете, то не контролируете
4. **Fail safe** — при сбоях блокируйте, а не пропускайте
5. **Continuous improvement** — анализируйте атаки и улучшайте защиту
