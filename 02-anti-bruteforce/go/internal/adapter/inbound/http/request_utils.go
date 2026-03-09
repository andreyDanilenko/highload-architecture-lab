package http

import "net/http"

// clientIP extracts the client IP from common proxy headers or RemoteAddr.
func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return x
	}
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return x
	}
	return r.RemoteAddr
}

