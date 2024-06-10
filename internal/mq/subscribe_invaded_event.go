package mq

import (
	"context"
	"fmt"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// InvadedEventSubscriber is the interface for services to subscribe to invaded event messages.
type InvadedEventSubscriber interface {
	// SubscribeInvadedEvent subscribes only the invaded event messages.
	SubscribeInvadedEvent(ctx context.Context) (SubscribeResponse[TraceableInvadedEventDelivery], error)
}

// InvadedEventMessage is the message type for invaded events.
//
// It is basically the destructured version of eventpb.EventMessage.
type InvadedEventMessage struct {
	ParentMovementID string
	Invaders         []*eventpb.Invader
}

// GetParentMovementID returns the parent movement ID of the invaded event message.
func (i *InvadedEventMessage) GetParentMovementID() string {
	return i.ParentMovementID
}

// GetInvaders returns the invaders info of the invaded event message.
func (i *InvadedEventMessage) GetInvaders() []*eventpb.Invader {
	return i.Invaders
}

// TraceableInvadedEventDelivery is the interface for the invaded event message delivery that includes
// the metadata extraction, the span extractor, and the body extraction.
type TraceableInvadedEventDelivery interface {
	TraceableEventDeliveryMetadata

	// Body returns the parsed invaded event message body.
	Body() (InvadedEventMessage, error)
}

// Deprecate: Use mqv2 instead.
func (mq *amqpMQ) SubscribeInvadedEvent(ctx context.Context) (SubscribeResponse[TraceableInvadedEventDelivery], error) {
	const consumer = "subscribe-invaded-event"

	_, span := mq.tracer.Start(ctx, "mq/subscribe_invaded_event", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	cs := utils.NewCleanupStack()
	closedChan := make(chan error, 2)
	response := SubscribeResponse[TraceableInvadedEventDelivery]{
		DeliveryChan: nil,
		Cleanup:      cs.Cleanup,
		ClosedChan:   closedChan,
	}

	span.AddEvent("prepare AMQP [subscribe] channel")
	mqConnection, err := mq.getSingletonSubscribeConnection()
	if err != nil {
		span.SetStatus(codes.Error, "get subscribe channel failed")
		span.RecordError(err)

		return response, fmt.Errorf("get subscribe channel: %w", err)
	}
	go func() {
		err := <-mqConnection.NotifyClose(make(chan *amqp091.Error, 1))
		closedChan <- err
	}()

	mqChannel, err := mqConnection.Channel()
	if err != nil {
		span.SetStatus(codes.Error, "get subscribe channel failed")
		span.RecordError(err)

		return response, fmt.Errorf("get subscribe channel: %w", err)
	}
	go func() {
		err := <-mqChannel.NotifyClose(make(chan *amqp091.Error, 1))
		closedChan <- err
	}()
	cs.Push(mqChannel.Close)

	span.AddEvent("prepare AMQP queue")
	exchangeName, err := event.declareEventsTopic(mqChannel)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange failed")
		span.RecordError(err)

		return response, fmt.Errorf("declare exchange: %w", err)
	}

	queue, prefetchCount, err := event.declareDurableQueueToTopic(mqChannel, exchangeName, mo.Some(models.EventTypeInvaded))
	if err != nil {
		span.SetStatus(codes.Error, "declare queue failed")
		span.RecordError(err)

		return response, fmt.Errorf("declare queue: %w", err)
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

		return response, fmt.Errorf("consume: %w", err)
	}
	cs.Push(func() error {
		return mqChannel.Cancel(consumer, false)
	})

	eventsCh := make(chan TraceableInvadedEventDelivery, prefetchCount)
	span.AddEvent("create a goroutine to map raw messages to wrapped delivery")

	go func() {
		for rawMessage := range rawMessageCh {
			// If the Acknowledger is nil, it means this message is not valid;
			// if this message is over the requeue limit, we should reject it.
			if rawMessage.Acknowledger == nil || event.isDeliveryOverRequeueLimit(rawMessage) {
				_ = rawMessage.Reject(false)
				continue
			}

			eventsCh <- &traceableInvadedEventMessageDelivery{
				TraceableEventDelivery: &amqpTraceableEventMessageDelivery{
					Delivery: rawMessage,
				},
			}
		}

		close(eventsCh)
	}()

	response.DeliveryChan = eventsCh
	span.SetStatus(codes.Ok, "created a subscriber channel of "+queue.Name)

	return response, nil
}

// traceableInvadedEventMessageDelivery is a wrapper around TraceableEventDelivery
// that makes sure the event message is a invaded event message.
type traceableInvadedEventMessageDelivery struct {
	TraceableEventDelivery
}

// Body returns the parsed invaded event message body.
func (t *traceableInvadedEventMessageDelivery) Body() (InvadedEventMessage, error) {
	body, err := t.TraceableEventDelivery.Body()
	if err != nil {
		return InvadedEventMessage{}, err
	}

	if _, ok := body.GetEvent().(*eventpb.EventMessage_InvadedInfo); !ok {
		return InvadedEventMessage{}, fmt.Errorf("failed to cast the event message to InvadedInfo")
	}

	return InvadedEventMessage{
		ParentMovementID: body.GetInvadedInfo().GetParentMovementId(),
		Invaders:         body.GetInvadedInfo().GetInvaders(),
	}, nil
}

// Reject rejects the message.
func (t *traceableInvadedEventMessageDelivery) Reject(requeue bool) error {
	return Reject(t.TraceableEventDelivery, requeue)
}

func (t *traceableInvadedEventMessageDelivery) Ack() error {
	return Ack(t.TraceableEventDelivery)
}

var _ Rejectable = (*traceableInvadedEventMessageDelivery)(nil)
var _ Ackable = (*traceableInvadedEventMessageDelivery)(nil)
