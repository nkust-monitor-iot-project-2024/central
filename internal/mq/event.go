package mq

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/semaphore"
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
	_, span := mq.tracer.Start(ctx, "mq.SubscribeEvent._prepare")
	defer span.End()

	span.AddEvent("declare exchange of AMQP")
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
	span.AddEvent("declared exchange of AMQP")

	span.AddEvent("declare queue of AMQP")
	queue, err := mq.channel.QueueDeclare(
		"",
		false,
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
	span.AddEvent("declared queue of AMQP")

	span.AddEvent("bind queue to exchange of AMQP")
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
	span.AddEvent("bound queue to exchange of AMQP")

	span.AddEvent("prepare raw message channel from AMQP")
	consumer := "event-subscriber-" + uuid.New().String()
	rawMessageCh, err := mq.channel.ConsumeWithContext(
		ctx,
		queue.Name,
		consumer,
		true,
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
	span.AddEvent("prepared the raw channel")

	eventsCh := make(chan TypedDelivery[models.Metadata, *eventpb.EventMessage], runtime.NumCPU())

	// handle raw messages
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		sema := semaphore.NewWeighted(int64(runtime.NumCPU() - 1 /* for the main goroutine */))

		wg := sync.WaitGroup{}

		for {
			if err := sema.Acquire(ctx, 1); err != nil {
				break
			}

			select {
			case <-ctx.Done():
				break
			case rawMessage := <-rawMessageCh:
				wg.Add(1)

				go func() {
					defer sema.Release(1)
					defer wg.Done()

					if err := mq.handleRawMessage(ctx, rawMessage, eventsCh); err != nil {
						if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
							mq.logger.WarnContext(ctx, "handle raw message failed", slogext.Error(err))
						}
					}
				}()
			}
		}

		wg.Wait()

		// cleanup
		if err := mq.channel.Cancel(consumer, false); err != nil {
			mq.logger.WarnContext(ctx, "cancel consumer failed", slogext.Error(err))
		}
		close(eventsCh)
	}()

	return eventsCh, nil
}

func (mq *amqpMQ) handleRawMessage(ctx context.Context, message amqp091.Delivery, ch chan<- TypedDelivery[models.Metadata, *eventpb.EventMessage]) error {
	ctx = mq.propagator.Extract(ctx, NewMessageHeaderCarrier(message.Headers))
	ctx, span := mq.tracer.Start(ctx, "mq.handleRawMessage")
	defer span.End()

	metadata, err := extractMetadataFromHeader(message.Headers)
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

func extractMetadataFromHeader(headers amqp091.Table) (models.Metadata, error) {
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

	emittedAt, ok := headers["emitted_at"].(time.Time)
	if !ok {
		return models.Metadata{}, fmt.Errorf("missing or invalid emitted_at")
	}

	return models.Metadata{
		EventID:   eventUUID,
		DeviceID:  deviceID,
		EmittedAt: emittedAt,
	}, nil
}
