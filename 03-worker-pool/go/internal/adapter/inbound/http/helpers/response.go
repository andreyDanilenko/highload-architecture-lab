package helpers

import (
	"net/http"

	"labshared/httpx"
)

// WriteJSON — thin wrapper над общим labshared/httpx (единый контракт для всех лаб).
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	httpx.WriteJSON(w, status, payload)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	httpx.WriteError(w, status, code, message)
}
