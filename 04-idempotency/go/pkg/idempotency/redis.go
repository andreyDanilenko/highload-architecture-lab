package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProvider хранит общий для экземпляров ключ с SET NX и TTL.
// После истечения pending TTL новый владелец может повторить внешний эффект:
// это намеренное ограничение базового опыта, а не гарантия exactly-once.
type RedisProvider struct {
	client *redis.Client
	prefix string
}

func NewRedisProvider(client *redis.Client, prefix string) *RedisProvider {
	return &RedisProvider{client: client, prefix: prefix}
}

func (p *RedisProvider) Name() string { return "redis" }

// redisRecord не содержит исходный запрос: только fingerprint, владение и ответ.
// Ответ хранится отдельной JSON-строкой, чтобы Lua не менял байты/форму заголовков.
type redisRecord struct {
	Fingerprint  string `json:"fingerprint"`
	Owner        string `json:"owner"`
	Status       string `json:"status"`
	LeaseUntilMS int64  `json:"lease_until_ms,omitempty"`
	ResponseJSON string `json:"response_json,omitempty"`
}

func (p *RedisProvider) Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (Claim, error) {
	if err := ctx.Err(); err != nil {
		return Claim{}, err
	}
	if key == "" || fingerprint == "" || lease <= 0 {
		return Claim{}, ErrInvalidArgument
	}
	if p.client == nil {
		return Claim{}, ErrUnavailable
	}
	leaseMS, err := redisTTL(lease)
	if err != nil {
		return Claim{}, err
	}
	owner, err := newOwner()
	if err != nil {
		return Claim{}, fmt.Errorf("%w: create owner", ErrUnavailable)
	}
	record := redisRecord{Fingerprint: fingerprint, Owner: owner, Status: "pending"}
	data, err := json.Marshal(record)
	if err != nil {
		return Claim{}, ErrUnavailable
	}
	storageKey := p.prefix + key

	// Шаг 1: TTL ограничивает владение, но также удаляет сведения о незавершённой работе.
	acquired, err := p.client.SetNX(ctx, storageKey, data, time.Duration(leaseMS)*time.Millisecond).Result()
	if err != nil {
		return Claim{}, redisFailure(err)
	}
	if acquired {
		return Claim{Owner: owner}, nil
	}

	// Шаг 2: исчезновение между SET NX и GET оставляет исход неизвестным;
	// в этом вызове новую попытку выполнения не запускаем.
	data, err = p.client.Get(ctx, storageKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return Claim{}, ErrOutcomeUnknown
	}
	if err != nil {
		return Claim{}, redisFailure(err)
	}
	if err := json.Unmarshal(data, &record); err != nil {
		return Claim{}, ErrUnavailable
	}
	return claimFromRedis(record, fingerprint)
}

func (p *RedisProvider) Complete(ctx context.Context, key, owner string, response Response, retention time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if retention <= 0 {
		return ErrInvalidArgument
	}
	if p.client == nil {
		return ErrUnavailable
	}
	retentionMS, err := redisTTL(retention)
	if err != nil {
		return err
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return ErrUnavailable
	}
	storageKey := p.prefix + key

	// Шаг 3: WATCH связывает проверку владельца с записью результата. В современном
	// Redis истечение TTL тоже отменяет EXEC: старый worker не перезапишет новый ключ.
	err = p.client.Watch(ctx, func(tx *redis.Tx) error {
		data, err := tx.Get(ctx, storageKey).Bytes()
		if errors.Is(err, redis.Nil) {
			return ErrNotOwner
		}
		if err != nil {
			return redisFailure(err)
		}
		var record redisRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return ErrUnavailable
		}
		if owner == "" || record.Owner != owner || record.Status != "pending" {
			return ErrNotOwner
		}
		ttl, err := tx.PTTL(ctx, storageKey).Result()
		if err != nil {
			return redisFailure(err)
		}
		if ttl <= 0 {
			return ErrNotOwner
		}
		record.Status = "completed"
		record.ResponseJSON = string(responseJSON)
		data, err = json.Marshal(record)
		if err != nil {
			return ErrUnavailable
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, storageKey, data, time.Duration(retentionMS)*time.Millisecond)
			return nil
		})
		return err
	}, storageKey)
	if errors.Is(err, redis.TxFailedErr) {
		return ErrNotOwner
	}
	return redisFailure(err)
}

func claimFromRedis(record redisRecord, fingerprint string) (Claim, error) {
	if record.Fingerprint != fingerprint {
		return Claim{}, ErrMismatch
	}
	switch record.Status {
	case "pending":
		return Claim{}, ErrInProgress
	case "completed":
		return replayFromJSON(record.ResponseJSON)
	default:
		return Claim{}, ErrUnavailable
	}
}

func replayFromJSON(data string) (Claim, error) {
	var response Response
	if err := json.Unmarshal([]byte(data), &response); err != nil || response.Status < 200 || response.Status > 999 {
		return Claim{}, ErrUnavailable
	}
	return Claim{Response: &response}, nil
}

// redisTTL округляет вверх: положительный TTL не должен стать бессрочным SET.
func redisTTL(duration time.Duration) (int64, error) {
	if duration <= 0 {
		return 0, ErrInvalidArgument
	}
	ms := duration.Milliseconds()
	if duration%time.Millisecond != 0 {
		ms++
	}
	if ms > int64(time.Duration(1<<63-1)/time.Millisecond) {
		return 0, ErrInvalidArgument
	}
	return ms, nil
}

func redisFailure(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	if errors.Is(err, ErrUnavailable) || errors.Is(err, ErrNotOwner) || errors.Is(err, ErrOutcomeUnknown) {
		return err
	}
	// Не добавляем текст команды/ключа: ошибка хранилища не раскрывает данные запроса.
	return ErrUnavailable
}

var _ Provider = (*RedisProvider)(nil)
