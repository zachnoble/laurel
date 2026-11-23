package rpc

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
)

func newKeys(t *testing.T, ttl time.Duration) (*Signer, *Verifier) {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	s, err := NewSigner(base64.StdEncoding.EncodeToString(priv.Seed()), ttl)
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewVerifier([]string{base64.StdEncoding.EncodeToString(pub)})
	if err != nil {
		t.Fatal(err)
	}
	return s, v
}

func TestInterceptorRequiresIdentityForNonPublicProcedures(t *testing.T) {
	signer, verifier := newKeys(t, time.Minute)
	i := NewAuthenticator(verifier)
	spec := connect.Spec{Procedure: "/orders.v1.OrdersService/GetOrder"} // no descriptor: not public

	if _, err := i.authenticate(context.Background(), spec, http.Header{}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("no token: err = %v, want Unauthenticated", err)
	}

	token, _ := signer.Sign("user-1", "a@example.com")
	ctx, err := i.authenticate(context.Background(), spec, http.Header{IdentityHeader: {token}})
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := IdentityFrom(ctx); !ok || id.UserID != "user-1" || id.Email != "a@example.com" {
		t.Fatalf("identity = %+v, %v", id, ok)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	signer, verifier := newKeys(t, -time.Minute) // already past the 30s leeway
	token, _ := signer.Sign("user-1", "")
	if _, err := verifier.Verify(token); err == nil {
		t.Fatal("expired token verified")
	}
}

func TestVerifierAcceptsAnyConfiguredKeyForRotation(t *testing.T) {
	oldSigner, oldVerifier := newKeys(t, time.Minute)
	_, newVerifier := newKeys(t, time.Minute)
	both := &Verifier{keys: append(newVerifier.keys, oldVerifier.keys...)}

	token, _ := oldSigner.Sign("user-1", "")
	if _, err := both.Verify(token); err != nil {
		t.Fatalf("token from old key rejected during rotation: %v", err)
	}
}
