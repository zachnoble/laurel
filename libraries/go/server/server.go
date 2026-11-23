package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Config configures the HTTP server. Zero fields use the defaults below.
type Config struct {
	// Host to listen on. Default "0.0.0.0".
	Host string
	// Port to listen on. Default 8080.
	Port int

	// ReadHeaderTimeout limits how long a client may take to send request
	// headers; this is the main slowloris protection. Default 5s.
	ReadHeaderTimeout time.Duration
	// ReadTimeout limits reading the whole request, body included. Default 30s.
	ReadTimeout time.Duration
	// WriteTimeout limits the whole request, from the end of the headers until
	// the response is written. Default 30s. Streaming handlers that need longer
	// can extend their own deadline with http.ResponseController.
	WriteTimeout time.Duration
	// IdleTimeout for keep-alive connections. Default 120s: longer than the
	// ALB's 60s idle timeout, so the ALB always closes first (otherwise it can
	// reuse a connection the server just closed and return a 502).
	IdleTimeout time.Duration

	// MaxHeaderBytes per request. Default 64 KiB (session cookie + identity
	// token fit easily).
	MaxHeaderBytes int
	// MaxBodyBytes per request; larger bodies fail to read. Default 4 MiB.
	MaxBodyBytes int64

	// ShutdownTimeout for in-flight requests after SIGTERM. Default 25s, which
	// fits inside ECS's default 30s stopTimeout before SIGKILL.
	ShutdownTimeout time.Duration
}

const (
	defaultHost              = "0.0.0.0"
	defaultPort              = 8080
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultMaxHeaderBytes    = 64 << 10
	defaultMaxBodyBytes      = 4 << 20
	defaultShutdownTimeout   = 25 * time.Second
)

type Server struct {
	http            *http.Server
	shutdownTimeout time.Duration
}

// New returns a server for handler, serving HTTP/1.1 and unencrypted HTTP/2.
func New(cfg Config, handler http.Handler) *Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &Server{
		http: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", or(cfg.Host, defaultHost), or(cfg.Port, defaultPort)),
			Handler:           http.MaxBytesHandler(handler, or(cfg.MaxBodyBytes, defaultMaxBodyBytes)),
			Protocols:         protocols,
			ReadHeaderTimeout: or(cfg.ReadHeaderTimeout, defaultReadHeaderTimeout),
			ReadTimeout:       or(cfg.ReadTimeout, defaultReadTimeout),
			WriteTimeout:      or(cfg.WriteTimeout, defaultWriteTimeout),
			IdleTimeout:       or(cfg.IdleTimeout, defaultIdleTimeout),
			MaxHeaderBytes:    or(cfg.MaxHeaderBytes, defaultMaxHeaderBytes),
		},
		shutdownTimeout: or(cfg.ShutdownTimeout, defaultShutdownTimeout),
	}
}

// Run serves until ctx is cancelled, then waits up to ShutdownTimeout for
// in-flight requests. It returns an error if the server fails to start or
// doesn't shut down cleanly.
func (s *Server) Run(ctx context.Context, logger *slog.Logger) error {
	errs := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "addr", s.http.Addr)
		errs <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return fmt.Errorf("HTTP server: %w", err)
	case <-ctx.Done():
	}

	logger.Info("HTTP server shutting down", "timeout", s.shutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}
	if err := <-errs; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server: %w", err)
	}
	return nil
}

func or[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}
