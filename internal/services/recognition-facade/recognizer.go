package recognition_facade

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
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
func (r *Recognizer) recognizeEntities(ctx context.Context, image []byte) ([]*entityrecognitionpb.Entity, error) {
	ctx, span := r.tracer.Start(ctx, "recognizeEntities")
	defer span.End()

	span.AddEvent("call EntityRecognitionClient to recognition entities in the image")
	recognition, err := r.entityRecognitionClient.Recognize(ctx, &entityrecognitionpb.RecognizeRequest{
		Image: image,
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

func (r *Recognizer) Run(ctx context.Context, movementEvents <-chan mq.TraceableTypedDelivery[models.Metadata, *mq.MovementEventMessage]) {
	for movementEvent := range movementEvents {
		func() {
			ctx := trace.ContextWithSpanContext(ctx, movementEvent.SpanContext)
			ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/run/handle_event",
				trace.WithAttributes(
					otelattrext.UUID("event_id", movementEvent.Metadata.GetEventID()),
					attribute.String("device_id", movementEvent.Metadata.GetDeviceID())))
			defer span.End()

			_ /*entity*/, err := r.recognizeEntities(ctx, movementEvent.Body.GetMovementInfo().GetPicture())
			if err != nil {
				span.SetStatus(codes.Error, "failed to recognize entities")
				_ = movementEvent.Reject(true)
				return
			}

			// wip

			span.SetStatus(codes.Ok, "recognized and passed as invader_event successfully")
			_ = movementEvent.Ack(false)
		}()
	}
}
