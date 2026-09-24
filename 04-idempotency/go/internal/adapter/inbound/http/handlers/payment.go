package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"idempotency/internal/domain"
	"idempotency/internal/usecase"
	"labshared/httpx"
)

// PaymentHandler переводит HTTP в команду; идемпотентность подключается на уровне маршрута.
type PaymentHandler struct{ payments *usecase.PaymentService }

func NewPaymentHandler(payments *usecase.PaymentService) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: принимаем ровно один JSON-объект, без неизвестных полей и хвостовых данных.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var req domain.PaymentRequest
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "expected from, to and a positive integer amount")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "exactly one JSON object is required")
		return
	}
	if err := req.Validate(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_payment", "distinct nonempty accounts and positive integer amount are required")
		return
	}
	// Шаг 2: выполняем одну попытку. Журнал хранит факт эффекта отдельно от HTTP-кэша.
	result, err := h.payments.ProcessPayment(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			httpx.WriteError(w, http.StatusRequestTimeout, "request_canceled", "operation was canceled before recording an effect")
		case errors.Is(err, usecase.ErrLedgerFull):
			httpx.WriteError(w, http.StatusServiceUnavailable, "ledger_full", "lab effect ledger reached its configured capacity")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "operation could not be completed")
		}
		return
	}
	// Шаг 3: middleware сохранит эти байты до отправки клиенту.
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *PaymentHandler) ListEffects(w http.ResponseWriter, _ *http.Request) {
	effects := h.payments.Effects()
	httpx.WriteJSON(w, http.StatusOK, struct {
		Count    int                    `json:"count"`
		Payments []domain.PaymentResult `json:"payments"`
	}{len(effects), effects})
}
