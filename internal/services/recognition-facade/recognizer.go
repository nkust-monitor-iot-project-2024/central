package recognition_facade

import (
	"context"
	"fmt"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/otelattrext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
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
	invadedEventsCh := make(chan *rawInvadedEvent, 64)
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		defer close(invadedEventsCh)

		for amqpConnectionResult := range r.amqp.NewConnectionSupplier(ctx) {
			connection, err := amqpConnectionResult.Get()
			if err != nil {
				slog.Error("failed to get AMQP connection", slogext.Error(err))
			}
			if err := r.recognizeTask(ctx, connection, invadedEventsCh); err != nil {
				slog.ErrorContext(ctx, "seems like connection has something wrong", slogext.Error(err))
			} else {
				slog.WarnContext(ctx, "recognizeTask long-running task seems stopped normally", slogext.Error(err))
			}
			_ = connection.CloseDeadline(time.Now().Add(2 * time.Second))
		}

		return ctx.Err()
	})

	group.Go(func() error {
		for connectionResult := range r.amqp.NewConnectionSupplier(ctx) {
			connection, err := connectionResult.Get()
			if err != nil {
				slog.Error("failed to get AMQP connection", slogext.Error(err))
			}
			connectionClosed := connection.NotifyClose(make(chan *amqp091.Error, 1))

		taskloop:
			for {
				select {
				case <-ctx.Done():
					break taskloop // clean up connection supplier
				case <-connectionClosed:
					break taskloop // renew connection
				case invadedEvent := <-invadedEventsCh:
					if err := r.triggerInvadedEvent(ctx, connection, invadedEvent); err != nil {
						slog.Error("failed to trigger invaded event", slogext.Error(err))
					}
				default:
				}
			}

			_ = connection.CloseDeadline(time.Now().Add(2 * time.Second))
		}

		return ctx.Err()
	})

}

func (r *Recognizer) recognizeTask(ctx context.Context, connection *amqp091.Connection, rawInvadedEventChan chan<- *rawInvadedEvent) error {
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

	key := mqevent.GetRoutingKey(mo.Some(models.EventTypeMovement))

	const queueName = "recognition-facade-recognizer-queue"
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

	const consumer = "subscribe-movement-event"

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

		rawInvadedInfo, err := r.processMovementEvent(ctx, eventDelivery)
		if err != nil {
			slog.ErrorContext(ctx, "failed to process movement event", slogext.Error(err))
			continue
		}
		if len(rawInvadedInfo.Invaders) == 0 {
			// no human found in the entities. ACK too.
			_ = eventDelivery.Ack()
			continue
		}

		select {
		case rawInvadedEventChan <- rawInvadedInfo:
			// successfully sent the invaded event to the channel.
			_ = eventDelivery.Ack()
		case <-ctx.Done():
			// system restarting; we should reject the event to prevent it from being lost.
			_ = eventDelivery.Reject(true)
		}
	}

	return nil
}

// processMovementEvent handles the movement event.
//
// This task processed the REJECT of the event; therefore,
// you do not need to reject the event again if an error occurs.
// However, if the event is processed successfully, you must ACK the event.
func (r *Recognizer) processMovementEvent(ctx context.Context, movementDelivery mqevent.AmqpEventMessageDelivery) (*rawInvadedEvent, error) {
	ctx = movementDelivery.Extract(ctx, r.propagator)
	ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/run/handle_event",
		trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	metadata, err := movementDelivery.Metadata()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get metadata from movement event")
		span.RecordError(err)

		_ = movementDelivery.Reject(false)
		return nil, fmt.Errorf("get metadata: %w", err)
	}

	span.SetAttributes(
		otelattrext.UUID("event_id", metadata.GetEventID()),
		attribute.String("device_id", metadata.GetDeviceID()),
	)

	body, err := movementDelivery.Body()
	if err != nil {
		span.SetStatus(codes.Error, "failed to get body from movement event")
		span.RecordError(err)

		_ = movementDelivery.Reject(false)
		return nil, fmt.Errorf("get body: %w", err)
	}

	r.handledMovementEvents.Add(ctx, 1)

	span.AddEvent("extracting movement information")
	movementEventID := metadata.GetEventID()

	span.AddEvent("recognizing entities in the image")
	movementInfo := body.GetMovementInfo()
	entity, err := r.recognizeEntities(ctx, movementInfo.GetPicture(), movementInfo.GetPictureMime())
	if err != nil {
		if code := status.Code(err); code == rpccodes.InvalidArgument {
			span.SetStatus(codes.Error, "users provides a unprocessable image")
			span.RecordError(err)

			_ = movementDelivery.Reject(false)
			return nil, fmt.Errorf("recognize entities: %w", err)
		}

		span.SetStatus(codes.Error, "failed to recognizeTask entities in the image")
		span.RecordError(err)

		_ = movementDelivery.Reject(true)
		return nil, fmt.Errorf("recognize entities: %w", err)
	}

	span.AddEvent("find human in entities")
	invaders, found := findHumanInEntities(entity)
	if !found {
		span.AddEvent("no human found in entities")

		return &rawInvadedEvent{
			ParentMovementID: movementEventID,
			Invaders:         nil,
		}, nil
	}

	span.AddEvent("found human in entities",
		trace.WithAttributes(
			attribute.Int("invaders", len(invaders))))
	span.SetStatus(codes.Ok, "recognized and passed as invader_event successfully")

	return &rawInvadedEvent{
		ParentMovementID: movementEventID,
		Invaders:         invaders,
	}, nil
}

