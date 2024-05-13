package central

import (
	"log/slog"

	"github.com/knadh/koanf/v2"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"google.golang.org/grpc"
)

type Service struct {
	centralpb.UnimplementedCentralServer

	logger *slog.Logger
	config *koanf.Koanf
	db     database.Collection
}

func NewService(parentLogger *slog.Logger, config *koanf.Koanf, database database.Collection) *Service {
	return &Service{
		logger: parentLogger.With(slog.String("service", "central")),
		config: config,
		db:     database,
	}
}

func (s *Service) Register(server *grpc.Server) {
	centralpb.RegisterCentralServer(server, s)
}

var _ centralpb.CentralServer = (*Service)(nil)
