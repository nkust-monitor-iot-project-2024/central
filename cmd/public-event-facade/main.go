package main

import (
	"log/slog"

	public_event_facade "github.com/nkust-monitor-iot-project-2024/central/internal/services/public-event-facade"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fx.New(
		fx.WithLogger(utils.FxWithLoggerFn),
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("public-event-facade", "v0")),
		telemetry.FxModule,
		public_event_facade.FxModule,
	).Run()
}
