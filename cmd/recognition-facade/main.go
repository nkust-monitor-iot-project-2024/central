package main

import (
	"log/slog"

	recognition_facade "github.com/nkust-monitor-iot-project-2024/central/internal/services/recognition-facade"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fx.New(
		fx.WithLogger(utils.FxWithLoggerFn),
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("recognition-facade", "v0")),
		telemetry.FxModule,
		recognition_facade.FxModule,
	).Run()
}
