package recognition_facade

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Recognizer is the service that receives the image and recognizes the entities in the image.
type Recognizer struct {
	*Service

	tracer trace.Tracer
	logger *slog.Logger
}

// NewRecognizer creates a new Recognizer.
func NewRecognizer(service *Service) (*Recognizer, error) {
	const name = "services/recognition/receiver"

	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	return &Recognizer{
		Service: service,
		tracer:  tracer,
		logger:  logger,
	}, nil
}

// recognizeEntities calls the entityrecognitionpb.EntityRecognitionClient to recognize the entities in the image.
func (r *Recognizer) recognizeEntities(ctx context.Context, image []byte, imageMime string) ([]*entityrecognitionpb.Entity, error) {
	ctx, span := r.tracer.Start(ctx, "recognizeEntities", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.AddEvent("call EntityRecognitionClient to recognition entities in the image")
	recognition, err := r.entityRecognitionClient.Recognize(ctx, &entityrecognitionpb.RecognizeRequest{
		Image:     image,
		ImageMime: imageMime,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to recognize entities in the image")
		span.RecordError(err)

		return nil, fmt.Errorf("recognize entities: %w", err)
	}
	span.AddEvent("done call EntityRecognitionClient to recognition entities in the image")

	span.SetStatus(codes.Ok, "recognized entities in the image")
	return recognition.Entities, nil
}

// triggerInvadedEvent send the invaded event with the parent movement ID to Message Queue.
func (r *Recognizer) triggerInvadedEvent(ctx context.Context, parentMovementID uuid.UUID, invaders []*eventpb.Invader) error {
	ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/trigger_invaded_event", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	span.AddEvent("create invaded event")
	metadata := models.Metadata{
		EventID:   uuid.Must(uuid.NewV7()),
		DeviceID:  "central/recognition-facade/recognizer",
		EmittedAt: time.Now(),
	}
	invadedEvent := &eventpb.EventMessage{
		Event: &eventpb.EventMessage_InvadedInfo{
			InvadedInfo: &eventpb.InvadedInfo{
				ParentMovementId: parentMovementID.String(),
				Invaders:         invaders,
			},
		},
	}

	err := r.messageQueue.PublishEvent(ctx, metadata, invadedEvent)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create invaded event")
		span.RecordError(err)

		return fmt.Errorf("create invaded event: %w", err)
	}

	span.AddEvent("created invaded event")
	span.SetStatus(codes.Ok, "invaded event created successfully")

	return nil
}

// Run retrieves the movement events continuously and blocks until the service is stopped by ctx or something wrong.
func (r *Recognizer) Run(ctx context.Context, movementEvents <-chan mq.TraceableTypedDelivery[models.Metadata, *mq.MovementEventMessage]) {
	for movementEvent := range movementEvents {
		func() {
			ctx := trace.ContextWithSpanContext(ctx, movementEvent.SpanContext)
			ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/run/handle_event",
				trace.WithAttributes(
					otelattrext.UUID("event_id", movementEvent.Metadata.GetEventID()),
					attribute.String("device_id", movementEvent.Metadata.GetDeviceID())),
				trace.WithSpanKind(trace.SpanKindConsumer))
			defer span.End()

			span.AddEvent("extracing movement information")
			movementEventID := movementEvent.Metadata.GetEventID()

			span.AddEvent("recognizing entities in the image")
			movementInfo := movementEvent.Body.GetMovementInfo()
			entity, err := r.recognizeEntities(ctx, movementInfo.GetPicture(), movementInfo.GetPictureMime())
			if err != nil {
				span.SetStatus(codes.Error, "failed to recognize entities")
				_ = movementEvent.Reject(true)
				return
			}

			span.AddEvent("find human in entities")
			invaders, found := findHumanInEntities(entity)
			if !found {
				span.AddEvent("no human found in entities")
				span.SetStatus(codes.Ok, "recognized; no human found, no invasion.")
				_ = movementEvent.Ack(false)

				return
			}

			span.AddEvent("trigger invaded event")
			if err := r.triggerInvadedEvent(ctx, movementEventID, invaders); err != nil {
				span.SetStatus(codes.Error, "failed to trigger invaded event")
				_ = movementEvent.Reject(true)

				return
			}

			span.SetStatus(codes.Ok, "recognized and passed as invader_event successfully")
			_ = movementEvent.Ack(false)
		}()
	}
}

// findHumanInEntities finds if there is a human in the entities.
//
// It returns the non-identified (no UUID) eventpb.Invader, so we can wrap it in eventpb.InvadedInfo.
func findHumanInEntities(entities []*entityrecognitionpb.Entity) (invaders []*eventpb.Invader, found bool) {
	for _, entity := range entities {
		if entity.GetLabel() == "person" {
			invaders = append(invaders, &eventpb.Invader{
				Picture:     entity.GetImage(),
				PictureMime: entity.GetImageMime(),
				Confidence:  entity.GetConfidence(),
			})
		}
	}

	return invaders, len(invaders) > 0
}
