package handlers

import (
	"net/http"
	"time"

	"worker-pool/internal/adapter/inbound/http/helpers"
	"worker-pool/internal/config"
)

// HealthHandler exposes basic service info.
type HealthHandler struct {
	Config *config.Config
}

// NewHealthHandler builds a health handler.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{Config: cfg}
}

// Health returns basic service and pool sizing from config.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
		"config": map[string]interface{}{
			"workers":             h.Config.Workers,
			"queueSize":           h.Config.QueueSize,
			"advancedWorkers":     h.Config.AdvancedWorkers,
			"advancedQueueSize":   h.Config.AdvancedQueueSize,
			"taskSimulateMs":      h.Config.TaskSimulateMs,
		},
	}
	helpers.WriteJSON(w, http.StatusOK, response)
}
