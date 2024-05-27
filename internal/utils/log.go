package utils

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.uber.org/fx/fxevent"
)

// FxWithLoggerFn provides the fxevent.Logger for the fx.WithLogger.
func FxWithLoggerFn() fxevent.Logger {
	return &fxevent.SlogLogger{
		Logger: slog.Default(),
	}
}

// NewLogger creates a new slog.Logger that sends logs to the collector service with the given name.
func NewLogger(name string) *slog.Logger {
	return otelslog.NewLogger(otelslog.WithInstrumentationScope(instrumentation.Scope{
		Name: name,
	}))
}
