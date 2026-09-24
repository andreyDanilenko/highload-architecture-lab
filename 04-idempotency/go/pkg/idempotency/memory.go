package idempotency

import (
	"context"
	"sync"
	"time"
)

// MemoryProvider — subtask 2: map[key]record + TTL, только один процесс.
type MemoryProvider struct {
	mu      sync.RWMutex
	records map[string]*memoryEntry
	ttl     time.Duration
}

type memoryEntry struct {
	record    *Record
	expiresAt time.Time
}

func NewMemoryProvider(ttl time.Duration) *MemoryProvider {
	p := &MemoryProvider{
		records: make(map[string]*memoryEntry),
		ttl:     ttl,
	}
	go p.cleanupLoop()
	return p
}

func (p *MemoryProvider) Name() string { return "memory" }

func (p *MemoryProvider) TryLock(_ context.Context, key string, lockTTL time.Duration) (*Record, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.records[key]
	if !exists {
		return p.insertPending(key, lockTTL), nil
	}

	if entry.record.Status == StatusCompleted || entry.record.Status == StatusFailed {
		return entry.record, nil
	}

	if time.Now().After(entry.expiresAt) {
		return p.insertPending(key, lockTTL), nil
	}

	return entry.record, ErrKeyAlreadyProcessing
}

func (p *MemoryProvider) insertPending(key string, lockTTL time.Duration) *Record {
	record := NewRecord(key, lockTTL)
	p.records[key] = &memoryEntry{
		record:    record,
		expiresAt: time.Now().Add(p.ttl),
	}
	return record
}

func (p *MemoryProvider) Complete(_ context.Context, key string, result any, err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.records[key]
	if !exists {
		entry = &memoryEntry{record: &Record{Key: key, CreatedAt: time.Now()}}
		p.records[key] = entry
	}

	if err != nil {
		entry.record.Status = StatusFailed
		entry.record.Error = err.Error()
		entry.record.Result = nil
	} else {
		entry.record.Status = StatusCompleted
		entry.record.Result = result
		entry.record.Error = ""
	}
	entry.expiresAt = time.Now().Add(p.ttl)
	return nil
}

func (p *MemoryProvider) GetRecord(_ context.Context, key string) (*Record, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.records[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, ErrRecordNotFound
	}
	return entry.record, nil
}

func (p *MemoryProvider) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		p.cleanup()
	}
}

func (p *MemoryProvider) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for key, entry := range p.records {
		if now.After(entry.expiresAt) {
			delete(p.records, key)
		}
	}
}

var _ Provider = (*MemoryProvider)(nil)
