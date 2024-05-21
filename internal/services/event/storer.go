package event

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const storerName = "services/event/storer"

var (
	storerTracer = otel.Tracer(storerName)
	storerMeter  = otel.Meter(storerName)
	storerLogger = utils.NewLogger(storerName)

	HandledEventsCounter = lo.Must(
		storerMeter.Int64Counter("event.handled",
			metric.WithDescription("The total number of events handled by the storer service"),
		),
	)
	InsertedEventsCounter = lo.Must(
		storerMeter.Int64Counter("event.recorded",
			metric.WithDescription("The total number of events that is successfully stored into the database"),
		),
	)
	FailedEventsCounter = lo.Must(
		storerMeter.Int64Counter("event.failed",
			metric.WithDescription("The total number of events that cannot stored into the database"),
		),
	)
)

func (s *Service) storeSingleEvent(ctx context.Context, event *eventpb.MovementEventMessage) bool {
	ctx, span := storerTracer.Start(ctx, "storeSingleEvent")
	defer span.End()

	HandledEventsCounter.Add(ctx, 1)

	metadata := event.GetMetadata()
	storerLogger.DebugContext(ctx, "storing single event", slog.Any("metadata", metadata))

	eventID, err := uuid.Parse(metadata.GetEventId())
	if err != nil {
		span.SetStatus(codes.Error, "failed to parse event ID")
		span.RecordError(err)

		return false
	}

	span.AddEvent("create movement information in database")
	movementModel, err := s.client.Movement.Create().SetPicture(event.GetPicture()).Save(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create movement information in database")
		span.RecordError(err)

		return false
	}
	span.AddEvent("created movement information in database",
		trace.WithAttributes(otelattrext.UUID("movement_id", movementModel.ID)))

	span.AddEvent("create event with movement information in database")
	err = s.client.Event.Create().
		SetID(eventID).
		SetDeviceID(metadata.GetDeviceId()).
		SetCreatedAt(metadata.GetEmittedAt().AsTime()).
		SetType("movement").
		AddMovements(movementModel).
		Exec(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create event with movement information in database")
		span.RecordError(err)

		return false
	}

	span.AddEvent("created event with movement information in database",
		trace.WithAttributes(otelattrext.UUID("event_id", eventID)))
	return true
}

func (s *Service) storeEventsTask(ctx context.Context, events <-chan *eventpb.MovementEventMessage) {
	for event := range events {
		storerLogger.InfoContext(ctx, "received event",
			slog.Any("metadata", event.GetMetadata()))
		if s.storeSingleEvent(ctx, event) {
			InsertedEventsCounter.Add(ctx, 1)
		} else {
			FailedEventsCounter.Add(ctx, 1)
		}
	}
}
