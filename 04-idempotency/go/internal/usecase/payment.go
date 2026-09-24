package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"idempotency/internal/domain"
)

var ErrLedgerFull = errors.New("effect ledger is full")

// PaymentService — ограниченный журнал учебных эффектов, не банковский ledger.
// Он не дедуплицирует запросы: дубли должны быть видны независимо от ответов middleware.
type PaymentService struct {
	mu         sync.Mutex
	effects    []domain.PaymentResult
	delay      time.Duration
	maxEffects int
}

func NewPaymentService(delay time.Duration, maxEffects int) *PaymentService {
	return &PaymentService{delay: delay, maxEffects: maxEffects}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *domain.PaymentRequest) (*domain.PaymentResult, error) {
	// Шаг 1: проверяем инварианты и имитируем только заданную, воспроизводимую задержку.
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if s.delay > 0 {
		timer := time.NewTimer(s.delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	// Шаг 2: запись в журнал — отдельный эффект, не атомарный с хранилищем идемпотентности.
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(s.effects) >= s.maxEffects {
		return nil, ErrLedgerFull
	}
	result := domain.PaymentResult{TransactionID: fmt.Sprintf("tx_%d", len(s.effects)+1), Status: "completed", Amount: req.Amount, FromAccount: req.FromAccountID, ToAccount: req.ToAccountID, CompletedAt: time.Now().UTC()}
	s.effects = append(s.effects, result)
	return &result, nil
}

// Effects возвращает снимок, чтобы чтение журнала не участвовало в изменении эффекта.
func (s *PaymentService) Effects() []domain.PaymentResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.PaymentResult{}, s.effects...)
}
