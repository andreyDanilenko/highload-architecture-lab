package helpers

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with given status code.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError writes an error response with a simple JSON body.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	body := map[string]string{
		"code":    code,
		"message": message,
	}
	WriteJSON(w, status, body)
}
