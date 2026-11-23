package main

import (
	"context"

	"github.com/zachnoble/laurel/libraries/go/rpc"
	"github.com/zachnoble/laurel/libraries/go/service"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
	"github.com/zachnoble/laurel/services/gateway/internal/config"
	"github.com/zachnoble/laurel/services/gateway/internal/gateway"
)

func main() {
	service.Run(service.Options{Name: "gateway", Edge: true}, setup)
}

func setup(_ context.Context, s *service.Service) error {
	cfg := config.New(s.Env)

	authClient := rpc.NewClient(s.Clients, cfg.SessionService, authv1connect.NewAuthInternalServiceClient)
	sessions := gateway.NewSessions(authClient, cfg.SessionCookie, cfg.SessionCacheSize, cfg.SessionCacheTTL)

	s.Mux.Handle("/", gateway.New(s.Clients.URL, rpc.Transport(), cfg.SessionService, sessions))
	s.Use(rpc.CORS(cfg.AllowedOrigins))
	return nil
}
