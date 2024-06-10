package mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// EventSubscriber is the interface for services to subscribe to event messages.
type EventSubscriber interface {
	// SubscribeEvent subscribes to the event messages.
	SubscribeEvent(ctx context.Context) (SubscribeResponse[TraceableEventDelivery], error)
}

// TraceableEventDeliveryMetadata is the interface for the event message delivery that includes
// the metadata extraction and the span extractor.
type TraceableEventDeliveryMetadata interface {
	// Extract extracts the span for tracing from the event message.
	Extract(ctx context.Context, propagator propagation.TextMapPropagator) context.Context

	// Metadata returns the parsed event message metadata.
	Metadata() (models.Metadata, error)
}

// TraceableEventDelivery is the interface for the event message delivery that includes
// the metadata extraction, the span extractor, and the body extraction.
type TraceableEventDelivery interface {
	TraceableEventDeliveryMetadata

	// Body returns the parsed event message body.
	Body() (*eventpb.EventMessage, error)
}

// SubscribeEvent subscribes to the event messages.
//
// Deprecate: Use mqv2 instead.
func (mq *amqpMQ) SubscribeEvent(ctx context.Context) (SubscribeResponse[TraceableEventDelivery], error) {
	const consumer = "subscribe-event"

	_, span := mq.tracer.Start(ctx, "mq/subscribe_event", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	cs := utils.NewCleanupStack()
	closedChan := make(chan error, 2)
	response := SubscribeResponse[TraceableEventDelivery]{
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
		err := <-mqConnection.NotifyClose(make(chan *amqp091.Error, 1))
		closedChan <- err
	}()
	cs.Push(mqChannel.Close)

	key := mqevent.GetRoutingKey(mo.None[models.EventType]())

	queue, err = channel.QueueDeclare(
		"",
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-queue-type": "quorum",
		},
	)

	span.AddEvent("prepare AMQP queue")
	exchangeName, err := event.declareEventsTopic(mqChannel)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange failed")
		span.RecordError(err)

		return response, fmt.Errorf("declare exchange: %w", err)
	}

	queue, prefetchCount, err := event.declareDurableQueueToTopic(mqChannel, exchangeName, mo.None[models.EventType]())
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

	eventsCh := make(chan TraceableEventDelivery, prefetchCount)
	span.AddEvent("create a goroutine to map raw messages to wrapped delivery")

	go func() {
		for rawMessage := range rawMessageCh {
			// If the Acknowledger is nil, it means this message is not valid;
			// if this message is over the requeue limit, we should reject it.
			if rawMessage.Acknowledger == nil || event.isDeliveryOverRequeueLimit(rawMessage) {
				_ = rawMessage.Reject(false)
				continue
			}

			eventsCh <- &amqpTraceableEventMessageDelivery{Delivery: rawMessage}
		}

		close(eventsCh)
	}()

	response.DeliveryChan = eventsCh
	span.SetStatus(codes.Ok, "created a subscriber channel of "+queue.Name)

	return response, nil
}

// amqpTraceableEventMessageDelivery is a wrapper around amqp091.Delivery that includes a function
// to extract the metadata and the body from Delivery, and provide the span extractor of
// the Delivery.
type amqpTraceableEventMessageDelivery struct {
	amqp091.Delivery
}

func (w *amqpTraceableEventMessageDelivery) Extract(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	supplier := amqpext.NewHeaderSupplier(w.Headers)

	return propagator.Extract(ctx, supplier)
}

func (w *amqpTraceableEventMessageDelivery) Metadata() (models.Metadata, error) {
	eventID := w.Delivery.MessageId
	if eventID == "" {
		return models.Metadata{}, errors.New("missing or invalid MessageId (-> event_id)")
	}
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("parse event_id: %w", err)
	}

	deviceID := w.Delivery.AppId
	if deviceID == "" {
		return models.Metadata{}, errors.New("missing or invalid AppId (-> device_id)")
	}

	eventTs := w.Delivery.Timestamp

	metadata := models.Metadata{
		EventID:       eventUUID,
		DeviceID:      deviceID,
		EmittedAt:     eventTs,
		ParentEventID: mo.None[uuid.UUID](),
	}

	if parentEventID, ok := w.Headers["parent_event_id"]; ok {
		parentEventUUID, err := uuid.Parse(parentEventID.(string))
		if err != nil {
			return models.Metadata{}, fmt.Errorf("parse parent_event_id: %w", err)
		}

		metadata.ParentEventID = mo.Some[uuid.UUID](parentEventUUID)
	}

	return metadata, nil
}

func (w *amqpTraceableEventMessageDelivery) Body() (*eventpb.EventMessage, error) {
	if w.Delivery.ContentType != "application/x-google-protobuf" || w.Delivery.Type != "eventpb.EventMessage" {
		return nil, errors.New("invalid header")
	}

	event := &eventpb.EventMessage{}

	if err := proto.Unmarshal(w.Delivery.Body, event); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return event, nil
}

func (w *amqpTraceableEventMessageDelivery) Reject(requeue bool) error {
	return w.Delivery.Reject(requeue)
}

func (w *amqpTraceableEventMessageDelivery) Ack() error {
	return w.Delivery.Ack(false)
}

var _ Rejectable = (*amqpTraceableEventMessageDelivery)(nil)
var _ Ackable = (*amqpTraceableEventMessageDelivery)(nil)
