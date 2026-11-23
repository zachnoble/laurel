package db

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // pgx5:// driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrate applies every pending up migration in migrations (golang-migrate
// layout: NNNNNN_name.up.sql / .down.sql) to the database at url.
//
// Deploys run this as a one-off task before rolling out the new version
// (`server migrate`), while the old version is still serving, so every
// migration must work with the previous version of the code: add columns
// and tables first, drop them a release later. CI lints new migrations with
// squawk for operations that lock or break running code.
func Migrate(url string, migrations fs.FS, logger *slog.Logger) error {
	src, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL(url))
	if err != nil {
		return fmt.Errorf("open database for migrations: %w", err)
	}
	defer m.Close()

	before, _, _ := m.Version()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: %w", err)
	}
	after, dirty, _ := m.Version()
	logger.Info("migrations applied", "from", before, "to", after, "dirty", dirty)
	return nil
}

// migrateURL switches a postgres:// URL to golang-migrate's pgx v5 driver.
func migrateURL(url string) string {
	for _, scheme := range []string{"postgresql://", "postgres://"} {
		if rest, ok := strings.CutPrefix(url, scheme); ok {
			return "pgx5://" + rest
		}
	}
	return url
}
