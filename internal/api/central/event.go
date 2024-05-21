package central

import (
	"context"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/lo"
)

func (s *service) TriggerEvent(ctx context.Context, request *centralpb.TriggerEventRequest) (*centralpb.TriggerEventReply, error) {
	s.logger.DebugContext(ctx, "received trigger event request", slog.Any("request", request))

	switch payload := request.Payload.Event.(type) {
	case *centralpb.TriggerEventPayload_EventInvaded:
		event := payload.EventInvaded
		s.logger.InfoContext(ctx, "received event: invaded", slog.String("type", "invaded"), attributext.EventMetadataPb(event.GetMetadata()))

		go func() {
			ctx := context.WithoutCancel(ctx)

			_, err := s.db.CreateInvadedEvent(ctx, &database.CreateInvadedEventRequest{
				Metadata: models.EventMetadataFromProto(event.Metadata),
				Invaders: lo.Map(event.GetInvaders(), func(invader *eventpb.Invader, _ int) database.InvaderImageRequest {
					return database.InvaderImageRequest{
						Picture:    invader.GetPicture(),
						Confidence: float64(invader.Confidence),
					}
				}),
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to create invaded event", slogext.Error(err), attributext.EventMetadataPb(event.GetMetadata()))
			}
		}()

	case *centralpb.TriggerEventPayload_EventMovement:
		event := payload.EventMovement
		s.logger.InfoContext(ctx, "received event: movement", slog.String("type", "movement"), attributext.EventMetadataPb(event.GetMetadata()))

		go func() {
			ctx := context.WithoutCancel(ctx)

			_, err := s.db.CreateMovementEvent(ctx, &database.CreateMovementEventRequest{
				Metadata:        models.EventMetadataFromProto(event.Metadata),
				MovementPicture: event.GetPicture(),
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to create movement event", slogext.Error(err), attributext.EventMetadataPb(event.GetMetadata()))
			}
		}()
	}

	return &centralpb.TriggerEventReply{}, nil
}
