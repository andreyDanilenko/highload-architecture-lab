package handlers

import (
	"net/http"

	"anti-bruteforce/internal/adapter/inbound/http/helpers"
)

// NotFound returns 404 for unknown paths.
func NotFound(w http.ResponseWriter, r *http.Request) {
	helpers.WriteError(w, http.StatusNotFound, "not_found", "Not found")
}

