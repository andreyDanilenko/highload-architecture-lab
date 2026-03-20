package helpers

import (
	"net"
	"net/http"
)

// ClientIP extracts the client IP from trusted proxy headers or RemoteAddr.
// - In production, X-Forwarded-For / X-Real-IP should be set by a trusted
//   frontend (nginx, Envoy, load balancer) which strips any client-supplied
//   values.
// - We always strip the ephemeral port from RemoteAddr so that multiple
//   requests from the same client are counted under the same IP key.
func ClientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return x
	}
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return x
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}


