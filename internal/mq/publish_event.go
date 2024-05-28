package mq

import (
	"context"
	"errors"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// EventPublisher is the interface for services to publish an event messages.
type EventPublisher interface {
	// PublishEvent publishes the event to MQ.
	PublishEvent(ctx context.Context, metadata models.Metadata, event *eventpb.EventMessage) error
}

// PublishEvent publishes the event to AMQP instance.
//
// amqpMQ will open a publish-only channel, as the amqp library advised.
func (mq *amqpMQ) PublishEvent(ctx context.Context, metadata models.Metadata, event *eventpb.EventMessage) error {
	_, span := mq.tracer.Start(ctx, "mq/publish_event", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.AddEvent("prepare AMQP [publish] channel")
	mqChannel, err := mq.getPubChan()
	if err != nil {
		span.SetStatus(codes.Error, "get publish channel failed")
		span.RecordError(err)

		return fmt.Errorf("get publish channel: %w", err)
	}
	span.AddEvent("prepared AMQP [publish] channel")

	span.AddEvent("prepare AMQP exchange")
	err = mqChannel.ExchangeDeclare(
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

		return fmt.Errorf("declare exchange: %w", err)
	}
	span.AddEvent("prepared AMQP exchange")

	span.AddEvent("publish message to AMQP")
	key, publishing, err := mq.createPublishingMessage(ctx, metadata, event)
	if err != nil {
		span.SetStatus(codes.Error, "create publishing message failed")
		span.RecordError(err)

		return fmt.Errorf("create publishing message: %w", err)
	}

	err = mqChannel.PublishWithContext(
		ctx,
		"events_topic",
		key,
		false,
		false,
		publishing,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			span.SetStatus(codes.Unset, "context canceled")
			return err
		}

		span.SetStatus(codes.Error, "publish message failed")
		span.RecordError(err)

		return fmt.Errorf("publish message: %w", err)
	}
	span.AddEvent("published message to AMQP")

	span.SetStatus(codes.Ok, "published message successfully")
	return nil
}

func (mq *amqpMQ) createPublishingMessage(ctx context.Context, metadata models.Metadata, event *eventpb.EventMessage) (key string, publishing amqp091.Publishing, err error) {
	var eventType models.EventType
	switch event.GetEvent().(type) {
	case *eventpb.EventMessage_MovementInfo:
		eventType = models.EventTypeMovement
	case *eventpb.EventMessage_InvadedInfo:
		eventType = models.EventTypeInvaded
	default:
		return "", amqp091.Publishing{}, fmt.Errorf("unknown event type: %T", event.GetEvent())
	}

	// Inject the context into the header, so we can trace the span across the system boundaries.
	header := amqp091.Table{}
	mq.propagator.Inject(ctx, amqpext.NewHeaderSupplier(header))

	marshalledBody, err := protojson.Marshal(event)
	if err != nil {
		return "", amqp091.Publishing{}, fmt.Errorf("marshal event: %w", err)
	}

	return "event.v1." + string(eventType), amqp091.Publishing{
		ContentType: "application/json",
		MessageId:   metadata.GetEventID().String(),
		Timestamp:   metadata.GetEmittedAt(),
		Type:        "eventpb.EventMessage",
		AppId:       metadata.GetDeviceID(),
		Headers:     header,
		Body:        marshalledBody,
	}, nil
}
