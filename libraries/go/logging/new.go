package logging

import (
	"log/slog"
	"os"
)

// New returns the standard JSON logger, writing to stdout.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
