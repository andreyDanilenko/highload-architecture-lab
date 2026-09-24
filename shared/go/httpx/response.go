// Package httpx — общие HTTP-ответы для всех Go-лабораторий челленджа.
// Один формат ошибок {code, message} → проще масштабировать и тестировать.
package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// WriteJSON — шаг 1 ответа: заголовки, статус, JSON-тело.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError — шаг ошибки: единый контракт для всех сервисов.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, map[string]string{
		"code":    code,
		"message": message,
	})
}

// WriteConflict — 409 + Retry-After (идемпотентность: ключ ещё в обработке).
func WriteConflict(w http.ResponseWriter, code, message string, retryAfterSec int) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
	WriteError(w, http.StatusConflict, code, message)
}
