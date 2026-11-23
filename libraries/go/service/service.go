// Package service is the standard main for a Go service: config, logging,
// tracing, auth, clients for other services, the database, health
// checks, the HTTP server and graceful shutdown. A service's main is:
//
//	func main() {
//		service.Run(service.Options{Name: "orders", Migrations: db.Migrations}, setup)
//	}
//
//	func setup(ctx context.Context, s *service.Service) error {
//		pool, err := s.Database(ctx)
//		...
//		mux.Handle(ordersv1connect.NewOrdersServiceHandler(h, s.HandlerOptions()...))
//		return nil
//	}
//
// Environment (besides what setup reads from s.Env):
//
//	APP_ENV                  development | test | staging | production
//	HOST, PORT               listen address (default 0.0.0.0:8080)
//	IDENTITY_PUBLIC_KEYS     user identity token keys (required to serve RPCs)
//	UPSTREAM_URL             template for other services' URLs (http://{name}:8080)
//	SERVICE_URL_<NAME>       override one service's URL
//	DATABASE_URL             for s.Database
//	AUTO_MIGRATE             apply migrations on startup (local dev)
//	OTEL_EXPORTER_OTLP_*     telemetry export (see libraries/go/telemetry)
//	SERVICE_VERSION          reported in telemetry (deploys set the git SHA)
//
// `server migrate` applies Options.Migrations and exits; deploys run it as a
// one-off task before rolling out.
package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zachnoble/laurel/libraries/go/config"
	"github.com/zachnoble/laurel/libraries/go/db"
	"github.com/zachnoble/laurel/libraries/go/logging"
	"github.com/zachnoble/laurel/libraries/go/rpc"
	"github.com/zachnoble/laurel/libraries/go/server"
	"github.com/zachnoble/laurel/libraries/go/telemetry"
)

type Options struct {
	// Name is the service name: its directory, proto package, and DNS name.
	Name string
	// Migrations (golang-migrate layout, at the FS root) for `server migrate`,
	// AUTO_MIGRATE and s.Database.
	Migrations fs.FS
	// Edge marks the public entry point (the gateway): incoming request IDs,
	// trace context and identity headers from clients aren't trusted.
	Edge bool
}

// Service is what setup builds the service from.
type Service struct {
	Name   string
	AppEnv string
	Env    *config.Env
	Logger *slog.Logger
	// Mux serves the service. Health checks are mounted after setup.
	Mux *http.ServeMux
	// Clients calls other services: rpc.NewClient(s.Clients, "orders", ordersv1connect.NewOrdersServiceClient).
	Clients *rpc.Clients

	opts       Options
	auth       *rpc.Authenticator
	errs       []error
	middleware []server.Middleware
	checks     []server.Check
	background []func(context.Context)
	cleanups   []func()
}

// Run runs the service and exits the process on failure.
func Run(opts Options, setup func(ctx context.Context, s *Service) error) {
	if err := run(opts, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", opts.Name, err)
		os.Exit(1)
	}
}

func run(opts Options, setup func(ctx context.Context, s *Service) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	env := config.Load()
	logger := logging.New().With("service", opts.Name)

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if opts.Migrations == nil {
			return errors.New("migrate: this service has no migrations")
		}
		url := env.Required("DATABASE_URL")
		if err := env.Err(); err != nil {
			return err
		}
		return db.Migrate(url, opts.Migrations, logger)
	}

	shutdownTelemetry, err := telemetry.Setup(ctx, opts.Name, env.String("SERVICE_VERSION", "dev"))
	if err != nil {
		return fmt.Errorf("telemetry: %w", err)
	}

	s := &Service{
		Name:   opts.Name,
		AppEnv: env.String("APP_ENV", "production"),
		Env:    env,
		Logger: logger,
		Mux:    http.NewServeMux(),
		opts:   opts,
	}
	defer func() {
		stop() // background work sees shutdown before resources close
		for _, c := range slices.Backward(s.cleanups) {
			c()
		}
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(flushCtx); err != nil {
			logger.Error("flush telemetry", "error", err)
		}
	}()

	if err := s.init(); err != nil {
		return err
	}
	if err := setup(ctx, s); err != nil {
		return err
	}
	if err := errors.Join(append(s.errs, env.Err())...); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	server.RegisterHealth(s.Mux, s.checks...)

	chain := []server.Middleware{server.Trace(opts.Edge)}
	if opts.Edge {
		// Only the edge assigns request IDs, and only it may assert identity.
		chain = append(chain, server.DropHeader(server.RequestIDHeader, rpc.IdentityHeader))
	}
	chain = append(chain, server.RequestLogger(logger), server.Recover)
	chain = append(chain, s.middleware...)

	srv := server.New(server.Config{Host: env.String("HOST", "0.0.0.0"), Port: env.Int("PORT", 8080)}, server.Chain(s.Mux, chain...))

	for _, fn := range s.background {
		go fn(ctx)
	}
	return srv.Run(ctx, logger)
}

func (s *Service) init() error {
	if userKeys := s.Env.List("IDENTITY_PUBLIC_KEYS"); len(userKeys) > 0 {
		users, err := rpc.NewVerifier(userKeys)
		if err != nil {
			return err
		}
		s.auth = rpc.NewAuthenticator(users)
	}

	s.Clients = rpc.NewClients(rpc.ClientsConfig{
		URLTemplate: s.Env.String("UPSTREAM_URL", "http://{name}:8080"),
		Overrides:   serviceURLOverrides(),
	})
	return nil
}

// serviceURLOverrides reads SERVICE_URL_<NAME>=<url>.
func serviceURLOverrides() map[string]string {
	out := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		if name, ok := strings.CutPrefix(k, "SERVICE_URL_"); ok && v != "" {
			out[strings.ToLower(name)] = v
		}
	}
	return out
}

// HandlerOptions returns the options for mounting a Connect handler: error
// mapping, the schema's access rules (rpc.Authenticator), validation, and
// any extra interceptors.
func (s *Service) HandlerOptions(extra ...connect.Interceptor) []connect.HandlerOption {
	if s.auth == nil {
		s.errs = append(s.errs, errors.New("IDENTITY_PUBLIC_KEYS is required to serve RPCs"))
		return nil
	}
	return rpc.HandlerOptions(append([]connect.Interceptor{s.auth}, extra...)...)
}

// Database connects to DATABASE_URL (applying migrations first if
// AUTO_MIGRATE is set), adds it to the readiness check, and closes it on
// shutdown.
func (s *Service) Database(ctx context.Context) (*pgxpool.Pool, error) {
	url := s.Env.Required("DATABASE_URL")
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if s.opts.Migrations != nil && s.Env.Bool("AUTO_MIGRATE", false) {
		if err := db.Migrate(url, s.opts.Migrations, s.Logger); err != nil {
			return nil, err
		}
	}
	pool, err := db.Connect(ctx, db.Config{URL: url, ApplicationName: s.Name})
	if err != nil {
		return nil, err
	}
	s.AddCheck(pool.Ping)
	s.OnShutdown(pool.Close)
	return pool, nil
}

// AddCheck adds a readiness check.
func (s *Service) AddCheck(c server.Check) { s.checks = append(s.checks, c) }

// Use adds HTTP middleware inside the standard chain.
func (s *Service) Use(m ...server.Middleware) { s.middleware = append(s.middleware, m...) }

// Go runs fn in the background until shutdown.
func (s *Service) Go(fn func(ctx context.Context)) { s.background = append(s.background, fn) }

// OnShutdown runs fn after the server stops, in reverse order of registration.
func (s *Service) OnShutdown(fn func()) { s.cleanups = append(s.cleanups, fn) }
