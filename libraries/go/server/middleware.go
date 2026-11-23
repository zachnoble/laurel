package server

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/zachnoble/laurel/libraries/go/logging"
)

type Middleware func(http.Handler) http.Handler

// Chain wraps h with middlewares, the first being the outermost.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range slices.Backward(middlewares) {
		h = middleware(h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// RequestIDHeader correlates logs across services. The gateway sets it and it
// is forwarded on every proxied request.
const RequestIDHeader = "X-Request-Id"

// RequestLogger attaches a request-scoped logger (see logging.FromContext)
// and logs each request's start and completion. It reuses an incoming
// X-Request-Id, or sets a new one on the request so proxies forward it.
func RequestLogger(base *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
				r.Header.Set(RequestIDHeader, requestID)
			}
			l := base.With("requestId", requestID)
			if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
				l = l.With("traceId", sc.TraceID().String(), "spanId", sc.SpanID().String())
			}
			r = r.WithContext(logging.WithLogger(r.Context(), l))

			method := r.Method
			path := r.URL.Path
			start := time.Now()

			l.Info("Incoming request", "method", method, "path", path)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			l.Info("Request completed",
				"method", method,
				"path", path,
				"status", rec.status,
				"durationMs", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Recover turns panics into a 500 response.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logging.FromContext(r.Context()).Error("panic recovered", "error", rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Trace starts a server span per request (named by path, which for Connect is
// the procedure), continuing the caller's trace from its traceparent header.
// At the public edge, pass public=true: incoming trace context is linked
// rather than trusted as the parent. Put it first, so RequestLogger can log
// the trace ID.
func Trace(public bool) Middleware {
	opts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return r.URL.Path }),
		otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != "/health" && r.URL.Path != "/readiness" }),
	}
	if public {
		opts = append(opts, otelhttp.WithPublicEndpointFn(func(*http.Request) bool { return true }))
	}
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "server", opts...)
	}
}

// DropHeader removes a request header, e.g. one only this hop may set.
func DropHeader(names ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, name := range names {
				r.Header.Del(name)
			}
			next.ServeHTTP(w, r)
		})
	}
}
