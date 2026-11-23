package main

import (
	"context"
	"time"

	"github.com/resend/resend-go/v3"

	"github.com/zachnoble/laurel/libraries/go/rpc"
	"github.com/zachnoble/laurel/libraries/go/service"
	"github.com/zachnoble/laurel/services/auth/internal/auth"
	"github.com/zachnoble/laurel/services/auth/internal/config"
	"github.com/zachnoble/laurel/services/auth/internal/db"
)

func main() {
	service.Run(service.Options{Name: "auth", Migrations: db.Migrations}, setup)
}

func setup(ctx context.Context, s *service.Service) error {
	cfg, err := config.New(s.Env)
	if err != nil {
		return err
	}

	pool, err := s.Database(ctx)
	if err != nil {
		return err
	}
	database := db.New(pool)

	signer, err := rpc.NewSigner(cfg.IdentityPrivateKey, 5*time.Minute)
	if err != nil {
		return err
	}

	authService := auth.New(database, resend.NewClient(cfg.ResendAPIKey), cfg)
	auth.NewHandler(authService, cfg.CookieSecret, signer).Mount(s.Mux, s.HandlerOptions()...)

	s.Go(func(ctx context.Context) { database.RunCleanup(ctx, time.Hour, s.Logger) })
	return nil
}
