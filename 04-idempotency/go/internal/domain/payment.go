package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidPayment = errors.New("invalid payment")

// PaymentRequest — учебная операция; amount выражен целым числом минимальных денежных единиц.
type PaymentRequest struct {
	FromAccountID string `json:"from"`
	ToAccountID   string `json:"to"`
	Amount        int64  `json:"amount"`
}

// Validate проверяет инварианты до эффекта, независимо от транспортного адаптера.
func (p PaymentRequest) Validate() error {
	if strings.TrimSpace(p.FromAccountID) == "" || strings.TrimSpace(p.ToAccountID) == "" || p.FromAccountID == p.ToAccountID || p.Amount <= 0 || len(p.FromAccountID) > 128 || len(p.ToAccountID) > 128 {
		return ErrInvalidPayment
	}
	return nil
}

// PaymentResult фиксируется в журнале эффектов и сохраняется как HTTP-ответ для replay.
type PaymentResult struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        int64     `json:"amount"`
	FromAccount   string    `json:"from_account"`
	ToAccount     string    `json:"to_account"`
	CompletedAt   time.Time `json:"completed_at"`
}
