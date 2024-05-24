package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

type EventSubscriber interface {
	SubscribeEvent(ctx context.Context) (<-chan TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage], error)
}

// TypedDelivery is a wrapper around amqp091.Delivery that includes the typed metadata and body.
type TypedDelivery[M any, B any] struct {
	amqp091.Delivery

	Metadata M
	Body     B
}

// TraceableTypedDelivery is a wrapper around TypedDelivery that includes the trace context.
type TraceableTypedDelivery[M any, B any] struct {
	TypedDelivery[M, B]

	SpanContext trace.SpanContext
}

// SubscribeEvent subscribes to the event messages.
func (mq *amqpMQ) SubscribeEvent(ctx context.Context) (<-chan TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage], error) {
	_, span := mq.tracer.Start(ctx, "mq/subscribe_event")
	defer span.End()

	span.AddEvent("prepare AMQP queue")
	err := mq.channel.ExchangeDeclare(
		"events_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange failed")
		span.RecordError(err)

		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	queue, err := mq.channel.QueueDeclare(
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "declare queue failed")
		span.RecordError(err)

		return nil, fmt.Errorf("declare queue: %w", err)
	}

	err = mq.channel.QueueBind(
		queue.Name,
		"event.v1.*",
		"events_topic",
		false,
		nil)
	if err != nil {
		span.SetStatus(codes.Error, "bind queue failed")
		span.RecordError(err)

		return nil, fmt.Errorf("bind queue: %w", err)
	}
	span.AddEvent("prepared the AMQP queue")

	span.AddEvent("prepare raw message channel from AMQP")
	consumer := "event-subscriber-" + uuid.New().String()
	rawMessageCh, err := mq.channel.ConsumeWithContext(
		ctx,
		queue.Name,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "consume failed")
		span.RecordError(err)

		return nil, fmt.Errorf("consume: %w", err)
	}
	span.AddEvent("prepared raw message channel from AMQP")

	eventsCh := make(chan TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage], runtime.NumCPU())

	// handle raw messages
	go func() {
		wg := sync.WaitGroup{}

		for {
			if err := ctx.Err(); err != nil {
				break
			}
			if mq.channel.IsClosed() {
				// If the channel is closed, we may receive a lot of
				// zero-value messages, which is not expected.
				break
			}

			slog.DebugContext(ctx, "SubscribeEvent: waiting for raw message")
			rawMessage := <-rawMessageCh

			wg.Add(1)
			go func() {
				ctx, span := mq.tracer.Start(ctx, "mq/subscribe_event/handle_raw_message", trace.WithAttributes(
					attribute.String("message_id", rawMessage.MessageId),
				))
				defer span.End()

				defer wg.Done()

				if err := mq.handleRawEventMessage(ctx, rawMessage, eventsCh); err != nil {
					if err := rawMessage.Reject(false); err != nil {
						mq.logger.WarnContext(ctx, "reject raw message failed", slogext.Error(err))
					}
				} else {
					if err := rawMessage.Ack(false); err != nil {
						mq.logger.WarnContext(ctx, "ack raw message failed", slogext.Error(err))
					}
				}

			}()
		}

		wg.Wait()
		close(eventsCh)
	}()

	return eventsCh, nil
}

func (mq *amqpMQ) handleRawEventMessage(ctx context.Context, message amqp091.Delivery, ch chan<- TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage]) error {
	mq.logger.DebugContext(ctx, "handle raw message", slog.Any("message", message))

	ctx = mq.propagator.Extract(ctx, NewMessageHeaderCarrier(message))
	ctx, span := mq.tracer.Start(ctx, "mq/handle_raw_event_message")
	defer span.End()

	if message.ContentType != "application/json" || message.Type != "eventpb.EventMessage" || message.Timestamp.IsZero() {
		span.SetStatus(codes.Error, "invalid header")

		return errors.New("invalid header")
	}

	metadata, err := extractMetadataFromHeader(message)
	if err != nil {
		span.SetStatus(codes.Error, "extract metadata failed")
		span.RecordError(err)

		return fmt.Errorf("extract metadata: %w", err)
	}

	event := &eventpb.EventMessage{}

	if err := protojson.Unmarshal(message.Body, event); err != nil {
		span.SetStatus(codes.Error, "unmarshal failed")
		span.RecordError(err)

		return fmt.Errorf("unmarshal: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		ch <- TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage]{
			TypedDelivery: TypedDelivery[models.Metadata, *eventpb.EventMessage]{
				Delivery: message,
				Metadata: metadata,
				Body:     event,
			},
			SpanContext: span.SpanContext(),
		}
		return nil
	}
}

func extractMetadataFromHeader(delivery amqp091.Delivery) (models.Metadata, error) {
	eventID := delivery.MessageId
	if eventID == "" {
		return models.Metadata{}, errors.New("missing or invalid MessageId (-> event_id)")
	}
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("parse event_id: %w", err)
	}

	deviceID := delivery.AppId
	if deviceID == "" {
		return models.Metadata{}, errors.New("missing or invalid AppId (-> device_id)")
	}

	eventTs := delivery.Timestamp

	return models.Metadata{
		EventID:   eventUUID,
		DeviceID:  deviceID,
		EmittedAt: eventTs,
	}, nil
}
