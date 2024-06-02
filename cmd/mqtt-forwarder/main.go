package main

import (
	"log/slog"

	mqtt_forwarder "github.com/nkust-monitor-iot-project-2024/central/internal/services/mqtt-forwarder"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fx.New(
		fx.WithLogger(utils.FxWithLoggerFn),
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("mqtt-forwarder", "v0")),
		telemetry.FxModule,
		mqtt_forwarder.FxModule,
	).Run()
}
