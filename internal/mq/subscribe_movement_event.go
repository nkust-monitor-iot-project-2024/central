package mq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MovementEventMessage is the message type for movement events.
//
// It is basically the destructured version of eventpb.EventMessage.
type MovementEventMessage struct {
	MovementInfo *eventpb.MovementInfo
}

// GetMovementInfo returns the movement info of the movement event message.
func (m *MovementEventMessage) GetMovementInfo() *eventpb.MovementInfo {
	return m.MovementInfo
}

type MovementEventSubscriber interface {
	// SubscribeMovementEvent subscribes only the movement event messages.
	SubscribeMovementEvent(ctx context.Context) (<-chan TraceableTypedDelivery[models.Metadata, *MovementEventMessage], error)
}

func (mq *amqpMQ) SubscribeMovementEvent(ctx context.Context) (<-chan TraceableTypedDelivery[models.Metadata, *MovementEventMessage], error) {
	_, span := mq.tracer.Start(ctx, "mq/subscribe_movement_event")
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
		"movement_v1_events",
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
		"event.v1.movement",
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

	eventsCh := make(chan TraceableTypedDelivery[models.Metadata, *MovementEventMessage], 64)

	// handle raw messages
	span.AddEvent("create a goroutine to handle raw messages")
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

			slog.DebugContext(ctx, "SubscribeMovementEvent: waiting for raw message")
			rawMessage := <-rawMessageCh

			wg.Add(1)
			go func() {
				defer wg.Done()

				ctx, span := mq.tracer.Start(ctx, "mq/subscribe_movement_event/handle_raw_message", trace.WithAttributes(
					attribute.String("message_id", rawMessage.MessageId),
				))
				defer span.End()

				span.AddEvent("unmarshal raw event message")
				typedDelivery, err := mq.unmarshalRawEventMessage(ctx, rawMessage)
				if err != nil {
					span.SetStatus(codes.Error, "failed to unmarshal raw event message")
					span.RecordError(err)

					_ = rawMessage.Reject(false)
					return
				}
				span.AddEvent("done unmarshalled raw event message")

				span.AddEvent("process and send unmarshalled event message to eventsCh")

				// check if the delivery is Movement instance
				if event, ok := typedDelivery.Body.GetEvent().(*eventpb.EventMessage_MovementInfo); ok {
					eventsCh <- TraceableTypedDelivery[models.Metadata, *MovementEventMessage]{
						TypedDelivery: TypedDelivery[models.Metadata, *MovementEventMessage]{
							Delivery: typedDelivery.TypedDelivery.Delivery,
							Metadata: typedDelivery.TypedDelivery.Metadata,
							Body:     &MovementEventMessage{MovementInfo: event.MovementInfo},
						},
						SpanContext: typedDelivery.SpanContext,
					}
				} else {
					span.SetStatus(codes.Error, "failed to cast the event message to MovementInfo")

					_ = rawMessage.Reject(false)
					return
				}

				span.AddEvent("done processed and sent unmarshalled event message to eventsCh")
				span.SetStatus(codes.Ok, "handle succeed")
			}()
		}

		wg.Wait()
		close(eventsCh)
	}()

	span.SetStatus(codes.Ok, "created a subscriber channel of movement_v1_events")
	return eventsCh, nil
}