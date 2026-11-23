package rpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"

	authv1 "github.com/zachnoble/laurel/rpc/generated/go/auth/v1"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
)

type recordingAuth struct {
	authv1connect.UnimplementedAuthServiceHandler
	sawTimeout atomic.Bool
}

func (f *recordingAuth) GetSession(_ context.Context, req *connect.Request[authv1.GetSessionRequest]) (*connect.Response[authv1.GetSessionResponse], error) {
	f.sawTimeout.Store(req.Header().Get("Connect-Timeout-Ms") != "")
	return connect.NewResponse(&authv1.GetSessionResponse{}), nil
}

func newClients(t *testing.T, h *recordingAuth) authv1connect.AuthServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(authv1connect.NewAuthServiceHandler(h))
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.Protocols = new(http.Protocols)
	srv.Config.Protocols.SetUnencryptedHTTP2(true)
	srv.Start()
	t.Cleanup(srv.Close)
	c := NewClients(ClientsConfig{URLTemplate: srv.URL, Timeout: time.Second})
	return NewClient(c, "auth", authv1connect.NewAuthServiceClient)
}

func TestClientsSendDeadline(t *testing.T) {
	h := &recordingAuth{}
	if _, err := newClients(t, h).GetSession(context.Background(), connect.NewRequest(&authv1.GetSessionRequest{})); err != nil {
		t.Fatal(err)
	}
	if !h.sawTimeout.Load() {
		t.Error("no default deadline sent")
	}
}

func TestOverridesReplaceTemplate(t *testing.T) {
	c := NewClients(ClientsConfig{URLTemplate: "http://{name}:8080", Overrides: map[string]string{"orders": "https://orders.dev.example.com"}})
	if got := c.URL("auth"); got != "http://auth:8080" {
		t.Errorf("auth = %s", got)
	}
	if got := c.URL("orders"); got != "https://orders.dev.example.com" {
		t.Errorf("orders = %s", got)
	}
}
