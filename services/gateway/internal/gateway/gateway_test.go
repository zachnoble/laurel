package gateway

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/zachnoble/laurel/libraries/go/rpc"
	authv1 "github.com/zachnoble/laurel/rpc/generated/go/auth/v1"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
)

type fakeAuth struct {
	authv1connect.UnimplementedAuthInternalServiceHandler
	signer *rpc.Signer
	calls  atomic.Int32
}

func (f *fakeAuth) Authenticate(_ context.Context, req *connect.Request[authv1.AuthenticateRequest]) (*connect.Response[authv1.AuthenticateResponse], error) {
	f.calls.Add(1)
	if req.Msg.GetCookie() != "session_token=good" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no session"))
	}
	token, err := f.signer.Sign("user-1", "a@example.com")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&authv1.AuthenticateResponse{Token: token}), nil
}

type setup struct {
	gateway  *httptest.Server
	auth     *fakeAuth
	verifier *rpc.Verifier
}

func newSetup(t *testing.T) setup {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	signer, err := rpc.NewSigner(base64.StdEncoding.EncodeToString(priv.Seed()), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := rpc.NewVerifier([]string{base64.StdEncoding.EncodeToString(pub)})
	if err != nil {
		t.Fatal(err)
	}

	auth := &fakeAuth{signer: signer}
	mux := http.NewServeMux()
	mux.Handle(authv1connect.NewAuthInternalServiceHandler(auth))
	// Stand-in for any other service: echoes what the gateway forwarded.
	mux.HandleFunc("/orders.v1.OrdersService/GetOrder", func(w http.ResponseWriter, r *http.Request) {
		if id, err := verifier.Verify(r.Header.Get(rpc.IdentityHeader)); err == nil {
			w.Header().Set("Echo-User", id.UserID)
		}
		w.Header().Set("Echo-Cookie", r.Header.Get("Cookie"))
		w.Header().Set("Echo-Via-Gateway", r.Header.Get(rpc.ViaGatewayHeader))
	})

	upstream := httptest.NewUnstartedServer(mux)
	upstream.Config.Protocols = new(http.Protocols)
	upstream.Config.Protocols.SetUnencryptedHTTP2(true)
	upstream.Start()
	t.Cleanup(upstream.Close)

	client := authv1connect.NewAuthInternalServiceClient(rpc.HTTPClient(), upstream.URL)
	upstreamURL := func(string) string { return upstream.URL }
	gw := httptest.NewServer(New(upstreamURL, rpc.Transport(), "auth", NewSessions(client, "session_token", 1000, time.Minute)))
	t.Cleanup(gw.Close)

	return setup{gateway: gw, auth: auth, verifier: verifier}
}

func get(t *testing.T, url string, header http.Header) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	maps.Copy(req.Header, header)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return res
}

func TestSignedInRequestGetsIdentityAndNoCookie(t *testing.T) {
	s := newSetup(t)
	res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{"Cookie": {"session_token=good"}})

	if got := res.Header.Get("Echo-User"); got != "user-1" {
		t.Errorf("identity user = %q, want user-1", got)
	}
	if got := res.Header.Get("Echo-Cookie"); got != "" {
		t.Errorf("cookie forwarded to non-session service: %q", got)
	}
}

func TestClientCannotSpoofIdentity(t *testing.T) {
	s := newSetup(t)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	forger, _ := rpc.NewSigner(base64.StdEncoding.EncodeToString(priv.Seed()), time.Minute)
	forged, _ := forger.Sign("admin", "")

	res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{rpc.IdentityHeader: {forged}})
	if got := res.Header.Get("Echo-User"); got != "" {
		t.Errorf("spoofed identity reached upstream: %q", got)
	}
	if _, err := s.verifier.Verify(forged); err == nil {
		t.Error("token signed with another key verified")
	}
}

func TestBadCookieIsCachedAndAnonymous(t *testing.T) {
	s := newSetup(t)
	for range 3 {
		res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{"Cookie": {"session_token=bad"}})
		if got := res.Header.Get("Echo-User"); got != "" {
			t.Errorf("bad cookie produced identity %q", got)
		}
	}
	if n := s.auth.calls.Load(); n != 1 {
		t.Errorf("auth called %d times, want 1 (cached)", n)
	}
}

