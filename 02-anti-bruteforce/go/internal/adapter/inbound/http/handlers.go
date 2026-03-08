package http

import (
	"encoding/json"
	"net/http"
	"time"

	"anti-bruteforce/internal/usecase"
)

// healthCheck — GET /health.
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
		"config": map[string]interface{}{
			"maxRequests": s.deps.Config.RateLimitMax,
			"windowSec":   s.deps.Config.RateLimitWindow,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// notFound returns 404 for unknown paths.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not found", http.StatusNotFound)
}

// handleLoginRequired returns a handler that runs the checker and responds 200/429/500.
func (s *Server) handleLoginRequired(checker usecase.LoginChecker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if checker == nil {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
			return
		}
		ip := clientIP(r)
		result := checker.Check(ip)
		if result.Err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if result.Exceeded {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		// Allowed: mock login success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return x
	}
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return x
	}
	return r.RemoteAddr
}
