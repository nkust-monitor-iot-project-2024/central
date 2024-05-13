package utils

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	h := slog.NewTextHandler(os.Stderr, nil)
	return slog.New(h)
}
