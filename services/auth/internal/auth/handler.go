package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"github.com/zachnoble/laurel/libraries/go/rpc"
	authv1 "github.com/zachnoble/laurel/rpc/generated/go/auth/v1"
	"github.com/zachnoble/laurel/rpc/generated/go/auth/v1/authv1connect"
	"github.com/zachnoble/laurel/services/auth/internal/db/sqlc"
)

type Handler struct {
	service      *Service
	cookieSecret string
	signer       *rpc.Signer
}

var (
	_ authv1connect.AuthServiceHandler         = (*Handler)(nil)
	_ authv1connect.AuthInternalServiceHandler = (*Handler)(nil)
)

func NewHandler(service *Service, cookieSecret string, signer *rpc.Signer) *Handler {
	return &Handler{service: service, cookieSecret: cookieSecret, signer: signer}
}

func (h *Handler) Mount(mux *http.ServeMux, opts ...connect.HandlerOption) {
	mux.Handle(authv1connect.NewAuthServiceHandler(h, opts...))
	mux.Handle(authv1connect.NewAuthInternalServiceHandler(h, opts...))
}

// Authenticate is called by the gateway to exchange a session cookie for an
// identity token.
func (h *Handler) Authenticate(ctx context.Context, req *connect.Request[authv1.AuthenticateRequest]) (*connect.Response[authv1.AuthenticateResponse], error) {
	token, err := readToken(http.Header{"Cookie": {req.Msg.GetCookie()}}, h.cookieSecret)
	if err != nil {
		return nil, errUnauthorized
	}

	user, err := h.service.Authenticate(ctx, token)
	if err != nil {
		return nil, err
	}

	identity, err := h.signer.Sign(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("sign identity: %w", err)
	}
	return connect.NewResponse(&authv1.AuthenticateResponse{Token: identity}), nil
}

func (h *Handler) SendOtp(ctx context.Context, req *connect.Request[authv1.SendOtpRequest]) (*connect.Response[authv1.SendOtpResponse], error) {
	if err := h.service.SendOTP(ctx, req.Msg.GetEmail()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&authv1.SendOtpResponse{}), nil
}

func (h *Handler) VerifyOtp(ctx context.Context, req *connect.Request[authv1.VerifyOtpRequest]) (*connect.Response[authv1.VerifyOtpResponse], error) {
	user, session, err := h.service.VerifyOTP(ctx, req.Msg.GetEmail(), req.Msg.GetOtp(), req.Header().Get("User-Agent"))
	if err != nil {
		return nil, err
	}

	res := connect.NewResponse(&authv1.VerifyOtpResponse{User: toProtoUser(user)})
	res.Header().Add("Set-Cookie", sessionCookie(h.cookieSecret, session.Token, session.ExpiresAt).String())
	return res, nil
}

func (h *Handler) SignOut(ctx context.Context, req *connect.Request[authv1.SignOutRequest]) (*connect.Response[authv1.SignOutResponse], error) {
	if token, err := readToken(req.Header(), h.cookieSecret); err == nil {
		if err := h.service.SignOut(ctx, token); err != nil {
			return nil, err
		}
	}

	res := connect.NewResponse(&authv1.SignOutResponse{})
	res.Header().Add("Set-Cookie", expiredCookie().String())
	return res, nil
}

func (h *Handler) GetSession(ctx context.Context, req *connect.Request[authv1.GetSessionRequest]) (*connect.Response[authv1.GetSessionResponse], error) {
	token, err := readToken(req.Header(), h.cookieSecret)
	if err != nil {
		return connect.NewResponse(&authv1.GetSessionResponse{}), nil
	}

	user, err := h.service.Authenticate(ctx, token)
	if errors.Is(err, errUnauthorized) {
		return connect.NewResponse(&authv1.GetSessionResponse{}), nil
	}
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&authv1.GetSessionResponse{User: toProtoUser(user)}), nil
}

func toProtoUser(u *sqlc.User) *authv1.User {
	return &authv1.User{
		Id:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
	}
}
