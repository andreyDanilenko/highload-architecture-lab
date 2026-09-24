package domain

import "time"

// PaymentRequest — вход use case (без HTTP-деталей).
type PaymentRequest struct {
	FromAccountID string  `json:"from"`
	ToAccountID   string  `json:"to"`
	Amount        float64 `json:"amount"`
}

// PaymentResult — результат side-effect; кэшируется провайдером идемпотентности.
type PaymentResult struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	FromAccount   string    `json:"from_account"`
	ToAccount     string    `json:"to_account"`
	CompletedAt   time.Time `json:"completed_at"`
	Error         string    `json:"error,omitempty"`
}

func NewSuccessfulPayment(txID, from, to string, amount float64) *PaymentResult {
	return &PaymentResult{
		TransactionID: txID,
		Status:        "completed",
		Amount:        amount,
		FromAccount:   from,
		ToAccount:     to,
		CompletedAt:   time.Now(),
	}
}

func NewFailedPayment(reason string) *PaymentResult {
	return &PaymentResult{
		Status:      "failed",
		CompletedAt: time.Now(),
		Error:       reason,
	}
}
