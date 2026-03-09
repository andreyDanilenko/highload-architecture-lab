package http

import "net/http"

// notFound returns 404 for unknown paths.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "Not found")
}
