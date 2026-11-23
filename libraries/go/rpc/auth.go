package rpc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
)

// ViaGatewayHeader marks a request the gateway forwarded from a client. The
// gateway sets it on every request it proxies (overwriting any client
// value), and it is never forwarded between services. A request without it
// came from inside the cluster: only the gateway is reachable from outside
// (infra/modules/environment security groups).
const ViaGatewayHeader = "X-Via-Gateway"

var (
	errUnauthenticated = connect.NewError(connect.CodeUnauthenticated, errors.New("Not signed in"))
	errInternalOnly    = connect.NewError(connect.CodePermissionDenied, errors.New("internal procedure; not callable through the gateway"))
)

// Authenticator is the handler interceptor that applies the access rules
// declared in the schema (rpc/proto/options/v1/options.proto):
//
//   - A valid identity token (X-Identity) puts the user in the context
//     (IdentityFrom).
//   - Procedures of `internal` services reject requests forwarded by the
//     gateway, so only other services can call them.
//   - Procedures not marked `public` require a user.
type Authenticator struct {
	users *Verifier
}

func NewAuthenticator(users *Verifier) *Authenticator {
	return &Authenticator{users: users}
}

func (a *Authenticator) authenticate(ctx context.Context, spec connect.Spec, header http.Header) (context.Context, error) {
	if isInternal(spec) && header.Get(ViaGatewayHeader) != "" {
		return nil, errInternalOnly
	}
	if token := header.Get(IdentityHeader); token != "" {
		if id, err := a.users.Verify(token); err == nil {
			ctx = context.WithValue(ctx, identityKey{}, id)
		}
	}
	if _, ok := IdentityFrom(ctx); !ok && !IsPublic(spec) {
		return nil, errUnauthenticated
	}
	return ctx, nil
}

func (a *Authenticator) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		ctx, err := a.authenticate(ctx, req.Spec(), req.Header())
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (a *Authenticator) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (a *Authenticator) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := a.authenticate(ctx, conn.Spec(), conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}
