package http

import (
	"net/http"
	"time"
)

// healthCheck returns basic service and config info.
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
		"config": map[string]interface{}{
			"maxRequests": s.deps.Config.RateLimitMax,
			"windowSec":   s.deps.Config.RateLimitWindow,
		},
	}
	writeJSON(w, http.StatusOK, response)
}

