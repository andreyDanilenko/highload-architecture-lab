package routes

import (
	"context"
	"net/http"
	"time"

	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/pkg/idempotency"
	"labshared/httpx"
)

// Register задаёт границу middleware: только endpoint, который создаёт эффект.
func Register(mux *http.ServeMux, payment *handlers.PaymentHandler, middleware *idempotency.HTTPMiddleware, ready func(context.Context) error) {
	mux.Handle("POST /api/v1/payments", middleware.Wrap(http.HandlerFunc(payment.CreatePayment)))
	mux.HandleFunc("GET /api/v1/effects", payment.ListEffects)
	mux.Handle("GET /metrics", middleware.MetricsHandler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if ready != nil {
			if err := ready(ctx); err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "storage is unavailable")
				return
			}
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}
