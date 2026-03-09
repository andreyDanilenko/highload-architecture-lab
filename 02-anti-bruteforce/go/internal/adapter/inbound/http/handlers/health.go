package handlers

import (
	"net/http"
	"time"

	"anti-bruteforce/internal/adapter/inbound/http/helpers"
	"anti-bruteforce/internal/config"
)

type HealthHandler struct {
	Config *config.Config
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{Config: cfg}
}

// Health returns basic service and config info.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
		"config": map[string]interface{}{
			"maxRequests": h.Config.RateLimitMax,
			"windowSec":   h.Config.RateLimitWindow,
		},
	}
	helpers.WriteJSON(w, http.StatusOK, response)
}

