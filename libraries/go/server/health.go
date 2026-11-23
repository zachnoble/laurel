package server

import (
	"context"
	"net/http"
	"time"
)

// checkTimeout keeps a hung dependency from hanging the readiness probe
// (ALB and ECS health checks time out at 5s by default).
const checkTimeout = 2 * time.Second

// Check reports whether a dependency is ready, e.g. (*pgxpool.Pool).Ping.
type Check func(ctx context.Context) error

// RegisterHealth mounts GET /health (liveness) and GET /readiness, which
// fails with 503 if any check fails.
func RegisterHealth(mux *http.ServeMux, checks ...Check) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readiness", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), checkTimeout)
		defer cancel()
		for _, check := range checks {
			if err := check(ctx); err != nil {
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
		}
		_, _ = w.Write([]byte("ok"))
	})
}
