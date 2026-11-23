package db

import (
	"embed"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zachnoble/laurel/services/auth/internal/db/sqlc"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrations are applied by `server migrate` (deploys) or AUTO_MIGRATE (dev).
var Migrations, _ = fs.Sub(migrations, "migrations")

type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func New(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool, Queries: sqlc.New(pool)}
}
