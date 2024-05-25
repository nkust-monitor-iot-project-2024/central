package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Storer struct {
	*Service

	tracer trace.Tracer
	logger *slog.Logger
}

func NewStorer(service *Service) (*Storer, error) {
	const name = "event-aggregator/storer"

	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	return &Storer{
		Service: service,
		tracer:  tracer,
		logger:  logger,
	}, nil
}

func (s *Storer) storeSingleEvent(ctx context.Context, event *eventpb.EventMessage, metadata models.Metadata) bool {
	ctx, span := s.tracer.Start(ctx, "event_storer/store_single_event")
	defer span.End()

	s.logger.DebugContext(ctx, "storing single event", slog.Any("metadata", metadata))

	status := false

	client := s.repo.Client()

	span.AddEvent("start transaction")
	tx, err := client.Tx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to start transaction")
		span.RecordError(err)

		return false
	}
	// commit Tx
	defer func(clientTx *ent.Tx) {
		switch status {
		case true:
			span.AddEvent("commit transaction")
			err := clientTx.Commit()
			if err != nil {
				span.SetStatus(codes.Error, "failed to commit transaction")
				span.RecordError(err)

				return
			}
			span.AddEvent("transaction committed")
		case false:
			span.AddEvent("rollback transaction")
			err := clientTx.Rollback()
			if err != nil {
				span.SetStatus(codes.Error, "failed to rollback transaction")
				span.RecordError(err)

				return
			}
			span.AddEvent("transaction rolled back")
		}

		span.SetStatus(codes.Ok, "transaction finished")
	}(tx)

	switch eventInfo := event.GetEvent().(type) {
	case *eventpb.EventMessage_MovementInfo:
		status = s.storeMovementEvent(ctx, tx, eventInfo.MovementInfo, metadata)

		span.AddEvent("created event",
			trace.WithAttributes(otelattrext.UUID("event_id", metadata.GetEventID())))
		return status
	default:
		span.SetStatus(codes.Error, "event type is not implemented")
		span.RecordError(fmt.Errorf("event type is not implemented: %T", event.GetEvent()))
		return false
	}
}

func (s *Storer) storeMovementEvent(ctx context.Context, transaction *ent.Tx, movementInfo *eventpb.MovementInfo, metadata models.Metadata) (status bool) {
	ctx, span := s.tracer.Start(ctx, "event_storer/store_movement_event")
	defer span.End()

	eventID := metadata.GetEventID()

	span.AddEvent("create event with movement information in database")
	eventModel, err := transaction.Event.Create().
		SetID(eventID).
		SetDeviceID(metadata.GetDeviceID()).
		SetCreatedAt(metadata.GetEmittedAt()).
		SetType("movement").
		Save(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create event with movement information in database")
		span.RecordError(err)

		return false
	}
	span.AddEvent("created event with movement information in database",
		trace.WithAttributes(otelattrext.UUID("event_id", eventModel.ID)))

	span.AddEvent("create movement information in database")
	movementModel, err := transaction.Movement.Create().
		AddEvent(eventModel).
		SetID(uuid.Must(uuid.NewV7())).
		SetPicture(movementInfo.GetPicture()).
		Save(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create movement information in database")
		span.RecordError(err)

		return false
	}
	span.AddEvent("created movement information in database",
		trace.WithAttributes(
			otelattrext.UUID("movement_id", movementModel.ID),
			otelattrext.UUID("event_id", eventModel.ID),
		))

	return true
}

func (s *Storer) Run(ctx context.Context, events <-chan mq.TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage]) {
	for event := range events {
		func() {
			ctx := trace.ContextWithSpanContext(ctx, event.SpanContext)
			ctx, span := s.tracer.Start(ctx, "event_storer/run/handle_event",
				trace.WithAttributes(
					otelattrext.UUID("event_id", event.Metadata.GetEventID()),
					attribute.String("device_id", event.Metadata.GetDeviceID())))
			defer span.End()

			if ok := s.storeSingleEvent(ctx, event.Body, event.Metadata); !ok {
				span.SetStatus(codes.Error, "failed to store event")
				_ = event.Reject(true)
			} else {
				span.SetStatus(codes.Ok, "event stored successfully")
				_ = event.Ack(false)
			}
		}()
	}
}
