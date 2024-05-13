package central

import (
	"context"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"google.golang.org/grpc"
)

type Service struct {
	centralpb.UnimplementedCentralServer

	logger *slog.Logger
}

func NewService(parentLogger *slog.Logger) *Service {
	return &Service{
		logger: parentLogger.With(slog.String("service", "central")),
	}
}

func (s *Service) Register(server *grpc.Server) {
	centralpb.RegisterCentralServer(server, s)
}

func (s *Service) TriggerEvent(ctx context.Context, request *centralpb.TriggerEventRequest) (*centralpb.TriggerEventReply, error) {
	switch payload := request.Payload.Event.(type) {
	case *centralpb.TriggerEventPayload_EventInvaded:
		event := payload.EventInvaded
		s.logger.Info("received event: invaded", slogext.EventMetadataPb(event.GetMetadata()))
	}

	return &centralpb.TriggerEventReply{}, nil
}

var _ centralpb.CentralServer = (*Service)(nil)
