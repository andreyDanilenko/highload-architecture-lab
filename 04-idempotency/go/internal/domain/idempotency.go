package domain

import (
	"encoding/json"
	"time"
)

type IdempotencyStatus string

const (
	StatusPending   IdempotencyStatus = "pending"
	StatusCompleted IdempotencyStatus = "completed"
	StatusFailed    IdempotencyStatus = "failed"
)

type IdempotencyRecord struct {
	Key         string            `json:"key"`
	Status      IdempotencyStatus `json:"status"`
	Result      *PaymentResult    `json:"result,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	LockedUntil time.Time         `json:"locked_until,omitempty"`
}

func NewIdempotencyRecord(key string, ttl time.Duration) *IdempotencyRecord {
	return &IdempotencyRecord{
		Key:         key,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		LockedUntil: time.Now().Add(ttl),
	}
}

func (r *IdempotencyRecord) IsLockExpired() bool {
	return time.Now().After(r.LockedUntil)
}

func (r *IdempotencyRecord) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalIdempotencyRecord(data []byte) (*IdempotencyRecord, error) {
	var record IdempotencyRecord
	err := json.Unmarshal(data, &record)
	return &record, err
}
