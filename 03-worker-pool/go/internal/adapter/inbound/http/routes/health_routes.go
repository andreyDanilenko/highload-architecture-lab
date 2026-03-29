package routes

import (
	"net/http"

	"worker-pool/internal/adapter/inbound/http/handlers"
	"worker-pool/internal/config"
)

func registerHealthRoutes(mux *http.ServeMux, cfg *config.Config) {
	health := handlers.NewHealthHandler(cfg)
	mux.HandleFunc("/health", health.Health)
}
