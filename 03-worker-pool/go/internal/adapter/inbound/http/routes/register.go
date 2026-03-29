package routes

import (
	"net/http"

	"worker-pool/internal/adapter/inbound/http/handlers"
	"worker-pool/internal/config"
	"worker-pool/internal/usecase"
)

// Register wires all HTTP routes to their handlers.
func Register(
	mux *http.ServeMux,
	cfg *config.Config,
	naive, bounded, reliable, advanced usecase.TaskDispatcher,
) {
	registerHealthRoutes(mux, cfg)
	registerWorkRoutes(mux, cfg, naive, bounded, reliable, advanced)

	mux.HandleFunc("/", handlers.NotFound)
}
