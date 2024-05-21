package utils

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

func NewLogger(name string) *slog.Logger {
	return otelslog.NewLogger(otelslog.WithInstrumentationScope(instrumentation.Scope{
		Name: name,
	}))
}
