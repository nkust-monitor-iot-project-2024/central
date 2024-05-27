package recognition

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"go.opentelemetry.io/otel"
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
	recognition, err := r.entityRecognitionClient.Recognize(ctx, &entityrecognitionpb.RecognizeRequest{})
	if err != nil {
		span.SetStatus(codes.Error, "failed to recognize entities in the image")
		span.RecordError(err)

		return nil, fmt.Errorf("recognize entities: %w", err)
	}
	span.AddEvent("done call EntityRecognitionClient to recognition entities in the image")

	span.SetStatus(codes.Ok, "recognized entities in the image")
	return recognition.Entities, nil
}
