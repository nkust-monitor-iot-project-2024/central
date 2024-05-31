package recognition_facade

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	rpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recognizer is the service that receives the image and recognizes the entities in the image.
type Recognizer struct {
	*Service

	tracer     trace.Tracer
	logger     *slog.Logger
	propagator propagation.TextMapPropagator

	handledMovementEvents  metric.Int64Counter
	triggeredInvadedEvents metric.Int64Counter
}

// NewRecognizer creates a new Recognizer.
func NewRecognizer(service *Service) (*Recognizer, error) {
	const name = "services/recognition/receiver"

	tracer := otel.GetTracerProvider().Tracer(name)
	meter := otel.GetMeterProvider().Meter(name)
	logger := utils.NewLogger(name)
	propagator := otel.GetTextMapPropagator()

	handledMovementEvents, err := meter.Int64Counter("handled_movement_events",
		metric.WithDescription("The number of handled events"),
		metric.WithUnit("{events}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create handled_movement_events counter: %w", err)
	}

	triggeredInvadedEvents, err := meter.Int64Counter("triggered_invaded_events",
		metric.WithDescription("The number of pushed triggered events"),
		metric.WithUnit("{events}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create triggered_invaded_events counter: %w", err)
	}

	return &Recognizer{
		Service:                service,
		tracer:                 tracer,
		logger:                 logger,
		propagator:             propagator,
		handledMovementEvents:  handledMovementEvents,
		triggeredInvadedEvents: triggeredInvadedEvents,
	}, nil
}

// Run retrieves the movement events continuously and blocks until the service is stopped by ctx.
func (r *Recognizer) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		delivery, err := r.messageQueue.SubscribeMovementEvent(ctx)
		if err != nil {
			slog.Error("failed to subscribe movement event", slogext.Error(err))
			return
		}

	deliveryloop:
		for {
			select {
			case <-ctx.Done():
			case <-delivery.ClosedChan:
				break deliveryloop

			case delivery, ok := <-delivery.DeliveryChan:
				if !ok {
					break deliveryloop
				}

				r.handleEvent(ctx, delivery)
			}
		}

		if err := delivery.Cleanup(); err != nil {
			slog.Debug("failed to cleanup the subscription", slogext.Error(err))
		}
	}
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

	r.triggeredInvadedEvents.Add(ctx, 1)

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

// handleEvent handles the movement event.
func (r *Recognizer) handleEvent(ctx context.Context, movementDelivery mq.TraceableMovementEventDelivery) {
	ctx = movementDelivery.Extract(ctx, r.propagator)
	ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/run/handle_event",
		trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	metadata, err := movementDelivery.Metadata()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get metadata from movement event")
		span.RecordError(err)

		_ = mq.Reject(movementDelivery, false)
		return
	}

	span.SetAttributes(
		otelattrext.UUID("event_id", metadata.GetEventID()),
		attribute.String("device_id", metadata.GetDeviceID()),
	)

	body, err := movementDelivery.Body()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get body from movement event")
		span.RecordError(err)

		_ = mq.Reject(movementDelivery, false)
		return
	}

	r.handledMovementEvents.Add(ctx, 1)

	span.AddEvent("extracing movement information")
	movementEventID := metadata.GetEventID()

	span.AddEvent("recognizing entities in the image")
	movementInfo := body.GetMovementInfo()
	entity, err := r.recognizeEntities(ctx, movementInfo.GetPicture(), movementInfo.GetPictureMime())
	if err != nil {
		if code := status.Code(err); code == rpccodes.InvalidArgument {
			span.SetStatus(codes.Error, "users provides a unprocessable image")
			span.RecordError(err)

			_ = mq.Reject(movementDelivery, true)
			return
		}

		span.SetStatus(codes.Error, "failed to recognize entities in the image")
		span.RecordError(err)

		_ = mq.Reject(movementDelivery, false)
		return
	}

	span.AddEvent("find human in entities")
	invaders, found := findHumanInEntities(entity)
	if !found {
		span.AddEvent("no human found in entities")
		span.SetStatus(codes.Ok, "recognized; no human found, no invasion.")

		_ = mq.Ack(movementDelivery)
		return
	}

	span.AddEvent("trigger invaded event")
	if err := r.triggerInvadedEvent(ctx, movementEventID, invaders); err != nil {
		span.SetStatus(codes.Error, "failed to trigger invaded event")

		_ = mq.Reject(movementDelivery, false)
		return
	}

	span.SetStatus(codes.Ok, "recognized and passed as invader_event successfully")
	_ = mq.Ack(movementDelivery)
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
