package central

import (
	"context"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
)

func (s *Service) TriggerEvent(ctx context.Context, request *centralpb.TriggerEventRequest) (*centralpb.TriggerEventReply, error) {
	switch payload := request.Payload.Event.(type) {
	case *centralpb.TriggerEventPayload_EventInvaded:
		event := payload.EventInvaded
		s.logger.Info("received event: invaded", slogext.EventMetadataPb(event.GetMetadata()))
	}

	return &centralpb.TriggerEventReply{}, nil
}
