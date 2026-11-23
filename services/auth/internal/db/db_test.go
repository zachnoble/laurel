package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zachnoble/laurel/libraries/go/db/dbtest"
	"github.com/zachnoble/laurel/services/auth/internal/db"
	"github.com/zachnoble/laurel/services/auth/internal/db/sqlc"
)

func TestSessionsExpire(t *testing.T) {
	ctx := context.Background()
	q := db.New(dbtest.New(t, db.Migrations)).Queries

	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{ID: uuid.NewString(), Name: "A", Email: "a@example.com", EmailVerified: true})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for token, expires := range map[string]time.Time{"live": now.Add(time.Hour), "expired": now.Add(-time.Hour)} {
		if _, err := q.CreateSession(ctx, sqlc.CreateSessionParams{
			ID:        uuid.NewString(),
			Token:     token,
			UserID:    user.ID,
			ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}); err != nil {
			t.Fatal(err)
		}
	}

	if row, err := q.GetUserBySessionToken(ctx, "live"); err != nil || row.User.ID != user.ID {
		t.Fatalf("live session: %+v, %v", row, err)
	}
	if _, err := q.GetUserBySessionToken(ctx, "expired"); err == nil {
		t.Fatal("expired session resolved to a user")
	}
	if err := q.DeleteExpiredSessions(ctx); err != nil {
		t.Fatal(err)
	}
}
