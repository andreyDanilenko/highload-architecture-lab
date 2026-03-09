package http

import (
	"net/http"

	"anti-bruteforce/internal/usecase"
)

// loginHandler handles POST /login using the naive checker.
func (s *Server) loginHandler() http.HandlerFunc {
	return s.handleLoginWithChecker(s.deps.NaiveChecker)
}

// naiveResourceHandler handles POST /resource/naive using the naive checker.
func (s *Server) naiveResourceHandler() http.HandlerFunc {
	return s.handleLoginWithChecker(s.deps.NaiveChecker)
}

// pessimisticResourceHandler handles POST /resource/pessimistic.
func (s *Server) pessimisticResourceHandler() http.HandlerFunc {
	return s.handleLoginWithChecker(s.deps.PessimisticChecker)
}

// optimisticResourceHandler handles POST /resource/optimistic.
func (s *Server) optimisticResourceHandler() http.HandlerFunc {
	return s.handleLoginWithChecker(s.deps.OptimisticChecker)
}

// vaultLoginHandler handles POST /vault/login.
func (s *Server) vaultLoginHandler() http.HandlerFunc {
	return s.handleLoginWithChecker(s.deps.VaultChecker)
}

// handleLoginWithChecker returns a handler that runs the checker and responds 200/429/500.
func (s *Server) handleLoginWithChecker(checker usecase.LoginChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
			return
		}
		if checker == nil {
			writeError(w, http.StatusNotImplemented, "not_implemented", "Strategy not implemented")
			return
		}
		ip := clientIP(r)
		result := checker.Check(ip)
		if result.Err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
			return
		}
		if result.Exceeded {
			writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
