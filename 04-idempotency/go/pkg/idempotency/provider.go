package idempotency

import (
	"context"
	"encoding/json"
	"time"
)

// Status — жизненный цикл идемпотентной записи.
type Status string

const (
	StatusPending   Status = "pending"   // lock захвачен, операция выполняется
	StatusCompleted Status = "completed" // результат сохранён, replay без side-effect
	StatusFailed    Status = "failed"    // ошибка сохранена, replay той же ошибки
)

// Record — snapshot ключа: статус + результат + TTL lock.
type Record struct {
	Key         string    `json:"key"`
	Status      Status    `json:"status"`
	Result      any       `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LockedUntil time.Time `json:"locked_until,omitempty"`
}

func NewRecord(key string, lockTTL time.Duration) *Record {
	now := time.Now()
	return &Record{
		Key:         key,
		Status:      StatusPending,
		CreatedAt:   now,
		LockedUntil: now.Add(lockTTL),
	}
}

func (r *Record) IsLockExpired() bool {
	return time.Now().After(r.LockedUntil)
}

func (r *Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRecord(data []byte) (*Record, error) {
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// Provider — outbound port: где хранить ключи (memory, redis, postgres…).
type Provider interface {
	TryLock(ctx context.Context, key string, lockTTL time.Duration) (*Record, error)
	Complete(ctx context.Context, key string, result any, err error) error
	GetRecord(ctx context.Context, key string) (*Record, error)
	Name() string
}
