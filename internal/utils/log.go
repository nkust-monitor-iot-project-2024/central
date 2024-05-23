package utils

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var FxInitLoggerModule = fx.Module("init-logger",
	fx.Provide(fx.Annotate(NewInitLogger, fx.ResultTags(`name:"initLogger"`))),
	fx.WithLogger(fxWithLoggerFn),
)

type fxWithLoggerParam struct {
	fx.In

	Logger *slog.Logger `name:"initLogger"`
}

func fxWithLoggerFn(param fxWithLoggerParam) fxevent.Logger {
	return &fxevent.SlogLogger{
		Logger: param.Logger,
	}
}

func NewInitLogger() *slog.Logger {
	return slog.Default()
}

func NewLogger(name string) *slog.Logger {
	return otelslog.NewLogger(otelslog.WithInstrumentationScope(instrumentation.Scope{
		Name: name,
	}))
}
