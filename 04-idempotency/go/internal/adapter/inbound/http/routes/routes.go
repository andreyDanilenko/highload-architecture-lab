package routes

import (
	"net/http"

	"idempotency/internal/adapter/inbound/http/handlers"
)

// Register — шаг сборки HTTP API: один mux, явные маршруты.
func Register(mux *http.ServeMux, payment *handlers.PaymentHandler) {
	mux.HandleFunc("POST /api/v1/payments", payment.CreatePayment)
}
