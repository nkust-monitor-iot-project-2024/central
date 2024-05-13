package central

import (
	"log/slog"

	"github.com/knadh/koanf/v2"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
)

type service struct {
	centralpb.UnimplementedCentralServer

	logger *slog.Logger
	config *koanf.Koanf
	db     database.Collection
}

func NewService(parentLogger *slog.Logger, conf *koanf.Koanf, database database.Collection) centralpb.CentralServer {
	return &service{
		logger: parentLogger.With(slog.String("service", "central")),
		config: conf,
		db:     database,
	}
}
