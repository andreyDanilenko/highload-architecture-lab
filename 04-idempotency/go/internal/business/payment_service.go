package business

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"idempotency/internal/domain"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *domain.PaymentRequest) (*domain.PaymentResult, error) {
	// Имитация задержки
	delay := time.Duration(rand.Intn(200)) * time.Millisecond

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Имитация ошибки (5%)
	if rand.Float64() < 0.05 {
		return domain.NewFailedPayment("insufficient funds"), nil
	}

	// Успех
	txID := fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
	return domain.NewSuccessfulPayment(txID, req.FromAccountID, req.ToAccountID, req.Amount), nil
}
