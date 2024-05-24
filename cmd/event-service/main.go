package main

import (
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services/event"
	"github.com/nkust-monitor-iot-project-2024/central/internal/telemetry"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.WithLogger(utils.FxWithLoggerFn),
		utils.ConfigFxModule,
		fx.Provide(utils.NewResourceBuilder("event-service", "v0")),
		telemetry.FxModule,
		database.EntFx,
		models.EventRepositoryEntFx,
		mq.FxModule,
		event.FxModule,
	).Run()
}
