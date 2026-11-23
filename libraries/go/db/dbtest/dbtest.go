// Package dbtest gives integration tests a fresh, migrated Postgres database.
//
//	func TestX(t *testing.T) {
//		pool := dbtest.New(t, db.Migrations)
//		...
//	}
//
// It creates a uniquely named database on the server at TEST_DATABASE_URL,
// applies the migrations, and drops it when the test ends, so tests can run
// in parallel against one server. Without TEST_DATABASE_URL the test is
// skipped locally (`just test-db` starts one) but fails in CI, so
// integration tests can't silently stop running.
package dbtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zachnoble/laurel/libraries/go/db"
)

func New(t testing.TB, migrations fs.FS) *pgxpool.Pool {
	t.Helper()
	server := os.Getenv("TEST_DATABASE_URL")
	if server == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL is not set")
		}
		t.Skip("TEST_DATABASE_URL not set; run `just test-db` for integration tests")
	}
	ctx := context.Background()

	admin, err := pgx.Connect(ctx, server)
	if err != nil {
		if os.Getenv("CI") == "" {
			t.Skipf("test database unreachable (%v); run `just test-db`", err)
		}
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(func() { admin.Close(context.Background()) })

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "test_" + hex.EncodeToString(b)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create database: %v", err)
	}

	u, err := url.Parse(server)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	dbURL := u.String()

	if migrations != nil {
		if err := db.Migrate(dbURL, migrations, slog.New(slog.DiscardHandler)); err != nil {
			t.Fatal(err)
		}
	}

	pool, err := db.Connect(ctx, db.Config{URL: dbURL, ApplicationName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})
	return pool
}
