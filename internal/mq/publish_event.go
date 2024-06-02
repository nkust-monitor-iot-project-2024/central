package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
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
	mqConnection, err := mq.getSingletonPublishConnection()
	if err != nil {
		span.SetStatus(codes.Error, "get publish connection failed")
		span.RecordError(err)

		return fmt.Errorf("get publish connection: %w", err)
	}

	mqChannel, err := mqConnection.Channel()
	if err != nil {
		span.SetStatus(codes.Error, "get publish connection failed")
		span.RecordError(err)

		return fmt.Errorf("get publish connection: %w", err)
	}
	defer func(mqChannel *amqp091.Channel) {
		err := mqChannel.Close()
		if err != nil {
			slog.Error("close channel failed", slogext.Error(err))
		}
	}(mqChannel)
	span.AddEvent("prepared AMQP [publish] channel")

	span.AddEvent("prepare AMQP exchange")
	exchangeName, err := declareEventsTopic(mqChannel)
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
		exchangeName,
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

	if parentEventID, ok := metadata.GetParentEventID(); ok {
		header["parent_event_id"] = parentEventID.String()
	}

	marshalledBody, err := proto.Marshal(event)
	if err != nil {
		return "", amqp091.Publishing{}, fmt.Errorf("marshal event: %w", err)
	}

	return getMessageKey(mo.Some(eventType)), amqp091.Publishing{
		ContentType: "application/x-google-protobuf",
		MessageId:   metadata.GetEventID().String(),
		Timestamp:   metadata.GetEmittedAt(),
		Type:        "eventpb.EventMessage",
		AppId:       metadata.GetDeviceID(),
		Headers:     header,
		Body:        marshalledBody,
	}, nil
}
