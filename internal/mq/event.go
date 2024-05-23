package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/encoding/protojson"
)

type EventSubscriber interface {
	SubscribeEvent(ctx context.Context) (<-chan TypedDelivery[models.Metadata, *eventpb.EventMessage], error)
}

// TypedDelivery is a wrapper around amqp091.Delivery that includes the typed metadata and body.
type TypedDelivery[M any, B any] struct {
	amqp091.Delivery

	Metadata M
	Body     B
}

// SubscribeEvent subscribes to the event messages.
func (mq *amqpMQ) SubscribeEvent(ctx context.Context) (<-chan TypedDelivery[models.Metadata, *eventpb.EventMessage], error) {
	ctx, span := mq.tracer.Start(ctx, "subscribe_event")
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
		false,
		true,
		true,
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
		true,
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

	eventsCh := make(chan TypedDelivery[models.Metadata, *eventpb.EventMessage], runtime.NumCPU())

	// handle raw messages
	go func() {
		ctx, span := mq.tracer.Start(ctx, "subscribe_event/handle_raw_messages")
		defer span.End()

		span.AddEvent("start handling raw messages")

		defer func() {
			span.AddEvent("cleaning up the resource")
			close(eventsCh)
		}()

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

			slog.InfoContext(ctx, "waiting for raw message")

			rawMessage := <-rawMessageCh

			wg.Add(1)
			go func() {
				defer wg.Done()

				if err := mq.handleRawMessage(ctx, rawMessage, eventsCh); err != nil {
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

		span.AddEvent("finished handling raw messages")
	}()

	return eventsCh, nil
}

func (mq *amqpMQ) handleRawMessage(ctx context.Context, message amqp091.Delivery, ch chan<- TypedDelivery[models.Metadata, *eventpb.EventMessage]) error {
	slog.InfoContext(ctx, "handle raw message", slog.Any("mes", message))

	ctx = mq.propagator.Extract(ctx, NewMessageHeaderCarrier(message.Headers))
	ctx, span := mq.tracer.Start(ctx, "handle_raw_message")
	defer span.End()

	if message.Timestamp.IsZero() {
		// Skip the message if the timestamp is zero.
		_ = message.Reject(false)

		span.SetStatus(codes.Error, "zero timestamp")

		return errors.New("zero timestamp")
	}

	if message.ContentType != "application/json; schema=EventMessage" {
		// Skip the message if the content type is not JSON (schema=EventMessage)
		_ = message.Reject(false)

		span.SetStatus(codes.Error, "invalid content type")

		return errors.New("invalid content type")
	}

	metadata, err := extractMetadataFromHeader(message.Headers, message.Timestamp)
	if err != nil {
		span.SetStatus(codes.Error, "extract metadata failed")
		span.RecordError(err)

		return fmt.Errorf("extract metadata: %w", err)
	}

	var event *eventpb.EventMessage

	if err := protojson.Unmarshal(message.Body, event); err != nil {
		span.SetStatus(codes.Error, "unmarshal failed")
		span.RecordError(err)

		return fmt.Errorf("unmarshal: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		ch <- TypedDelivery[models.Metadata, *eventpb.EventMessage]{
			Delivery: message,
			Metadata: metadata,
			Body:     event,
		}
		return nil
	}
}

func extractMetadataFromHeader(headers amqp091.Table, eventTs time.Time) (models.Metadata, error) {
	eventID, ok := headers["event_id"].(string)
	if !ok {
		return models.Metadata{}, fmt.Errorf("missing or invalid event_id")
	}
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("parse event_id: %w", err)
	}

	deviceID, ok := headers["device_id"].(string)
	if !ok {
		return models.Metadata{}, fmt.Errorf("missing or invalid device_id")
	}

	return models.Metadata{
		EventID:   eventUUID,
		DeviceID:  deviceID,
		EmittedAt: eventTs,
	}, nil
}
