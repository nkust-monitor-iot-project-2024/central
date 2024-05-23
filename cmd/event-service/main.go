package main

import (
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services/event"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		utils.FxInitLoggerModule,
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("event-service", "v0")),
		database.EntFx,
		telemetry.FxModule,
		mq.FxModule,
		event.FxModule,
	).Run()
}
