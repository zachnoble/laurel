package service

import (
	"testing"

	"github.com/zachnoble/laurel/libraries/go/config"
)

func newService(t *testing.T, appEnv string) *Service {
	t.Helper()
	return &Service{Name: "orders", AppEnv: appEnv, Env: config.Load("/nonexistent")}
}

func TestServiceURLOverrides(t *testing.T) {
	t.Setenv("SERVICE_URL_INVENTORY", "https://inventory.dev.example.com")
	s := newService(t, "development")
	if err := s.init(); err != nil {
		t.Fatal(err)
	}
	if got := s.Clients.URL("inventory"); got != "https://inventory.dev.example.com" {
		t.Errorf("inventory = %s", got)
	}
	if got := s.Clients.URL("auth"); got != "http://auth:8080" {
		t.Errorf("auth = %s", got)
	}
}

func TestHandlerOptionsWithoutIdentityKeysFails(t *testing.T) {
	s := newService(t, "development")
	if err := s.init(); err != nil {
		t.Fatal(err)
	}
	if opts := s.HandlerOptions(); opts != nil || len(s.errs) != 1 {
		t.Errorf("opts=%v errs=%v, want a recorded error", opts, s.errs)
	}
}
