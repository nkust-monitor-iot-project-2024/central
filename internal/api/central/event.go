package central

import (
	"bytes"
	"context"
	"log/slog"
	"path"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *service) TriggerEvent(ctx context.Context, request *centralpb.TriggerEventRequest) (*centralpb.TriggerEventReply, error) {
	s.logger.DebugContext(ctx, "received trigger event request", slog.Any("request", request))

	switch payload := request.Payload.Event.(type) {
	case *centralpb.TriggerEventPayload_EventInvaded:
		go func() {
			err := s.processInvadedEvent(context.WithoutCancel(ctx), payload.EventInvaded)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to process invaded event", slogext.Error(err))
			}
		}()
	}

	return &centralpb.TriggerEventReply{}, nil
}

func (s *service) processInvadedEvent(ctx context.Context, event *eventpb.EventInvaded) error {
	s.logger.InfoContext(ctx, "received event: invaded", slog.String("type", "invaded"), slogext.EventMetadataPb(event.GetMetadata()))

	eventID := primitive.NewObjectID()
	invaders := make([]primitive.ObjectID, 0, len(event.Invaders))

	// upload the invader to the GridFS
	for _, invader := range event.Invaders {
		invaderID := primitive.NewObjectID()
		picture := invader.GetPicture()

		err := s.db.Fs().UploadFromStreamWithID(
			invaderID,
			path.Join("/events", eventID.Hex(), "invader", invaderID.Hex()),
			bytes.NewReader(picture),
			options.GridFSUpload().SetMetadata(bson.M{
				"eventID": eventID,
			}),
		)
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"failed to upload invader picture",
				slogext.Error(err),
				slogext.ObjectID("eventID", eventID),
			)
			continue
		}

		invaders = append(invaders, invaderID)
	}

	_, err := s.db.Events().InsertOne(ctx, models.Event{
		ID:       eventID,
		Metadata: models.EventMetadataFromProto(event.Metadata),
		Type:     "invaded",
		Payload: models.EventPayload{
			InvaderPicture: invaders,
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to insert event", slogext.Error(err))
		return err
	}

	return nil
}
