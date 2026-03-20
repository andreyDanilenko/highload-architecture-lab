package routes

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/handlers"
	"anti-bruteforce/internal/config"
)

func registerHealthRoutes(mux *http.ServeMux, cfg *config.Config) {
	health := handlers.NewHealthHandler(cfg)
	mux.HandleFunc("/health", health.Health)
}

