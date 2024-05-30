package mq

import (
	"context"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MovementEventSubscriber is the interface for services to subscribe to movement event messages.
type MovementEventSubscriber interface {
	// SubscribeMovementEvent subscribes only the movement event messages.
	SubscribeMovementEvent(ctx context.Context) (deliveryChan <-chan TraceableMovementEventDelivery, cleanup func() error, err error)
}

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

// TraceableMovementEventDelivery is the interface for the movement event message delivery that includes
// the metadata extraction, the span extractor, and the body extraction.
type TraceableMovementEventDelivery interface {
	TraceableEventDeliveryMetadata

	// Body returns the parsed movement event message body.
	Body() (MovementEventMessage, error)
}

func (mq *amqpMQ) SubscribeMovementEvent(ctx context.Context) (deliveryChan <-chan TraceableMovementEventDelivery, cleanup func() error, err error) {
	const consumer = "subscribe-movement-event"

	_, span := mq.tracer.Start(ctx, "mq/subscribe_movement_event", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	cs := utils.NewCleanupStack()

	span.AddEvent("prepare AMQP [subscribe] channel")
	mqConnection, err := mq.getSingletonSubscribeConnection()
	if err != nil {
		span.SetStatus(codes.Error, "get subscribe channel failed")
		span.RecordError(err)

		return nil, cs.Cleanup, fmt.Errorf("get subscribe channel: %w", err)
	}
	mqChannel, err := mqConnection.Channel()
	if err != nil {
		span.SetStatus(codes.Error, "get subscribe channel failed")
		span.RecordError(err)

		return nil, cs.Cleanup, fmt.Errorf("get subscribe channel: %w", err)
	}
	cs.Push(mqChannel.Close)

	span.AddEvent("prepare AMQP queue")
	exchangeName, err := declareEventsTopic(mqChannel)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange failed")
		span.RecordError(err)

		return nil, cs.Cleanup, fmt.Errorf("declare exchange: %w", err)
	}

	queue, prefetchCount, err := declareDurableQueueToTopic(mqChannel, exchangeName, mo.Some(models.EventTypeMovement))
	if err != nil {
		span.SetStatus(codes.Error, "declare queue failed")
		span.RecordError(err)

		return nil, cs.Cleanup, fmt.Errorf("declare queue: %w", err)
	}

	span.AddEvent("prepare raw message channel from AMQP")
	rawMessageCh, err := mqChannel.Consume(
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

		return nil, cs.Cleanup, fmt.Errorf("consume: %w", err)
	}
	cs.Push(func() error {
		return mqChannel.Cancel(consumer, false)
	})

	eventsCh := make(chan TraceableMovementEventDelivery, prefetchCount)
	span.AddEvent("create a goroutine to map raw messages to wrapped delivery")

	go func() {
		for rawMessage := range rawMessageCh {
			// If the Acknowledger is nil, it means this message is not valid;
			// if this message is over the requeue limit, we should reject it.
			if rawMessage.Acknowledger == nil || isDeliveryOverRequeueLimit(rawMessage) {
				_ = rawMessage.Reject(false)
				continue
			}

			eventsCh <- &traceableMovementEventMessageDelivery{
				TraceableEventDelivery: &amqpTraceableEventMessageDelivery{
					Delivery: rawMessage,
				},
			}
		}

		close(eventsCh)
	}()

	span.SetStatus(codes.Ok, "created a subscriber channel of "+queue.Name)
	return eventsCh, cs.Cleanup, nil
}

// traceableMovementEventMessageDelivery is a wrapper around TraceableEventDelivery
// that makes sure the event message is a movement event message.
type traceableMovementEventMessageDelivery struct {
	TraceableEventDelivery
}

// Body returns the parsed movement event message body.
func (t *traceableMovementEventMessageDelivery) Body() (MovementEventMessage, error) {
	body, err := t.TraceableEventDelivery.Body()
	if err != nil {
		return MovementEventMessage{}, err
	}

	if _, ok := body.GetEvent().(*eventpb.EventMessage_MovementInfo); !ok {
		return MovementEventMessage{}, fmt.Errorf("failed to cast the event message to MovementInfo")
	}

	return MovementEventMessage{MovementInfo: body.GetMovementInfo()}, nil
}

// Reject rejects the message.
func (t *traceableMovementEventMessageDelivery) Reject(requeue bool) error {
	return Reject(t.TraceableEventDelivery, requeue)
}

func (t *traceableMovementEventMessageDelivery) Ack() error {
	return Ack(t.TraceableEventDelivery)
}

var _ Rejectable = (*traceableMovementEventMessageDelivery)(nil)
var _ Ackable = (*traceableMovementEventMessageDelivery)(nil)
