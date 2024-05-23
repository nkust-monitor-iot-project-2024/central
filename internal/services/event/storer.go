package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Storer struct {
	*Service

	tracer trace.Tracer
	meter  metric.Meter
	logger *slog.Logger

	handleEventsCounter   metric.Int64Counter
	insertedEventsCounter metric.Int64Counter
	failedEventsCounter   metric.Int64Counter
}

func NewStorer(service *Service) (*Storer, error) {
	const name = "services/event/storer"

	tracer := otel.GetTracerProvider().Tracer(name)
	meter := otel.GetMeterProvider().Meter(name)
	logger := utils.NewLogger(name)

	handleEventsCounter, err := meter.Int64Counter("event.handled",
		metric.WithDescription("The total number of events handled by the storer service"),
	)
	if err != nil {
		return nil, fmt.Errorf("create handled events counter: %w", err)
	}
	insertedEventsCounter, err := meter.Int64Counter("event.recorded",
		metric.WithDescription("The total number of events that is successfully stored into the database"),
	)
	if err != nil {
		return nil, fmt.Errorf("create inserted events counter: %w", err)
	}
	failedEventsCounter, err := meter.Int64Counter("event.failed",
		metric.WithDescription("The total number of events that cannot stored into the database"),
	)
	if err != nil {
		return nil, fmt.Errorf("create failed events counter: %w", err)
	}

	return &Storer{
		Service:               service,
		tracer:                tracer,
		meter:                 meter,
		logger:                logger,
		handleEventsCounter:   handleEventsCounter,
		insertedEventsCounter: insertedEventsCounter,
		failedEventsCounter:   failedEventsCounter,
	}, nil
}

func (s *Storer) storeSingleEvent(ctx context.Context, event *eventpb.EventMessage, metadata models.Metadata) bool {
	ctx, span := s.tracer.Start(ctx, "store_single_event")
	defer span.End()

	s.logger.DebugContext(ctx, "storing single event", slog.Any("metadata", metadata))

	switch eventInfo := event.GetEvent().(type) {
	case *eventpb.EventMessage_MovementInfo:
		status := s.storeMovementEvent(ctx, eventInfo.MovementInfo, metadata)

		span.AddEvent("created event",
			trace.WithAttributes(otelattrext.UUID("event_id", metadata.GetEventID())))
		return status
	default:
		span.SetStatus(codes.Error, "event type is not implemented")
		span.RecordError(fmt.Errorf("event type is not implemented: %T", event.GetEvent()))
		return false
	}
}

func (s *Storer) storeMovementEvent(ctx context.Context, movementInfo *eventpb.MovementInfo, metadata models.Metadata) bool {
	ctx, span := s.tracer.Start(ctx, "store_movement_event")
	defer span.End()

	eventID := metadata.GetEventID()

	span.AddEvent("create movement information in database")
	movementModel, err := s.client.Movement.Create().SetPicture(movementInfo.GetPicture()).Save(ctx)
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
		SetDeviceID(metadata.GetDeviceID()).
		SetCreatedAt(metadata.GetEmittedAt()).
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

func (s *Storer) Run(ctx context.Context, events <-chan mq.TypedDelivery[models.Metadata, *eventpb.EventMessage]) {
	for event := range events {
		s.logger.DebugContext(ctx, "received event",
			slog.Any("metadata", event.Metadata))

		s.handleEventsCounter.Add(ctx, 1)
		if s.storeSingleEvent(ctx, event.Body, event.Metadata) {
			s.insertedEventsCounter.Add(ctx, 1)
		} else {
			s.failedEventsCounter.Add(ctx, 1)
		}
	}
}
