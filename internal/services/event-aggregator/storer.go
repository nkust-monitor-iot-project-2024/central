package event_aggregator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Storer is the service that retrieves Event from mq.MessageQueue
// and stores them in the database.
type Storer struct {
	*Service

	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	logger     *slog.Logger
}

// NewStorer creates a new Storer.
func NewStorer(service *Service) (*Storer, error) {
	const name = "event-aggregator/storer"

	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	return &Storer{
		Service:    service,
		propagator: propagator,
		tracer:     tracer,
		logger:     logger,
	}, nil
}

// Run retrieves the events continuously from the channel until something wrong or the context is canceled.
func (s *Storer) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		delivery, err := s.messageQueue.SubscribeEvent(ctx)
		if err != nil {
			slog.Error("failed to subscribe event", slogext.Error(err))
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

				s.handleEvent(ctx, delivery)
			}
		}

		if err := delivery.Cleanup(); err != nil {
			slog.Error("failed to cleanup the subscription", slogext.Error(err))
		}
	}
}

// handleEvent handles the event and calls the storeSingleEvent function to store the event in the database.
func (s *Storer) handleEvent(ctx context.Context, delivery mq.TraceableEventDelivery) {
	ctx = delivery.Extract(ctx, s.propagator)
	ctx, span := s.tracer.Start(ctx, "event-aggregator/storer/handle_event")
	defer span.End()

	metadata, err := delivery.Metadata()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get metadata from event")
		span.RecordError(err)

		_ = mq.Reject(delivery, false)
		return
	}
	span.SetAttributes(
		otelattrext.UUID("event_id", metadata.GetEventID()),
		attribute.String("device_id", metadata.GetDeviceID()),
	)

	body, err := delivery.Body()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get body from event")
		span.RecordError(err)

		_ = mq.Reject(delivery, false)
		return
	}

	span.AddEvent("storing events")
	if ok := s.storeSingleEvent(ctx, body, metadata); !ok {
		span.SetStatus(codes.Error, "failed to store event")
		_ = mq.Reject(delivery, true)
	} else {
		span.SetStatus(codes.Ok, "event stored successfully")
		_ = mq.Ack(delivery)
	}
}

// storeSingleEvent stores a single event in the database.
//
// It creates the transaction in the database connection, delegates the transaction
// and the event to the corresponding store functions (storeMovementEvent),
// and commits the transaction.
func (s *Storer) storeSingleEvent(ctx context.Context, event *eventpb.EventMessage, metadata models.Metadata) bool {
	ctx, span := s.tracer.Start(ctx, "event-aggregator/storer/store_single_event", trace.WithSpanKind(trace.SpanKindInternal))
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
		span.SetStatus(codes.Ok, "event stored successfully")

		return status
	case *eventpb.EventMessage_InvadedInfo:
		status = s.storeInvadedEvent(ctx, tx, eventInfo.InvadedInfo, metadata)

		span.AddEvent("created event",
			trace.WithAttributes(otelattrext.UUID("event_id", metadata.GetEventID())))
		span.SetStatus(codes.Ok, "event stored successfully")

		return status
	default:
		span.SetStatus(codes.Error, "event type is not implemented")
		span.RecordError(fmt.Errorf("event type is not implemented: %T", event.GetEvent()))
		return false
	}
}

// storeMovementEvent stores the movement event in the database.
func (s *Storer) storeMovementEvent(ctx context.Context, transaction *ent.Tx, movementInfo *eventpb.MovementInfo, metadata models.Metadata) bool {
	ctx, span := s.tracer.Start(ctx, "event-aggregator/storer/store_movement_event", trace.WithSpanKind(trace.SpanKindInternal))
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
		SetPictureMime(movementInfo.GetPictureMime()).
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

	span.SetStatus(codes.Ok, "created movement information in database")

	return true
}

// storeInvadedEvent stores the invaded event in the database.
func (s *Storer) storeInvadedEvent(ctx context.Context, transaction *ent.Tx, invadedInfo *eventpb.InvadedInfo, metadata models.Metadata) (status bool) {
	ctx, span := s.tracer.Start(ctx, "event-aggregator/storer/store_invaded_event", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.AddEvent("create event with invaded information in database")
	eventModelBuilder := transaction.Event.Create().
		SetID(metadata.GetEventID()).
		SetDeviceID(metadata.GetDeviceID()).
		SetCreatedAt(metadata.GetEmittedAt()).
		SetType("invaded")

	if parentEventID, hasParent := metadata.GetParentEventID(); hasParent {
		eventModelBuilder = eventModelBuilder.SetParentEventID(parentEventID)
	}

	eventModel, err := eventModelBuilder.Save(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create event with invaded information in database")
		span.RecordError(err)

		return false
	}
	span.AddEvent("created event with invaded information in database",
		trace.WithAttributes(otelattrext.UUID("event_id", eventModel.ID)))

	span.AddEvent("create invader information in database")

	invaderBuilders := make([]*ent.InvaderCreate, 0, len(invadedInfo.GetInvaders()))
	for _, invader := range invadedInfo.GetInvaders() {
		invaderBuilder := transaction.Invader.Create().
			AddEvent(eventModel).
			SetID(uuid.Must(uuid.NewV7())).
			SetPicture(invader.GetPicture()).
			SetPictureMime(invader.GetPictureMime()).
			SetConfidence(float64(invader.GetConfidence()))

		invaderBuilders = append(invaderBuilders, invaderBuilder)
	}

	createdInvaders, err := transaction.Invader.CreateBulk(invaderBuilders...).Save(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create invader information in database")
		span.RecordError(err)

		return false
	}
	span.AddEvent("created invader information in database",
		trace.WithAttributes(
			attribute.StringSlice("invader_ids", lo.Map(createdInvaders, func(invader *ent.Invader, _ int) string {
				return invader.ID.String()
			})),
			otelattrext.UUID("event_id", eventModel.ID),
		))

	span.SetStatus(codes.Ok, "created invaded information in database")

	return true
}
