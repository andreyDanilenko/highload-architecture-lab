# Rate Limiting Best Practices
## Universal Architecture and Implementation Guide

---

## Document Information

| | |
|---|---|
| **Audience** | Developers, Architects, DevOps, Security Engineers |
| **Language** | Language-agnostic concepts with pseudocode |
| **Scope** | Brute force protection, DoS mitigation, API rate limiting |
| **Version** | 1.0.0 |

---

## Table of Contents

1. [Client Identification](#1-client-identification)
2. [Multi-layered Architecture](#2-multi-layered-architecture)
3. [Algorithms and Data Structures](#3-algorithms-and-data-structures)
4. [Business Logic and Reactions](#4-business-logic-and-reactions)
5. [Observability and Monitoring](#5-observability-and-monitoring)
6. [Configuration and Feature Flags](#6-configuration-and-feature-flags)
7. [Testing and Validation](#7-testing-and-validation)
8. [Implementation Checklist](#8-implementation-checklist)

---

## 1. Client Identification

### 1.1. Correct IP Address Extraction

**The Problem:**
Most applications sit behind proxies (nginx, ingress, CDN). Using `request.remoteAddr` directly gives you the proxy IP, not the real client.

**The Solution:**

```
function getClientIP(request):
    # Step 1: Check if request came from trusted proxy
    if isFromTrustedProxy(request.remoteAddr):
        # Step 2: Extract IP from headers in priority order
        for header in ["X-Forwarded-For", "X-Real-IP"]:
            if header exists in request.headers:
                # X-Forwarded-For can contain multiple IPs
                # Take the FIRST one (client IP)
                return first(parseHeader(request.headers[header]))
    
    # Step 3: Fallback to remote address (without port)
    return stripPort(request.remoteAddr)
```

**What Makes a Proxy "Trusted"?**

```
Trusted proxies are IPs/networks YOU control:
- Internal load balancers (10.0.0.0/8, 172.16.0.0/12)
- Kubernetes ingress controllers
- Cloudflare IP ranges (if configured)
- AWS ALB internal IPs

NEVER trust headers from untrusted sources!
```

**Configuration Structure:**

```yaml
ip_extraction:
  trust_proxies: true
  trusted_cidrs:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"
  headers_priority:
    - "X-Forwarded-For"
    - "X-Real-IP"
  strip_port: true
  normalize_ipv6: true
```

### 1.2. Rate Limit Key Formation

**Critical Rule:**
```
NEVER include port in IP-based keys!

Bad:  "ratelimit:192.168.1.1:8080"  ← Attacker can open 1000 ports
Good: "ratelimit:192.168.1.1"        ← All connections share limit
```

**Key Structure Best Practices:**

```
Format: [environment]:[limit_type]:[target]:[identifier]:[window]

Examples:
  "prod:login:ip:192.168.1.1:60s"
  "staging:login:user:alice:15m"
  "prod:api:key:abc123:1h"
  "prod:combined:192.168.1.1:alice:5m"
```

**Key Components Explained:**

| Component | Purpose | Example |
|-----------|---------|---------|
| **Environment** | Isolate dev/staging/prod | `prod`, `staging`, `test` |
| **Limit Type** | Different rules for different operations | `login`, `api`, `register` |
| **Target** | What are we limiting? | `ip`, `user`, `api_key` |
| **Identifier** | The actual value | normalized IP, userID hash |
| **Window** | Auto-expiry | `60s`, `15m`, `1h` |

### 1.3. Composite Keys for Defense

**Why Single Key is Dangerous:**

```
Attack Scenario:
- Attacker controls botnet (1000 different IPs)
- Each IP makes 1 login attempt
- If you limit ONLY by IP: 1000 attempts get through
- If you limit ONLY by login: still 1000 attempts

Solution: COMBINED keys
```

**Three-Level Protection Strategy:**

```
Level 1: By IP
  Key: "login:ip:192.168.1.1"
  Limit: 5 attempts per minute
  Protects against: Single IP brute force

Level 2: By Login
  Key: "login:user:alice"
  Limit: 10 attempts per 15 minutes
  Protects against: Distributed botnet attack

Level 3: Combined
  Key: "login:combined:192.168.1.1:alice"
  Limit: 3 attempts per minute
  Protects against: Fast succession from same IP
```

---

## 2. Multi-layered Architecture

### 2.1. Why Multiple Layers?

```
One layer of protection is like one lock on a door.
Good lock, but if attacker picks it - everything is open.

Defense in depth is like a bank vault:
- Guard at entrance (CDN/WAF)
- Gate before the door (API Gateway)
- Safe with combination (Application)
- Alarm system (Monitoring)
```

### 2.2. Three Mandatory Layers

**Layer 1: Edge (CDN/WAF)**

```
Purpose: Stop obvious attacks before they reach your infrastructure

Capabilities:
  - Coarse IP limits (1000 requests/minute)
  - Country blocking
  - JS challenges for suspicious IPs
  - CAPTCHA on threshold exceed
  - DDoS mitigation

Implementation Examples:
  - Cloudflare WAF
  - AWS WAF
  - Google Cloud Armor
  - Akamai Kona

Configuration Example:
```

```yaml
edge_layer:
  rules:
    - name: "global_ip_limit"
      key: "${ip}"
      match: "/api/*"
      limits:
        - count: 1000
          period: "1m"
        - count: 10000
          period: "1h"
    
    - name: "login_protection"
      key: "${ip}"
      match: "/login"
      limits:
        - count: 10
          period: "1m"
      action: "challenge"  # CAPTCHA on exceed
```

**Layer 2: API Gateway / Ingress**

```
Purpose: Protect backend resources from exhaustion

Capabilities:
  - Fine-grained limits by endpoint
  - API key based limiting
  - HTTP method differentiation
  - Request queuing and throttling
  - Rate limiting before backend processing

Implementation Examples:
  - Kong
  - NGINX
  - Envoy
  - Traefik
  - AWS API Gateway

Configuration Example:
```

```yaml
gateway_layer:
  limits:
    - name: "basic_api"
      key: "${ip}"
      window: "60s"
      limit: 1000
      paths: ["/api/*"]
    
    - name: "auth_endpoints"
      key: "${ip}"
      window: "60s"
      limit: 30
      paths: ["/login", "/register", "/reset-password"]
    
    - name: "api_key_limits"
      key: "${api_key}"
      window: "1h"
      limit: 10000
      plan_mapping:
        free: 1000
        pro: 10000
        enterprise: 100000
```

**Layer 3: Application**

```
Purpose: Business-aware protection

Capabilities:
  - Context-aware limits (IP + login combinations)
  - Temporary account blocking
  - Anomaly detection
  - Graduated responses
  - Business logic integration

Implementation Examples:
  - Custom middleware
  - Redis-based limiters
  - In-memory caches

Configuration Example:
```

```yaml
application_layer:
  rules:
    - name: "login_ip"
      endpoints: ["/login"]
      key_pattern: "login:ip:{ip}"
      window: "60s"
      limit: 5
      algorithm: "sliding_window_log"
    
    - name: "login_user"
      endpoints: ["/login"]
      key_pattern: "login:user:{login}"
      window: "900s"  # 15 minutes
      limit: 10
      algorithm: "sliding_window_log"
    
    - name: "combined"
      endpoints: ["/login"]
      key_pattern: "login:combined:{ip}:{login}"
      window: "60s"
      limit: 3
      algorithm: "sliding_window_log"
```

---

## 3. Algorithms and Data Structures

### 3.1. Algorithm Comparison

| Algorithm | Use Case | Pros | Cons | Memory | Precision |
|-----------|----------|------|------|--------|-----------|
| **Fixed Window** | Coarse limits, monitoring | Simplest, minimal overhead | Boundary problem | Low | Low |
| **Sliding Window Log** | Anti-bruteforce, critical ops | Exact precision | Stores all timestamps | High | Exact |
| **Sliding Window Counter** | High-load APIs | Good balance | Approximation | Medium | ~98% |
| **Token Bucket** | Bursty traffic, shaping | Supports bursts | Complex parameters | Low | N/A |
| **Leaky Bucket** | Stable outflow | Predictable | Not flexible | Low | N/A |

### 3.2. Fixed Window (Simple but Flawed)

**Concept:**
```
Time divided into fixed windows.
Counter resets at window boundaries.
```

**Implementation Pseudocode:**
```
function checkFixedWindow(key, windowSec, limit):
    now = currentTimeSeconds()
    window = floor(now / windowSec)
    windowKey = key + ":" + window
    
    current = increment(windowKey)
    setExpiry(windowKey, windowSec)
    
    if current > limit:
        return {allowed: false, current: current}
    return {allowed: true, current: current}
```

**The Boundary Problem:**
```
Window 1: 00:00 - 00:01 (limit 100)
Window 2: 00:01 - 00:02 (limit 100)

At 00:00:59: 100 requests (window 1 full)
At 00:01:01: 100 requests (window 2 fresh)

Reality: 200 requests in 2 seconds
Limit should have triggered, but didn't!
```

### 3.3. Sliding Window Log (Most Precise)

**Concept:**
```
Store timestamp of each request.
Count only requests within current window.
```

**Data Structure:**
```
Sorted Set in Redis:
  key: "ratelimit:login:ip:192.168.1.1"
  members: timestamp values
  scores: same timestamps for sorting
```

**Implementation Pseudocode:**
```
function checkSlidingWindowLog(key, windowSec, limit):
    now = currentTimeMillis()
    windowStart = now - (windowSec * 1000)
    
    # Remove old entries
    deleteWhereScoreLessThan(key, windowStart)
    
    # Count current entries
    current = countMembers(key)
    
    if current >= limit:
        # Get oldest entry for retry-after
        oldest = getMemberWithMinScore(key)
        retryAfter = ceil((oldest.score + windowSec*1000 - now) / 1000)
        
        return {
            allowed: false,
            current: current,
            limit: limit,
            retryAfter: retryAfter
        }
    
    # Add new request
    addMember(key, now, generateUniqueID())
    setExpiry(key, windowSec * 2)
    
    return {
        allowed: true,
        current: current + 1,
        limit: limit,
        remaining: limit - (current + 1)
    }
```

### 3.4. Sliding Window Counter (Optimized)

**Concept:**
```
Keep two counters: current window and previous window.
Weight = time in current window * current counter +
         remaining from previous * previous counter
```

**Implementation Pseudocode:**
```
function checkSlidingWindowCounter(key, windowSec, limit):
    now = currentTimeSeconds()
    currentWindow = floor(now / windowSec)
    previousWindow = currentWindow - 1
    
    # Get weights
    windowProgress = (now % windowSec) / windowSec  # 0 to 1
    
    currentCount = getCounter(key, currentWindow)
    previousCount = getCounter(key, previousWindow)
    
    # Calculate weighted count
    estimated = (previousCount * (1 - windowProgress)) + currentCount
    
    if estimated >= limit:
        return {allowed: false}
    
    # Increment current window
    incrementCounter(key, currentWindow)
    setExpiry(key, windowSec * 2)
    
    return {allowed: true}
```

### 3.5. Token Bucket (For Bursty Traffic)

**Concept:**
```
Bucket holds N tokens.
Tokens added at rate R per second.
Each request consumes one token.
If no tokens - request rejected.
```

**Implementation Pseudocode:**
```
function checkTokenBucket(key, capacity, refillRate):
    data = get(key) or {tokens: capacity, lastRefill: now()}
    
    # Refill tokens based on time passed
    now = currentTimeSeconds()
    timePassed = now - data.lastRefill
    newTokens = timePassed * refillRate
    data.tokens = min(capacity, data.tokens + newTokens)
    data.lastRefill = now
    
    if data.tokens >= 1:
        data.tokens -= 1
        save(key, data)
        return {allowed: true, tokens: data.tokens}
    
    save(key, data)
    return {allowed: false, tokens: 0}
```

### 3.6. Best Practice: Atomic Operations

**The Problem with Non-Atomic Operations:**

```
Time | Request A           | Request B           | Storage
-----|---------------------|---------------------|--------
t1   | current = GET(key)  |                     | key=99
t2   |                     | current = GET(key)  | key=99
t3   | if current < limit  |                     |
t4   | INCR(key)           |                     | key=100
t5   |                     | if current < limit  | (sees 99!)
t6   |                     | INCR(key)           | key=101

Result: Limit exceeded, but both requests allowed!
```

**The Solution: Atomic Operations**

```
function atomicCheck(key, windowSec, limit):
    # This entire function executes ATOMICALLY
    # No other request can interfere
    
    now = currentTimeMillis()
    windowStart = now - (windowSec * 1000)
    
    # Delete old entries
    deleteWhereScoreLessThan(key, windowStart)
    
    # Count and check
    current = countMembers(key)
    
    if current < limit:
        addMember(key, now, uniqueID())
        setExpiry(key, windowSec * 2)
        return {allowed: true, current: current + 1}
    else:
        oldest = getMemberWithMinScore(key)
        retryAfter = ceil((oldest.score + windowSec*1000 - now) / 1000)
        return {allowed: false, current: current, retryAfter: retryAfter}
```

**How to Achieve Atomicity:**

| Storage | Mechanism | Example |
|---------|-----------|---------|
| **Redis** | Lua scripts | `redis.call('ZADD', ...)` |
| **MySQL** | Stored procedures | `BEGIN ATOMIC ... END` |
| **PostgreSQL** | Advisory locks + functions | `pg_try_advisory_lock()` |
| **Memcached** | CAS (Check-And-Set) | `gets` + `cas` operations |

---

## 4. Business Logic and Reactions

### 4.1. Graduated Responses

**One Level is Not Enough:**

```
Fixed Response:
  Limit exceeded → 429 Too Many Requests

Problem:
  - First time offender gets same response as persistent attacker
  - No deterrent for botnets
  - No escalation for repeated violations
```

**Graduated Response System:**

```
Level 1: First Exceed
  Response: 429 with short Retry-After (30 seconds)
  Action: Log, increment counter
  
Level 2: Repeated Exceed (3+ times in 1 hour)
  Response: 429 with longer Retry-After (15 minutes)
  Action: Log with warning, notify security?
  
Level 3: Persistent Attack (10+ times in 1 hour)
  Response: 403 Account Temporarily Locked
  Action: Require CAPTCHA or email verification
  Log: Security incident
  
Level 4: Extreme (50+ times in 1 hour)
  Response: 403 Account Locked (24h)
  Action: Notify security team, block all attempts
  Log: Critical security event
```

**Implementation Pseudocode:**

```
function checkWithGraduation(ip, login):
    # Check basic limits
    ipResult = checkLimit("ip:" + ip)
    loginResult = checkLimit("user:" + login)
    combinedResult = checkLimit("combined:" + ip + ":" + login)
    
    if not ipResult.allowed or not loginResult.allowed:
        # Get violation count for this login
        violationCount = getViolationCount(login, "1h")
        
        if violationCount > 50:
            return {
                allowed: false,
                status: 403,
                message: "Account locked for 24 hours",
                retryAfter: 86400
            }
        elif violationCount > 10:
            return {
                allowed: false,
                status: 403,
                message: "Account temporarily locked. Verify email.",
                requireVerification: true
            }
        elif violationCount > 3:
            return {
                allowed: false,
                status: 429,
                retryAfter: 900,  # 15 minutes
                message: "Too many attempts. Try again later."
            }
        else:
            return {
                allowed: false,
                status: 429,
                retryAfter: 30,
                message: "Too many attempts."
            }
    
    return {allowed: true}
```

### 4.2. Standard HTTP Responses

**Response Headers (RFC 6585):**

```
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640995200
Content-Type: application/json
```

**Response Body:**

```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many login attempts. Please try again in 30 seconds.",
  "retry_after": 30,
  "limit": 100,
  "remaining": 0,
  "reset": 1640995200,
  "request_id": "req-123-abc"
}
```

**Retry-After Formats:**

```
# Seconds (preferred)
Retry-After: 30

# HTTP Date (RFC 1123)
Retry-After: Wed, 21 Oct 2015 07:28:00 GMT
```

### 4.3. Client-Side Best Practices

**What Frontend Should Do:**

```
On receiving 429:
  1. Parse Retry-After header
  2. Disable submit buttons
  3. Show countdown timer to user
  4. Queue any automatic retries
  5. Respect the retry time (don't retry early)
```

**What Frontend Should NOT Do:**

```
❌ Don't implement your own rate limiting
❌ Don't retry immediately on 429
❌ Don't ignore Retry-After
❌ Don't let users bypass by refreshing
```

---

## 5. Observability and Monitoring

### 5.1. Metrics to Collect

**For Developers:**

```
rate_limit_checks_total
  dimensions: limiter_type, endpoint, result (allow/deny)
  
rate_limit_exceeded_total
  dimensions: limiter_type, endpoint, severity
  
rate_limit_check_duration_seconds
  dimensions: limiter_type
  aggregations: p95, p99, max
```

**For DevOps:**

```
rate_limit_storage_operations_total
  dimensions: storage_type (redis/mysql), operation (read/write)
  
rate_limit_storage_latency_seconds
  dimensions: storage_type
  
rate_limit_active_keys
  dimensions: limiter_type
  
rate_limit_memory_usage_bytes
  dimensions: storage_type
```

**For Security:**

```
rate_limit_top_offenders_ip
rate_limit_top_offenders_login
rate_limit_attack_patterns
rate_limit_geographic_distribution
```

### 5.2. Structured Logging

**Every Rate Limit Event Should Log:**

```json
{
  "@timestamp": "2024-01-15T10:30:00Z",
  "event_type": "rate_limit_exceeded",
  "severity": "warning",
  
  "request_context": {
    "client_ip": "192.168.1.1",
    "login": "alice@example.com",
    "endpoint": "/api/v1/login",
    "method": "POST",
    "user_agent": "Mozilla/5.0..."
  },
  
  "rate_limit": {
    "type": "login_ip",
    "key": "ratelimit:login:ip:192.168.1.1",
    "limit": 5,
    "current": 6,
    "window_seconds": 60
  },
  
  "action_taken": {
    "status_code": 429,
    "retry_after": 30,
    "blocked": true,
    "severity_level": 2
  },
  
  "environment": "production",
  "service": "auth-service",
  "request_id": "req-abc-123-def"
}
```

### 5.3. Alerting Rules

**Warning Alerts (PagerDuty/Email):**

```yaml
- name: "High Rate Limit Exceeded Rate"
  condition: "rate(rate_limit_exceeded_total[5m]) > 100"
  severity: "warning"
  description: "Unusual number of rate limit violations"
  
- name: "Storage Latency"
  condition: "p95(rate_limit_storage_latency_seconds) > 0.1"
  severity: "warning"
  description: "Rate limiter slowing down requests"
```

**Critical Alerts (Immediate Action):**

```yaml
- name: "Possible DDoS Attack"
  condition: "rate(rate_limit_exceeded_total[1m]) > 1000"
  severity: "critical"
  description: "Potential DDoS in progress"
  
- name: "Multiple Account Lockouts"
  condition: "increase(account_lockouts_total[10m]) > 50"
  severity: "critical"
  description: "Possible distributed attack on multiple accounts"
  
- name: "Storage Failure"
  condition: "rate_limit_storage_errors_total > 0"
  severity: "critical"
  description: "Rate limiter storage unavailable"
```

---

## 6. Configuration and Feature Flags

### 6.1. Configuration Structure

```yaml
rate_limiter:
  # Global settings
  enabled: true
  default_strategy: "sliding_window_log"
  key_prefix: "prod:ratelimit"
  
  # Storage configuration
  storage:
    type: "redis"  # redis, mysql, memory
    connection:
      addresses: ["redis-1:6379", "redis-2:6379"]
      cluster_mode: true
      max_retries: 3
      pool_size: 100
    fallback_strategy: "block"  # block or allow on failure
  
  # IP extraction
  ip_source:
    trust_proxies: true
    trusted_cidrs:
      - "10.0.0.0/8"
      - "172.16.0.0/12"
      - "192.168.0.0/16"
      - "100.64.0.0/10"  # AWS private
    headers_priority:
      - "X-Forwarded-For"
      - "X-Real-IP"
      - "CF-Connecting-IP"  # Cloudflare
    strip_port: true
    normalize_ipv6: true
  
  # Rate limit rules
  rules:
    - name: "login_ip"
      endpoints: 
        - "/login"
        - "/api/v1/login"
        - "/auth"
      key_pattern: "{ip}"
      window: "60s"
      limit: 5
      algorithm: "sliding_window_log"
      severity: "medium"
      on_exceed:
        - action: "rate_limit"
          status: 429
          retry_after: 30
    
    - name: "login_user"
      endpoints: ["/login"]
      key_pattern: "user:{login}"
      window: "900s"  # 15 minutes
      limit: 10
      algorithm: "sliding_window_log"
      severity: "high"
      on_exceed:
        - action: "temporary_block"
          duration: "3600s"  # 1 hour
        - action: "require_captcha"
    
    - name: "api_general"
      endpoints: ["/api/*"]
      key_pattern: "{api_key}"
      window: "3600s"  # 1 hour
      limit: 1000
      burst: 100
      algorithm: "token_bucket"
      plan_mapping:
        free: 1000
        pro: 10000
        enterprise: 100000
```

### 6.2. Feature Flags for Safe Rollout

```yaml
feature_flags:
  rate_limiter_v2:
    enabled: true
    rollout_percentage: 10  # Canary deployment
    rules:
      - if: "user_id % 100 < 10"  # 10% of users
        use: "new_algorithm"
      - else:
        use: "legacy_algorithm"
  
  strict_mode:
    enabled: false
    rules:
      - if: "ip in canary_ips"
        use: "stricter_limits"
      - if: "environment == 'staging'"
        use: "test_limits"
  
  dynamic_limits:
    enabled: true
    refresh_interval: "60s"
    source: "config_server"  # or database, feature flag service
```

---

## 7. Testing and Validation

### 7.1. Functional Tests

**Test Case 1: Basic Limit**

```
Scenario: Single IP makes requests within limit

Steps:
  1. Make 5 requests from IP 192.168.1.1 to /login
  2. Verify all return 200 OK
  3. Make 6th request
  4. Verify returns 429 Too Many Requests
  5. Verify Retry-After header present
```

**Test Case 2: Different IPs Don't Interfere**

```
Scenario: Multiple IPs should have separate counters

Steps:
  1. Make 5 requests from IP A to /login
  2. Make 5 requests from IP B to /login
  3. Verify both IPs get 200 OK
  4. Make 6th request from IP A
  5. Verify IP A gets 429
  6. Verify IP B can still make requests
```

**Test Case 3: Distributed Attack Protection**

```
Scenario: Botnet attacking one login

Steps:
  1. Configure: login limit = 10 per 15 minutes
  2. From 20 different IPs, try to login as "admin"
  3. Make 1 attempt from each IP
  4. After 10 attempts, verify account is locked
  5. 11th attempt returns 403, not 429
```

### 7.2. Load Tests

**Test Case: Burst Traffic**

```
Scenario: 1000 simultaneous requests from same IP

Setup:
  - Limit: 100 requests per minute
  - Tool: 1000 concurrent goroutines/threads
  
Assertions:
  - Exactly 100 requests allowed
  - 900 requests rejected with 429
  - No race conditions (should never allow >100)
```

**Test Case: Sustained Load**

```
Scenario: Constant 50 RPS for 10 minutes

Setup:
  - Monitor Redis memory usage
  - Track p95 latency
  - Check for key leaks (expiry working)
  
Assertions:
  - Memory usage stable (no leaks)
  - Latency < 50ms p95
  - No errors from storage
```

### 7.3. Security Tests

**Test Case: Port Bypass Attempt**

```
Scenario: Attacker tries different ports

Steps:
  1. Make requests from 192.168.1.1:8080 to /login
  2. Make requests from 192.168.1.1:8081 to /login
  3. Make requests from 192.168.1.1:8082 to /login
  
Expected: All count toward same limit
Assert: After 5 total attempts, 429 returned
```

**Test Case: Header Spoofing**

```
Scenario: Client tries to spoof X-Forwarded-For

Steps:
  1. Send request with header "X-Forwarded-For: 1.2.3.4"
  2. Real IP is 192.168.1.1 from untrusted network
  
Expected: System IGNORES the header
Assert: Rate limit based on 192.168.1.1, not 1.2.3.4
```

**Test Case: Timing Attacks**

```
Scenario: Check if timing reveals information

Steps:
  1. Measure response time when limit not reached
  2. Measure response time when limit reached
  3. Compare distributions
  
Expected: No significant difference (<10ms)
Assert: Attacker cannot distinguish states by timing
```

---

## 8. Implementation Checklist

### 🔴 CRITICAL (Must Have)

**Client Identification**
- [ ] IP extracted considering trusted proxies only
- [ ] Port stripped from IP addresses
- [ ] Composite keys (IP + login) for login protection
- [ ] Keys never include client-supplied headers directly

**Storage Operations**
- [ ] All rate limit checks use atomic operations
- [ ] Race conditions impossible by design
- [ ] Fallback strategy defined (prefer blocking on failure)
- [ ] Storage connection pooling configured

**HTTP Interface**
- [ ] 429 status code returned on limit exceed
- [ ] Retry-After header always included
- [ ] X-RateLimit-* headers present
- [ ] JSON error response with request ID

**Observability**
- [ ] Every exceed event logged with context
- [ ] Prometheus metrics exposed
- [ ] Request ID propagated through logs
- [ ] Storage latency monitored

### 🟡 IMPORTANT (Should Have)

**Business Logic**
- [ ] Graduated responses based on violation count
- [ ] Different limits for different endpoints
- [ ] Temporary account blocking for repeated violations
- [ ] CAPTCHA/verification triggers

**Configuration**
- [ ] Limits configurable without code change
- [ ] Feature flags for gradual rollout
- [ ] Environment isolation (dev/staging/prod)
- [ ] Dynamic limit adjustment capability

**Testing**
- [ ] Integration tests for all limit types
- [ ] Load tests under burst conditions
- [ ] Security tests for bypass attempts
- [ ] Chaos tests (storage failure simulation)

**Documentation**
- [ ] API documentation with rate limit headers
- [ ] Configuration guide for operations
- [ ] Troubleshooting guide for support
- [ ] Runbook for rate limit incidents

### 🟢 NICE TO HAVE (Enterprise Level)

**Advanced Features**
- [ ] Distributed rate limiting (consistent hashing)
- [ ] Machine learning anomaly detection
- [ ] Automatic botnet pattern recognition
- [ ] Geographic distribution tracking

**Optimizations**
- [ ] Local caching for hot keys
- [ ] Batch processing for high throughput
- [ ] Adaptive limits based on system load
- [ ] Self-tuning parameters

**Compliance**
- [ ] Audit trail of all limit changes
- [ ] GDPR compliance (PII in logs handled)
- [ ] SOC2 compliance documentation
- [ ] Penetration test results

---

## Appendix: Quick Reference

### Common Rate Limit Scenarios

| Scenario | Key Pattern | Window | Limit | Algorithm |
|----------|-------------|--------|-------|-----------|
| Login by IP | `login:ip:{ip}` | 60s | 5 | Sliding Window Log |
| Login by user | `login:user:{login}` | 15m | 10 | Sliding Window Log |
| Registration | `register:ip:{ip}` | 1h | 3 | Sliding Window Log |
| Password reset | `reset:ip:{ip}` | 1h | 3 | Sliding Window Log |
| API (free tier) | `api:key:{key}` | 1h | 1000 | Token Bucket |
| API (pro) | `api:key:{key}` | 1h | 10000 | Token Bucket |
| Search endpoint | `search:ip:{ip}` | 1m | 30 | Sliding Window Counter |
| File upload | `upload:user:{id}` | 1h | 10 | Fixed Window |

### HTTP Status Codes for Rate Limiting

| Code | Meaning | When to Use |
|------|---------|-------------|
| **429** | Too Many Requests | Standard rate limit exceed |
| **403** | Forbidden | Account locked, severe violation |
| **401** | Unauthorized | With 2FA requirement |
| **503** | Service Unavailable | System overload, circuit breaker |

### Common Headers

| Header | Purpose | Example |
|--------|---------|---------|
| `Retry-After` | When to retry | `30` or `Wed, 21 Oct 2015 07:28:00 GMT` |
| `X-RateLimit-Limit` | Maximum requests | `100` |
| `X-rateLimit-Remaining` | Remaining in window | `45` |
| `X-RateLimit-Reset` | Window reset timestamp | `1640995200` |

---

*This documentation is living. If you discover new attack vectors or better practices, update it.*

---

**Document Version:** 1.0.0
**Last Updated:** March 2026
**Classification:** Internal / Public
