package idempotency

import (
	"context"
	"sync"
	"time"
)

// MemoryProvider защищает таблицу одним mutex. Владелец записи и бизнес-эффект —
// разные границы: после рестарта процесса таблица утрачивается целиком.
type MemoryProvider struct {
	mu         sync.Mutex
	records    map[string]memoryEntry
	maxEntries int
	now        func() time.Time
	nextSweep  time.Time
}

type memoryEntry struct {
	fingerprint string
	owner       string
	leaseUntil  time.Time
	response    *Response
	expiresAt   time.Time
}

// NewMemoryProvider ограничивает число записей. Не создаёт фоновых goroutines:
// завершённые записи убираются при обращении и периодическом проходе на Acquire.
// Неизвестные исходы сохраняются до разбора; при заполнении возвращается ErrCapacity.
func NewMemoryProvider(maxEntries int) *MemoryProvider {
	if maxEntries <= 0 {
		maxEntries = 10_000
	}
	return &MemoryProvider{records: make(map[string]memoryEntry), maxEntries: maxEntries, now: time.Now}
}

func (*MemoryProvider) Name() string { return "memory" }

func (p *MemoryProvider) Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (Claim, error) {
	if err := ctx.Err(); err != nil {
		return Claim{}, err
	}
	if key == "" || fingerprint == "" || lease <= 0 {
		return Claim{}, ErrInvalidArgument
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Claim{}, err
	}
	now := p.now()

	// Шаг 1: завершённый результат живёт ровно retention, независимо от уборки.
	if entry, ok := p.records[key]; ok {
		if entry.response != nil && !now.Before(entry.expiresAt) {
			delete(p.records, key)
		} else {
			if entry.fingerprint != fingerprint {
				return Claim{}, ErrMismatch
			}
			if entry.response != nil {
				response := cloneResponse(*entry.response)
				return Claim{Response: &response}, nil
			}
			// Expiry не останавливает старого исполнителя. Без выяснения эффекта повтор
			// не допускается: меняем доступность на защиту от автоматического дублирования.
			if !now.Before(entry.leaseUntil) {
				return Claim{}, ErrOutcomeUnknown
			}
			return Claim{}, ErrInProgress
		}
	}

	// Шаг 2: ограничиваем число ключей; pending нельзя вытеснять как обычный cache.
	if !now.Before(p.nextSweep) || len(p.records) >= p.maxEntries {
		for key, entry := range p.records {
			if entry.response != nil && !now.Before(entry.expiresAt) {
				delete(p.records, key)
			}
		}
		p.nextSweep = now.Add(time.Minute)
	}
	if len(p.records) >= p.maxEntries {
		return Claim{}, ErrCapacity
	}
	owner, err := newOwner()
	if err != nil {
		return Claim{}, err
	}
	p.records[key] = memoryEntry{fingerprint: fingerprint, owner: owner, leaseUntil: now.Add(lease)}
	return Claim{Owner: owner}, nil
}

func (p *MemoryProvider) Complete(ctx context.Context, key, owner string, response Response, retention time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if retention <= 0 {
		return ErrInvalidArgument
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	entry, ok := p.records[key]
	if !ok || entry.owner != owner || entry.response != nil {
		return ErrNotOwner
	}
	now := p.now()
	if !now.Before(entry.leaseUntil) {
		return ErrOutcomeUnknown
	}

	// Шаг 3: сохраняем независимую копию после проверки владения. Retention отсчитываем
	// от завершения, чтобы время выполнения не съедало окно повторного чтения ответа.
	stored := cloneResponse(response)
	entry.response = &stored
	entry.expiresAt = now.Add(retention)
	p.records[key] = entry
	return nil
}

var _ Provider = (*MemoryProvider)(nil)
