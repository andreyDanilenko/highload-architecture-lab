package routes

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/handlers"
	"anti-bruteforce/internal/config"
	"anti-bruteforce/internal/usecase"
)

// Register wires all HTTP routes to their handlers.
func Register(
	mux *http.ServeMux,
	cfg *config.Config,
	naive, pessimistic, optimistic, vault usecase.LoginChecker,
) {
	// Health / diagnostics
	registerHealthRoutes(mux, cfg)

	// Auth / login-like flows
	registerLoginRoutes(mux, naive, pessimistic, optimistic, vault)

	// Protected resources with different anti-bruteforce strategies
	registerResourceRoutes(mux, naive, pessimistic, optimistic)

	// Fallback
	mux.HandleFunc("/", handlers.NotFound)
}

