package http

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes a JSON response with given status code.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes an error response with a simple JSON body.
func writeError(w http.ResponseWriter, status int, code, message string) {
	body := map[string]string{
		"code":    code,
		"message": message,
	}
	writeJSON(w, status, body)
}

