package db

import (
	"context"
	"log/slog"
	"time"
)

func (db *DB) RunCleanup(ctx context.Context, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := db.Queries.DeleteExpiredSessions(ctx); err != nil {
				logger.Error("cleanup: delete expired sessions", "error", err)
			}
			if err := db.Queries.DeleteExpiredVerifications(ctx); err != nil {
				logger.Error("cleanup: delete expired verifications", "error", err)
			}
		}
	}
}
