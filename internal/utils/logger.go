package utils

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})
	return slog.New(h)
}
