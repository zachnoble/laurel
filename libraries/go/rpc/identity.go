package rpc

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

// Identity tokens are short-lived Ed25519 JWTs minted by the auth service.
// The gateway attaches one to each request that has a valid session cookie;
// services verify it locally with auth's public key. No service other than
// auth and the gateway ever sees the session cookie.

// IdentityHeader carries the identity token. The gateway strips any incoming value.
const IdentityHeader = "X-Identity"

const identityIssuer = "auth"

type Identity struct {
	UserID string
	Email  string

	token string // the raw token, so it can be forwarded downstream
}

type identityKey struct{}

// IdentityFrom returns the caller's identity, if the request carried a valid token.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

type identityClaims struct {
	Email string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// Signer mints identity tokens. Only the auth service holds the private key.
type Signer struct {
	key ed25519.PrivateKey
	ttl time.Duration
}

// NewSigner takes an Ed25519 private key (see keys.go and `just keygen`).
func NewSigner(privateKey string, ttl time.Duration) (*Signer, error) {
	key, err := ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("identity %w", err)
	}
	return &Signer{key: key, ttl: ttl}, nil
}

func (s *Signer) Sign(userID, email string) (string, error) {
	now := time.Now()
	claims := identityClaims{
		Email:     email,
		Issuer:    identityIssuer,
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(s.key)
}

// Verifier checks identity tokens. It accepts several public keys so the
// signing key can be rotated: add the new public key everywhere, switch auth
// to the new private key, then drop the old public key.
type Verifier struct {
	keys []ed25519.PublicKey
}

// NewVerifier takes base64-encoded 32-byte Ed25519 public keys.
func NewVerifier(publicKeys []string) (*Verifier, error) {
	if len(publicKeys) == 0 {
		return nil, errors.New("at least one identity public key is required")
	}
	v := &Verifier{}
	for _, k := range publicKeys {
		key, err := ParsePublicKey(k)
		if err != nil {
			return nil, fmt.Errorf("identity %w", err)
		}
		v.keys = append(v.keys, key)
	}
	return v, nil
}

func (v *Verifier) Verify(token string) (Identity, error) {
	var err error
	for _, key := range v.keys {
		var claims identityClaims
		_, err = jwt.ParseWithClaims(token, &claims,
			func(*jwt.Token) (any, error) { return key, nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
			jwt.WithIssuer(identityIssuer),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(30*time.Second),
		)
		if err == nil {
			return Identity{UserID: claims.Subject, Email: claims.Email, token: token}, nil
		}
	}
	return Identity{}, err
}

// ForwardIdentity is a client interceptor for service-to-service calls: it
// passes the current caller's identity token on to the downstream service.
//
//	ordersv1connect.NewOrdersServiceClient(httpClient, url, connect.WithInterceptors(rpc.ForwardIdentity{}))
type ForwardIdentity struct{}

func (ForwardIdentity) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if id, ok := IdentityFrom(ctx); ok && req.Spec().IsClient {
			req.Header().Set(IdentityHeader, id.token)
		}
		return next(ctx, req)
	}
}

func (ForwardIdentity) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		if id, ok := IdentityFrom(ctx); ok {
			conn.RequestHeader().Set(IdentityHeader, id.token)
		}
		return conn
	}
}

func (ForwardIdentity) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
