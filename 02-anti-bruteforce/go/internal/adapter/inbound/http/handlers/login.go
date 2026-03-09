package handlers

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/helpers"
	"anti-bruteforce/internal/usecase"
)

// LoginHandlers groups all handlers that share login checkers.
type LoginHandlers struct {
	Naive       usecase.LoginChecker
	Pessimistic usecase.LoginChecker
	Optimistic  usecase.LoginChecker
	Vault       usecase.LoginChecker
}

func NewLoginHandlers(
	naive, pessimistic, optimistic, vault usecase.LoginChecker,
) *LoginHandlers {
	return &LoginHandlers{
		Naive:       naive,
		Pessimistic: pessimistic,
		Optimistic:  optimistic,
		Vault:       vault,
	}
}

// Login handles POST /login using the naive checker.
func (h *LoginHandlers) Login(w http.ResponseWriter, r *http.Request) {
	h.handleWithChecker(w, r, h.Naive)
}

// NaiveResource handles POST /resource/naive using the naive checker.
func (h *LoginHandlers) NaiveResource(w http.ResponseWriter, r *http.Request) {
	h.handleWithChecker(w, r, h.Naive)
}

// PessimisticResource handles POST /resource/pessimistic.
func (h *LoginHandlers) PessimisticResource(w http.ResponseWriter, r *http.Request) {
	h.handleWithChecker(w, r, h.Pessimistic)
}

// OptimisticResource handles POST /resource/optimistic.
func (h *LoginHandlers) OptimisticResource(w http.ResponseWriter, r *http.Request) {
	h.handleWithChecker(w, r, h.Optimistic)
}

// VaultLogin handles POST /vault/login.
func (h *LoginHandlers) VaultLogin(w http.ResponseWriter, r *http.Request) {
	h.handleWithChecker(w, r, h.Vault)
}

// handleWithChecker runs the checker and responds 200/429/500.
func (h *LoginHandlers) handleWithChecker(w http.ResponseWriter, r *http.Request, checker usecase.LoginChecker) {
	if r.Method != http.MethodPost {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	if checker == nil {
		helpers.WriteError(w, http.StatusNotImplemented, "not_implemented", "Strategy not implemented")
		return
	}
	ip := helpers.ClientIP(r)
	result := checker.Check(ip)
	if result.Err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}
	if result.Exceeded {
		helpers.WriteError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests")
		return
	}
	helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

