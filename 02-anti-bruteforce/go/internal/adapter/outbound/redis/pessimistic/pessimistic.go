package redisadapter

import (
	"context"
	"fmt"
	"time"

	"anti-bruteforce/internal/domain"

	"github.com/redis/go-redis/v9"
)

// Maybe you can't understand is a lot of comments but learn have value forget and to return later and to remember what I did

// PessimisticLimiter implements rate limiting using pessimistic locks in Redis
// Core idea: "lock the resource, do the work, release the lock"
type PessimisticLimiter struct {
	client *redis.Client
}

func NewPessimisticLimiter(client *redis.Client) *PessimisticLimiter {
	return &PessimisticLimiter{client: client}
}

// Allow checks if request from this IP can be processed
// Returns:
//   - true, nil — request allowed
//   - false, nil — request rejected (rate limit exceeded)
//   - false, error — technical error (Redis down, etc)
func (p *PessimisticLimiter) Allow(ip string, limit int, windowSec int64) (bool, error) {
	// Context for Redis operations. Empty here, but in production you'd add timeout
	// Контекст для операций с Redis. В этом примере просто пустой,
	// но в реальном проекте сюда можно добавить таймаут: context.WithTimeout
	ctx := context.Background()

	// ---------- PART 1: ACQUIRE LOCK ----------
	// ЧАСТЬ 1: СОЗДАЕМ ЗАМОК
	// We'll have two keys in Redis:
	// 1. lock:rate:ip — the lock itself (flag "someone is editing")
	// 2. rate:ip — actual data (request history)
	// У нас будет два ключа в Redis:
	// 1. lock:rate:ip — сам замок (флаг "идет редактирование")
	// 2. rate:ip — данные (история запросов)

	// Lock key — used to block access to data
	// Ключ для замка — по нему мы будем блокировать доступ к данным
	lockKey := fmt.Sprintf("lock:rate:%s", ip)

	// Unique lock value. Not used for verification in this example,
	// but in proper implementation you MUST check you're deleting YOUR lock
	// Уникальное значение замка. В этом примере не используется для проверки,
	// но в правильной реализации нужно проверять при снятии, что снимаешь СВОЙ замок
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	const (
		lockTTL = time.Second // Lock TTL. If program crashes, lock auto-expires after 1 sec
		// Время жизни замка. Если программа упадет, замок сам исчезнет через 1 сек
		maxLockRetries = 3 // How many times to try acquiring lock
		// Сколько раз пробовать взять замок
		lockRetryDelay = 50 * time.Millisecond // Pause between retries
		// Пауза между попытками
	)

	// Try to acquire lock (up to 3 times)
	// Пытаемся взять замок (до 3 раз)
	acquired := false
	for i := 0; i < maxLockRetries; i++ {
		// REAL REDIS COMMAND:
		// SET lock:rate:192.168.1.1 167890123456789 NX EX 1
		//
		// This is Redis command: SET lock:rate:ip value NX EX 1
		// NX — set only if key DOES NOT EXIST (lock is free)
		// EX 1 — set TTL 1 second (protection against dead locks)
		//
		// Возможные ответы Redis:
		// - "OK" — замок успешно установлен
		// - (nil) — ключ уже существует (замок занят)
		//
		// Это Redis команда: SET lock:rate:ip значение NX EX 1
		// NX — установить только если ключа НЕТ (то есть замок свободен)
		// EX 1 — установить время жизни 1 секунда (защита от мертвых замков)
		res, err := p.client.SetArgs(ctx, lockKey, lockValue, redis.SetArgs{
			Mode: "NX",    // Set if Not eXists
			TTL:  lockTTL, // time to live
		}).Result()

		// Check real Redis errors (connection issues, etc)
		// Проверяем реальные ошибки Redis (соединение, etc)
		if err != nil && err != redis.Nil {
			return false, err
		}

		// If Redis returned "OK" — we successfully acquired the lock!
		// Если вернулось "OK" — мы успешно взяли замок!
		if res == "OK" {
			acquired = true
			break
		}

		// If lock not acquired — someone else holds it, wait and retry
		// Если не вышло — замок занят другим процессом, ждем и пробуем снова
		time.Sleep(lockRetryDelay)
	}

	// If we failed to acquire lock after all retries — return error
	// Если так и не смогли взять замок после всех попыток — возвращаем ошибку
	if !acquired {
		return false, fmt.Errorf("failed to acquire lock for ip %s", ip)
	}

	// Guarantee lock release when function exits
	// defer executes even if function panics
	// Гарантированно снимаем замок при выходе из функции
	// defer выполняется даже если функция упадет с паникой
	defer func() {
		// REAL REDIS COMMAND:
		// DEL lock:rate:192.168.1.1
		//
		// Delete lock key. Ignore errors — if delete failed,
		// lock will auto-expire after TTL=1 second
		//
		// Возможные ответы Redis:
		// - (integer) 1 — ключ существовал и удален
		// - (integer) 0 — ключа уже не было
		//
		// Удаляем ключ-замок. Ошибки игнорируем — если не получилось удалить,
		// замок сам истечет через TTL=1 секунду
		_, _ = p.client.Del(ctx, lockKey).Result()
	}()

	// ---------- PART 2: WORK WITH DATA (UNDER LOCK) ----------
	// Now we hold the lock — other requests for this IP are retrying in the loop
	// We can safely read and write data
	// ЧАСТЬ 2: РАБОТАЕМ С ДАННЫМИ (ПОД ЗАМКОМ)
	// Раз мы взяли замок — другие запросы для этого IP висят в цикле с ретраями
	// Теперь можно спокойно читать и писать данные

	now := time.Now()
	// Window boundary: now minus windowSec seconds
	// Example: now 10:30:00, windowSec=60 → cutoff = 10:29:00
	// Граница окна: сейчас минус windowSec секунд
	// Например, сейчас 10:30:00, windowSec=60 → cutoff = 10:29:00
	cutoff := now.Add(-time.Duration(windowSec) * time.Second).Unix()

	// Data key — stores request history for this IP
	// Ключ для данных — здесь хранится история запросов от этого IP
	key := fmt.Sprintf("rate:%s", ip)

	// 1. Clean up garbage — remove requests older than window
	// REAL REDIS COMMAND:
	// ZREMRANGEBYSCORE rate:192.168.1.1 -inf 1678901234
	// (where 1678901234 is cutoff-1)
	//
	// ZRemRangeByScore removes elements with score from -inf to cutoff-1
	// cutoff-1 because cutoff is inclusive, we need strictly less than cutoff
	//
	// Возвращает: количество удаленных элементов
	//
	// Чистим мусор — удаляем запросы старше окна
	// ZRemRangeByScore удаляет элементы с score от -inf до cutoff-1
	// cutoff-1 потому что cutoff включительно, а нам нужны строго меньше cutoff
	_, err := p.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff-1)).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// 2. Count requests in current window (from cutoff to now)
	// REAL REDIS COMMAND:
	// ZCOUNT rate:192.168.1.1 1678901235 1678901295
	// (where 1678901235 is cutoff, 1678901295 is now)
	//
	// ZCount returns number of elements with score in range
	//
	// Возвращает: количество элементов в диапазоне
	//
	// Считаем запросы в текущем окне (от cutoff до now)
	// ZCount возвращает количество элементов с score в диапазоне
	count, err := p.client.ZCount(ctx, key, fmt.Sprintf("%d", cutoff), fmt.Sprintf("%d", now.Unix())).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// 3. Check rate limit
	// Проверяем лимит
	if count >= int64(limit) {
		// Limit exceeded — reject request
		// Lock will be released via defer
		// Лимит исчерпан — возвращаем false (отклоняем запрос)
		// Замок автоматически снимется через defer
		return false, nil
	}

	// 4. Add current request to history
	// REAL REDIS COMMAND:
	// ZADD rate:192.168.1.1 1678901295 1678901295123456789
	// (where 1678901295 is timestamp, 1678901295123456789 is nanosecond timestamp)
	//
	// In sorted set, each element is a request
	// score — request timestamp (for counting and cleanup)
	// member — unique request ID (allows storing multiple requests with same timestamp)
	//
	// Возвращает: количество добавленных элементов (1 или 0)
	//
	// Добавляем текущий запрос в историю
	// В sorted set каждый элемент — это запрос
	// score — время запроса (чтобы потом считать и чистить)
	// member — уникальный идентификатор запроса (чтобы можно было хранить много запросов с одинаковым временем)
	score := float64(now.Unix())
	member := fmt.Sprintf("%d", now.UnixNano()) // nanoseconds guarantee uniqueness
	// наносекунды гарантируют уникальность

	_, err = p.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	if err != nil {
		return false, err
	}

	// ALL OPERATIONS SUCCESSFUL
	// Request allowed, lock will be released via defer
	// ВСЕ ОПЕРАЦИИ ВЫПОЛНЕНЫ УСПЕШНО
	// Запрос разрешен, замок скоро снимется через defer
	return true, nil
}

// Compile-time check that PessimisticLimiter implements domain.RateLimiter
// Проверка реализации интерфейса на этапе компиляции
var _ domain.RateLimiter = (*PessimisticLimiter)(nil)
