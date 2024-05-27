package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// EventSubscriber is the interface for services to subscribe to event messages.
type EventSubscriber interface {
	// SubscribeEvent subscribes to the event messages.
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
		"all_v1_events",
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

	span.AddEvent("set QoS of the channel")
	if err := mq.channel.Qos(64, 0, false); err != nil {
		span.SetStatus(codes.Error, "set QoS failed")
		span.RecordError(err)

		return nil, fmt.Errorf("set QoS: %w", err)
	}
	span.AddEvent("done setting QoS of the channel")

	span.AddEvent("prepare raw message channel from AMQP")
	rawMessageCh, err := mq.channel.ConsumeWithContext(
		ctx,
		queue.Name,
		"",
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

	eventsCh := make(chan TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage], 64)

	// handle raw messages
	go func() {
		wg := sync.WaitGroup{}

		for {
			if err := ctx.Err(); err != nil {
				slog.Debug("fuck, why context is cancelled??")
				break
			}
			if mq.channel.IsClosed() {
				// If the channel is closed, we may receive a lot of
				// zero-value messages, which is not expected.
				slog.Debug("fuck, why channel is cancelled??", slog.Bool("ctxCancelled", ctx.Err() != nil))
				break
			}

			slog.DebugContext(ctx, "SubscribeEvent: waiting for raw message")
			rawMessage := <-rawMessageCh

			wg.Add(1)
			go func() {
				defer wg.Done()

				ctx, span := mq.tracer.Start(ctx, "mq/subscribe_event/handle_raw_message", trace.WithAttributes(
					attribute.String("message_id", rawMessage.MessageId),
				))
				defer span.End()

				span.AddEvent("delegating event to handleRawEventMessage")
				if err := mq.handleRawEventMessage(ctx, rawMessage, eventsCh); err != nil {
					span.SetStatus(codes.Error, "handle raw message failed")
					span.RecordError(err)

					_ = rawMessage.Reject(false)
				}

				span.SetStatus(codes.Ok, "handle succeed")
			}()
		}

		wg.Wait()
		close(eventsCh)
	}()

	return eventsCh, nil
}

// handleRawEventMessage handles the received amqp091.Delivery, process it, and send it to ch.
func (mq *amqpMQ) handleRawEventMessage(ctx context.Context, message amqp091.Delivery, ch chan<- TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage]) error {
	mq.logger.DebugContext(ctx, "handle raw message", slog.Any("message", message))

	ctx = mq.propagator.Extract(ctx, NewMessageCarrier(message))
	ctx, span := mq.tracer.Start(ctx, "mq/handle_raw_event_message")
	defer span.End()

	span.AddEvent("extract metadata from header")
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
	span.AddEvent("extracted metadata from header")

	event := &eventpb.EventMessage{}

	span.AddEvent("unmarshal message body")
	if err := protojson.Unmarshal(message.Body, event); err != nil {
		span.SetStatus(codes.Error, "unmarshal failed")
		span.RecordError(err)

		return fmt.Errorf("unmarshal: %w", err)
	}
	span.AddEvent("unmarshaled message body")

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		span.AddEvent("pass Delivery to channel for further processing")
		ch <- TraceableTypedDelivery[models.Metadata, *eventpb.EventMessage]{
			TypedDelivery: TypedDelivery[models.Metadata, *eventpb.EventMessage]{
				Delivery: message,
				Metadata: metadata,
				Body:     event,
			},
			SpanContext: span.SpanContext(),
		}
		span.AddEvent("passed Delivery to channel for further processing")

		span.SetStatus(codes.Ok, "done handling raw message")
		return nil
	}
}

// extractMetadataFromHeader extracts the models.Metadata from the AMQP header.
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
