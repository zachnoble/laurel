package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	s := New(Config{}, http.NotFoundHandler())
	h := s.http
	if h.Addr != "0.0.0.0:8080" || h.ReadHeaderTimeout != 5*time.Second || h.ReadTimeout != 30*time.Second ||
		h.WriteTimeout != 30*time.Second || h.IdleTimeout != 120*time.Second || h.MaxHeaderBytes != 64<<10 ||
		s.shutdownTimeout != 25*time.Second {
		t.Errorf("unexpected defaults: addr=%s rh=%v r=%v w=%v idle=%v hdr=%d shutdown=%v",
			h.Addr, h.ReadHeaderTimeout, h.ReadTimeout, h.WriteTimeout, h.IdleTimeout, h.MaxHeaderBytes, s.shutdownTimeout)
	}
}

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestRunShutsDownOnCancelAndLimitsBodies(t *testing.T) {
	port := freePort(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		}
	})
	srv := New(Config{Host: "127.0.0.1", Port: port, MaxBodyBytes: 10}, handler)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx, slog.New(slog.DiscardHandler)) }()

	addr := srv.http.Addr
	var res *http.Response
	var err error
	for range 50 {
		if res, err = http.Post("http://"+addr, "text/plain", strings.NewReader(strings.Repeat("x", 100))); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status %d, want 413", res.StatusCode)
	}

	cancel()
	if err := <-done; err != nil {
		t.Errorf("Run after cancel = %v, want nil", err)
	}
}

func TestRunReturnsListenError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	srv := New(Config{Host: "127.0.0.1", Port: l.Addr().(*net.TCPAddr).Port}, http.NotFoundHandler())
	if err := srv.Run(context.Background(), slog.New(slog.DiscardHandler)); err == nil {
		t.Error("Run on a taken port returned nil")
	}
}

func TestReadinessTimesOutHungChecks(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux, func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })

	start := time.Now()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readiness", nil))
	if rec.Code != http.StatusServiceUnavailable || time.Since(start) > 3*time.Second {
		t.Errorf("status %d after %v", rec.Code, time.Since(start))
	}
}
