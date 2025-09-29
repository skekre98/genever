package logging

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	// Text in dev is easier to read; swap to JSON in prod later.
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(h)
}