// recognizeEntities calls the entityrecognitionpb.EntityRecognitionClient to recognizeTask the entities in the image.
func (r *Recognizer) recognizeEntities(ctx context.Context, image []byte, imageMime string) ([]*entityrecognitionpb.Entity, error) {
	ctx, span := r.tracer.Start(ctx, "recognizeEntities", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.AddEvent("call EntityRecognitionClient to recognition entities in the image")
	recognition, err := r.entityRecognitionClient.Recognize(ctx, &entityrecognitionpb.RecognizeRequest{
		Image:     image,
		ImageMime: imageMime,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to recognizeTask entities in the image")
		span.RecordError(err)

		return nil, fmt.Errorf("recognizeTask entities: %w", err)
	}
	span.AddEvent("done call EntityRecognitionClient to recognition entities in the image")

	span.SetStatus(codes.Ok, "recognized entities in the image")
	return recognition.Entities, nil
}

// rawInvadedEvent is the raw invaded information from processMovementEvent.
type rawInvadedEvent struct {
	ParentMovementID uuid.UUID
	Invaders         []*eventpb.Invader
}

// triggerInvadedEvent send the invaded event with the parent movement ID to Message Queue.
func (r *Recognizer) triggerInvadedEvent(ctx context.Context, connection *amqp091.Connection, event *rawInvadedEvent) error {
	ctx, span := r.tracer.Start(ctx, "recognition-facade/recognizer/trigger_invaded_event", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	r.triggeredInvadedEvents.Add(ctx, 1)

	span.AddEvent("prepare AMQP connection and exchange")

	channel, err := connection.Channel()
	if err != nil {
		span.SetStatus(codes.Error, "get channel")
		span.RecordError(err)

		return fmt.Errorf("get channel: %w", err)
	}
	defer func(channel *amqp091.Channel) {
		span.AddEvent("closing channel")
		err := channel.Close()
		if err != nil {
			span.SetStatus(codes.Error, "failed to close channel")
			span.RecordError(err)
		}
	}(channel)

	exchangeName, err := mqevent.DeclareEventsTopic(channel)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange")
		span.RecordError(err)

		return fmt.Errorf("declare exchange: %w", err)
	}

	span.AddEvent("publishing invaded event")
	metadata := models.Metadata{
		EventID:   uuid.Must(uuid.NewV7()),
		DeviceID:  "central/recognition-facade/recognizer",
		EmittedAt: time.Now(),
	}
	invadedEvent := &eventpb.EventMessage{
		Event: &eventpb.EventMessage_InvadedInfo{
			InvadedInfo: &eventpb.InvadedInfo{
				ParentMovementId: event.ParentMovementID.String(),
				Invaders:         event.Invaders,
			},
		},
	}
	eventType, publishing, err := mqevent.CreatePublishingEvent(ctx, mqevent.EventPublishingPayload{
		Propagator: r.propagator,
		Metadata:   metadata,
		Event:      invadedEvent,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to create publishing event")
		span.RecordError(err)

		return fmt.Errorf("failed to create publishing event: %w", err)
	}

	routingKey := mqevent.GetRoutingKey(mo.Some(eventType))

	err = channel.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false, /* mandatory */
		false, /* immediate */
		publishing,
	)
	if err != nil {
		span.SetStatus(codes.Error, "publish event")
		span.RecordError(err)

		return fmt.Errorf("publish event: %w", err)
	}

	span.AddEvent("published event")
	span.SetStatus(codes.Ok, "event published")

	return nil
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
