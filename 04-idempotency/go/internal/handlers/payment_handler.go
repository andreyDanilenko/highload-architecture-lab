package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"idempotency/internal/business"
	"idempotency/internal/domain"
	"idempotency/pkg/idempotency"
)

type PaymentHandler struct {
	paymentService *business.PaymentService
	idempotencyMw  *idempotency.Middleware
}

func NewPaymentHandler(
	paymentService *business.PaymentService,
	idempotencyMw *idempotency.Middleware,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		idempotencyMw:  idempotencyMw,
	}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем наличие idempotency key
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	// 2. Декодируем тело запроса
	var req domain.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 3. Выполняем с идемпотентностью
	result, err := idempotency.Execute(h.idempotencyMw, r.Context(), key, func() (*domain.PaymentResult, error) {
		return h.paymentService.ProcessPayment(r.Context(), &req)
	})

	// 4. Обрабатываем ошибки
	if err != nil {
		if errors.Is(err, idempotency.ErrKeyAlreadyProcessing) {
			respondConflict(w, "request with this key is already processing", 5)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. Успешный ответ
	respondJSON(w, http.StatusOK, result)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondConflict(w http.ResponseWriter, message string, retryAfter int) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	respondError(w, http.StatusConflict, message)
}
