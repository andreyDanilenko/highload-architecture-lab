package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"idempotency/internal/domain"
)

// PaymentService — бизнес-логика платежа (без HTTP и без идемпотентности).
type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPayment имитирует side-effect: списание, запись в БД, вызов PSP.
func (s *PaymentService) ProcessPayment(ctx context.Context, req *domain.PaymentRequest) (*domain.PaymentResult, error) {
	// Шаг 1: имитация сетевой/DB задержки
	delay := time.Duration(rand.Intn(200)) * time.Millisecond
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Шаг 2: имитация бизнес-ошибки (5%)
	if rand.Float64() < 0.05 {
		return domain.NewFailedPayment("insufficient funds"), nil
	}

	// Шаг 3: успешный платёж
	txID := fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
	return domain.NewSuccessfulPayment(txID, req.FromAccountID, req.ToAccountID, req.Amount), nil
}
