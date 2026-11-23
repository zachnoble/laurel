package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/resend/resend-go/v3"

	"github.com/zachnoble/laurel/libraries/go/logging"
	"github.com/zachnoble/laurel/services/auth/internal/config"
	"github.com/zachnoble/laurel/services/auth/internal/db"
	"github.com/zachnoble/laurel/services/auth/internal/db/sqlc"
)

const (
	otpTTL         = 15 * time.Minute
	sessionTTL     = 30 * 24 * time.Hour
	maxOTPAttempts = 5
)

var (
	errInvalidOTP   = connect.NewError(connect.CodeUnauthenticated, errors.New("Invalid or expired code"))
	errUnauthorized = connect.NewError(connect.CodeUnauthenticated, errors.New("Session not found"))
)

type Session struct {
	Token     string
	ExpiresAt time.Time
}

type Service struct {
	db     *db.DB
	resend *resend.Client
	cfg    *config.Config
}

func New(db *db.DB, resendClient *resend.Client, cfg *config.Config) *Service {
	return &Service{db: db, resend: resendClient, cfg: cfg}
}

func (s *Service) SendOTP(ctx context.Context, emailAddr string) error {
	otp, err := generateOTP()
	if err != nil {
		return err
	}

	_, err = s.db.Queries.CreateVerification(ctx, sqlc.CreateVerificationParams{
		ID:         uuid.NewString(),
		Identifier: emailAddr,
		Value:      hashOTP(otp),
		ExpiresAt:  pgtype.Timestamptz{Time: time.Now().Add(otpTTL), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	_, err = s.resend.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.cfg.EmailFromAddress,
		To:      []string{emailAddr},
		Subject: "Your verification code",
		Html:    fmt.Sprintf("<p>Your sign-in code is: <strong>%s</strong></p><p>This code expires in %d minutes.</p>", otp, int(otpTTL.Minutes())),
	})
	return err
}

func (s *Service) VerifyOTP(ctx context.Context, emailAddr, otp, userAgent string) (*sqlc.User, *Session, error) {
	verification, err := s.db.Queries.GetLatestVerificationByIdentifier(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, errInvalidOTP
		}
		return nil, nil, fmt.Errorf("get verification: %w", err)
	}

	if time.Now().After(verification.ExpiresAt.Time) || verification.Attempts >= maxOTPAttempts {
		logging.FromContext(ctx).Warn("otp verification rejected", "email", emailAddr, "attempts", verification.Attempts)
		return nil, nil, errInvalidOTP
	}

	if !otpMatches(otp, verification.Value) {
		if err := s.db.Queries.IncrementVerificationAttempts(ctx, verification.ID); err != nil {
			return nil, nil, fmt.Errorf("increment verification attempts: %w", err)
		}
		logging.FromContext(ctx).Warn("otp mismatch", "email", emailAddr, "attempts", verification.Attempts+1)
		return nil, nil, errInvalidOTP
	}

	if err := s.db.Queries.DeleteVerificationsByIdentifier(ctx, emailAddr); err != nil {
		return nil, nil, fmt.Errorf("delete verification: %w", err)
	}

	user, err := s.findOrCreateUser(ctx, emailAddr)
	if err != nil {
		return nil, nil, err
	}

	session, err := s.createSession(ctx, user.ID, userAgent)
	if err != nil {
		return nil, nil, err
	}

	logging.FromContext(ctx).Info("user signed in", "userId", user.ID, "email", emailAddr)

	return user, session, nil
}

func (s *Service) findOrCreateUser(ctx context.Context, emailAddr string) (*sqlc.User, error) {
	user, err := s.db.Queries.GetUserByEmail(ctx, emailAddr)
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get user: %w", err)
	}

	user, err = s.db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:            uuid.NewString(),
		Name:          emailAddr,
		Email:         emailAddr,
		EmailVerified: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	logging.FromContext(ctx).Info("user created", "userId", user.ID, "email", emailAddr)

	return &user, nil
}

func (s *Service) createSession(ctx context.Context, userID, userAgent string) (*Session, error) {
	token := rand.Text()
	now := time.Now()
	expiresAt := now.Add(sessionTTL)

	_, err := s.db.Queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        uuid.NewString(),
		Token:     token,
		UserID:    userID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		UserAgent: nonEmpty(userAgent),
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &Session{Token: token, ExpiresAt: expiresAt}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*sqlc.User, error) {
	row, err := s.db.Queries.GetUserBySessionToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errUnauthorized
		}
		logging.FromContext(ctx).Error("authenticate: get session", "error", err)
		return nil, fmt.Errorf("get session: %w", err)
	}

	return &row.User, nil
}

func (s *Service) SignOut(ctx context.Context, token string) error {
	if err := s.db.Queries.DeleteSessionByToken(ctx, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