func TestOtherCookiesDontCallAuth(t *testing.T) {
	s := newSetup(t)
	res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{"Cookie": {"_ga=1; theme=dark"}})
	if got := res.Header.Get("Echo-User"); got != "" {
		t.Errorf("identity %q without a session cookie", got)
	}
	if n := s.auth.calls.Load(); n != 0 {
		t.Errorf("auth called %d times, want 0", n)
	}
}

func TestSessionCacheIgnoresOtherCookies(t *testing.T) {
	s := newSetup(t)
	for _, cookie := range []string{"session_token=good", "session_token=good; _ga=1", "_ga=2; session_token=good"} {
		res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{"Cookie": {cookie}})
		if got := res.Header.Get("Echo-User"); got != "user-1" {
			t.Errorf("%s: identity user = %q", cookie, got)
		}
	}
	if n := s.auth.calls.Load(); n != 1 {
		t.Errorf("auth called %d times, want 1", n)
	}
}

func TestSessionCacheIsBounded(t *testing.T) {
	auth := &fakeAuth{}
	mux := http.NewServeMux()
	mux.Handle(authv1connect.NewAuthInternalServiceHandler(auth))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	sessions := NewSessions(authv1connect.NewAuthInternalServiceClient(http.DefaultClient, srv.URL), "session_token", 10, time.Minute)
	for i := range 100 {
		if _, err := sessions.Token(context.Background(), fmt.Sprintf("random-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if n := sessions.cache.Len(); n > 10 {
		t.Errorf("cache holds %d entries, want <= 10", n)
	}
}

func TestConcurrentMissesShareOneAuthCall(t *testing.T) {
	release := make(chan struct{})
	auth := &blockingAuth{release: release}
	mux := http.NewServeMux()
	mux.Handle(authv1connect.NewAuthInternalServiceHandler(auth))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	sessions := NewSessions(authv1connect.NewAuthInternalServiceClient(http.DefaultClient, srv.URL), "session_token", 10, time.Minute)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() { _, _ = sessions.Token(context.Background(), "same") })
	}
	time.Sleep(100 * time.Millisecond)
	close(release)
	wg.Wait()
	if n := auth.calls.Load(); n != 1 {
		t.Errorf("auth called %d times, want 1", n)
	}
}

type blockingAuth struct {
	authv1connect.UnimplementedAuthInternalServiceHandler
	release chan struct{}
	calls   atomic.Int32
}

func (b *blockingAuth) Authenticate(context.Context, *connect.Request[authv1.AuthenticateRequest]) (*connect.Response[authv1.AuthenticateResponse], error) {
	b.calls.Add(1)
	<-b.release
	return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no session"))
}

func TestForwardedRequestsAreMarked(t *testing.T) {
	s := newSetup(t)
	// A client can't clear the marker (an empty value would look internal).
	res := get(t, s.gateway.URL+"/orders.v1.OrdersService/GetOrder", http.Header{rpc.ViaGatewayHeader: {""}})
	if got := res.Header.Get("Echo-Via-Gateway"); got == "" {
		t.Error("forwarded request not marked as coming through the gateway")
	}
}

func TestUnknownPathsAreNotRouted(t *testing.T) {
	s := newSetup(t)
	for _, path := range []string{"/", "/health/../x", "/Auth.v1.X/Y", "/order_history.v1.X/Y"} {
		if res := get(t, s.gateway.URL+path, nil); res.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, res.StatusCode)
		}
	}
}

func TestServiceNotRunningIsAConnectError(t *testing.T) {
	sessions := NewSessions(authv1connect.NewAuthInternalServiceClient(rpc.HTTPClient(), "http://127.0.0.1:1"), "session_token", 10, time.Minute)
	gw := httptest.NewServer(New(func(string) string { return "http://127.0.0.1:1" }, rpc.Transport(), "auth", sessions))
	t.Cleanup(gw.Close)

	res, err := http.Get(gw.URL + "/orders.v1.OrdersService/GetOrder")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var body struct{ Code, Message string }
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != http.StatusServiceUnavailable || body.Code != "unavailable" || !strings.Contains(body.Message, "just up orders") {
		t.Errorf("got %d %+v", res.StatusCode, body)
	}
}
