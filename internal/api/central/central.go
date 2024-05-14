package central

import (
	"context"
	"log/slog"

	"github.com/knadh/koanf/v2"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
)

type service struct {
	centralpb.UnimplementedCentralServer

	logger *slog.Logger
	config *koanf.Koanf
	db     database.Database
}

func NewService(parentLogger *slog.Logger, conf *koanf.Koanf, database database.Database) centralpb.CentralServer {
	srv := &service{
		logger: parentLogger.With(slog.String("service", "central")),
		config: conf,
		db:     database,
	}

	// tasks: gc
	go func() {
		ctx := context.WithValue(context.Background(), "background_task", "gc")
		srv.gcEventLoop(ctx)
	}()

	return srv
}
