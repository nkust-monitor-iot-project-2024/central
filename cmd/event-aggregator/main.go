package main

import (
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/services/event-aggregator"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fx.New(
		fx.WithLogger(utils.FxWithLoggerFn),
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("event-aggregator", "v0")),
		telemetry.FxModule,
		event_aggregator.FxModule,
	).Run()
}
