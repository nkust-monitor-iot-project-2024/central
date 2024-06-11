package event_aggregator

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"time"
)

// Storer is the service that retrieves Events and stores them in the database.
type Storer struct {
	*Service

	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	logger     *slog.Logger

	handledEvents         metric.Int64Counter
	handledMovementEvents metric.Int64Counter
	handledInvadedEvents  metric.Int64Counter
}

// NewStorer creates a new Storer.
func NewStorer(service *Service) (*Storer, error) {
	const name = "event-aggregator/storer"

	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(name)
	meter := otel.GetMeterProvider().Meter(name)
	logger := utils.NewLogger(name)

	handledEvents, err := meter.Int64Counter("handled_events",
		metric.WithDescription("The number of handled events"),
		metric.WithUnit("{events}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create handled_events counter: %w", err)
	}

	handledMovementEvents, err := meter.Int64Counter("handled_movement_events",
		metric.WithDescription("The number of handled movement events"),
		metric.WithUnit("{events}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create handled_movement_events counter: %w", err)
	}

	handledInvadedEvents, err := meter.Int64Counter("handled_invaded_events",
		metric.WithDescription("The number of handled invaded events"),
		metric.WithUnit("{events}"))
	if err != nil {
		return nil, fmt.Errorf("failed to create handled_invaded_events counter: %w", err)
	}

	return &Storer{
		Service:               service,
		propagator:            propagator,
		tracer:                tracer,
		logger:                logger,
		handledEvents:         handledEvents,
		handledMovementEvents: handledMovementEvents,
		handledInvadedEvents:  handledInvadedEvents,
	}, nil
}

// Run retrieves the events continuously from the channel until something wrong or the context is canceled.
func (s *Storer) Run(ctx context.Context) {
	for amqpConnectionResult := range s.amqp.NewConnectionSupplier(ctx) {
		connection, err := amqpConnectionResult.Get()
		if err != nil {
			slog.Error("failed to get AMQP connection", slogext.Error(err))
		}
		if err := s.runInConnection(ctx, connection); err != nil {
			slog.ErrorContext(ctx, "seems like connection has something wrong", slogext.Error(err))
		} else {
			slog.WarnContext(ctx, "runInConnection long-running task seems stopped normally", slogext.Error(err))
		}
		_ = connection.CloseDeadline(time.Now().Add(2 * time.Second))
	}
}

func (s *Storer) runInConnection(ctx context.Context, connection *amqp091.Connection) error {
	slog.InfoContext(ctx, "prepare channel, exchange, and queue for storing events")

	channel, err := connection.Channel()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get AMQP channel", slogext.Error(err))
		return fmt.Errorf("get amqp connection: %w", err)
	}
	defer func(channel *amqp091.Channel) {
		err := channel.Close()
		if err != nil {
			slog.ErrorContext(ctx, "failed to close channel", slogext.Error(err))
		}
	}(channel)

	exchangeName, err := mqevent.DeclareEventsTopic(channel)
	if err != nil {
		slog.ErrorContext(ctx, "failed to declare events topic exchange", slogext.Error(err))
		return fmt.Errorf("declare events topic exchange: %w", err)
	}

	key := mqevent.GetRoutingKey(mo.None[models.EventType]())

	const queueName = "event-aggregator-storer-queue"
	queue, err := channel.QueueDeclare(
		queueName,
		true,  /* durable */
		false, /* autoDelete */
		false, /* exclusive */
		false, /* noWait */
		mqv2.QueueDeclareArgs,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to declare the queue",
			slog.String("name", queueName),
			slogext.Error(err))
		return fmt.Errorf("declare queue: %w", err)
	}

	err = channel.QueueBind(
		queue.Name,
		key,
		exchangeName,
		false, /* noWait */
		nil,   /* args */
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bind the queue",
			slog.String("name", queue.Name),
			slog.String("exchange", exchangeName),
			slog.String("key", key),
			slogext.Error(err))
		return fmt.Errorf("bind queue: %w", err)
	}

	const consumer = "subscribe-event"

	slog.InfoContext(ctx, "consume events from the queue",
		slog.String("queue", queue.Name),
		slog.String("exchange", exchangeName),
		slog.String("key", key),
		slog.String("consumer", consumer))
	deliveries, err := channel.Consume(
		queue.Name,
		consumer,
		false, /* autoAck */
		false, /* exclusive */
		false, /* noLocal */
		false, /* noWait */
		nil,   /* args */
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to consume events from the queue",
			slog.String("queue", queue.Name),
			slog.String("exchange", exchangeName),
			slog.String("key", key),
			slog.String("consumer", consumer),
			slogext.Error(err))
		return fmt.Errorf("consume events: %w", err)
	}
	defer func() {
		err := channel.Cancel(consumer, false)
		if err != nil {
			slog.ErrorContext(ctx, "failed to cancel the consumer",
				slog.String("consumer", consumer),
				slogext.Error(err))
		}
	}()

	for delivery := range deliveries {
		// If the Acknowledger is nil, it means this message is not valid;
		// if this message is over the requeue limit, we should reject it.
		if delivery.Acknowledger == nil || mqv2.IsDeliveryOverRequeueLimit(delivery) {
			_ = delivery.Reject(false)
			continue
		}

		eventDelivery := mqevent.AmqpEventMessageDelivery{
			Delivery: delivery,
		}

		s.handleEvent(ctx, eventDelivery)
	}

	return nil
}

// handleEvent handles the event and calls the storeSingleEvent function to store the event in the database.
func (s *Storer) handleEvent(ctx context.Context, delivery mqevent.AmqpEventMessageDelivery) {
	ctx = delivery.Extract(ctx, s.propagator)
	ctx, span := s.tracer.Start(ctx, "event-aggregator/storer/handle_event")
	defer span.End()

	metadata, err := delivery.Metadata()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get metadata from event")
		span.RecordError(err)

		_ = delivery.Reject(false)
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

		_ = delivery.Reject(false)
		return
	}

	span.AddEvent("storing events")
	if ok := s.storeSingleEvent(ctx, body, metadata); !ok {
		span.SetStatus(codes.Error, "failed to store event")
		_ = delivery.Reject(false)
	} else {
		span.SetStatus(codes.Ok, "event stored successfully")
		_ = delivery.Ack()
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

	s.handledEvents.Add(ctx, 1)

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

	s.handledMovementEvents.Add(ctx, 1)

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

	s.handledInvadedEvents.Add(ctx, 1)

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
