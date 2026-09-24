package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"idempotency/internal/domain"
	"idempotency/internal/usecase"
	"idempotency/pkg/idempotency"
	"labshared/httpx"
)

// PaymentHandler — inbound HTTP adapter: JSON ↔ use case.
type PaymentHandler struct {
	payments      *usecase.PaymentService
	idempotencyMw *idempotency.Middleware
}

func NewPaymentHandler(payments *usecase.PaymentService, mw *idempotency.Middleware) *PaymentHandler {
	return &PaymentHandler{payments: payments, idempotencyMw: mw}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: клиент обязан передать Idempotency-Key (RFC-подобный контракт)
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_key", "Idempotency-Key header is required")
		return
	}

	// Шаг 2: разбор тела запроса
	var req domain.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	// Шаг 3: обёртка бизнес-логики идемпотентностью (TryLock → fn → Complete)
	result, err := idempotency.Execute(h.idempotencyMw, r.Context(), key, func() (*domain.PaymentResult, error) {
		return h.payments.ProcessPayment(r.Context(), &req)
	})

	// Шаг 4: маппинг доменных/инфра ошибок в HTTP
	if err != nil {
		if errors.Is(err, idempotency.ErrKeyAlreadyProcessing) {
			httpx.WriteConflict(w, "already_processing", "request with this key is already processing", 5)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Шаг 5: ответ (повторный запрос с тем же ключом вернёт тот же JSON)
	httpx.WriteJSON(w, http.StatusOK, result)
}
