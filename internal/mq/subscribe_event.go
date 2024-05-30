package mq

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// EventSubscriber is the interface for services to subscribe to event messages.
type EventSubscriber interface {
	// SubscribeEvent subscribes to the event messages.
	SubscribeEvent(ctx context.Context) (<-chan TraceableEventDelivery, error)
}

type TraceableEventDelivery interface {
	Extract(ctx context.Context, propagator propagation.TextMapPropagator) context.Context
	Metadata() (models.Metadata, error)
	Body() (*eventpb.EventMessage, error)
}

// SubscribeEvent subscribes to the event messages.
func (mq *amqpMQ) SubscribeEvent(ctx context.Context) (deliveryChan <-chan TraceableEventDelivery, cleanup func() error, err error) {
	const consumer = "subscribe-event"

	_, span := mq.tracer.Start(ctx, "mq/subscribe_event", trace.WithSpanKind(trace.SpanKindInternal))
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

	queue, prefetchCount, err := declareDurableQueueToTopic(mqChannel, exchangeName, mo.None[models.EventType]())
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

	eventsCh := make(chan TraceableEventDelivery, prefetchCount)
	span.AddEvent("create a goroutine to map raw messages to amqpTraceableEventMessageDelivery")

	go func() {
		for rawMessage := range rawMessageCh {
			// If the Acknowledger is nil, it means this message is not valid.
			if rawMessage.Acknowledger == nil {
				_ = rawMessage.Reject(false)
				continue
			}

			eventsCh <- &amqpTraceableEventMessageDelivery{Delivery: rawMessage}
		}

		close(eventsCh)
	}()

	span.SetStatus(codes.Ok, "created a subscriber channel of "+queue.Name)
	return eventsCh, cs.Cleanup, nil
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

	return models.Metadata{
		EventID:   eventUUID,
		DeviceID:  deviceID,
		EmittedAt: eventTs,
	}, nil
}

func (w *amqpTraceableEventMessageDelivery) Body() (*eventpb.EventMessage, error) {
	if w.Delivery.ContentType != "application/json" || w.Delivery.Type != "eventpb.EventMessage" {
		return nil, errors.New("invalid header")
	}

	event := &eventpb.EventMessage{}

	if err := protojson.Unmarshal(w.Delivery.Body, event); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return event, nil
}
