package mqtt_forwarder

import (
	"context"
	"fmt"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// AmqpPublisher sends the message to the AMQP broker.
type AmqpPublisher struct {
	connection *amqp091.Connection

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewAmqpPublisher creates a new AMQP publisher.
func NewAmqpPublisher(connection *amqp091.Connection) *AmqpPublisher {
	return &AmqpPublisher{
		connection: connection,
		tracer:     otel.GetTracerProvider().Tracer("amqp-publisher"),
		propagator: otel.GetTextMapPropagator(),
	}
}

// Publish sends the TraceableEventMessage to the AMQP broker.
func (p *AmqpPublisher) Publish(ctx context.Context, message TraceableEventMessage) error {
	ctx = trace.ContextWithSpanContext(ctx, message.SpanContext)
	ctx, span := p.tracer.Start(ctx, "mqtt-forwarder/amqp-publisher/publish")
	defer span.End()

	span.AddEvent("prepare AMQP connection and exchange")

	channel, err := p.connection.Channel()
	if err != nil {
		span.SetStatus(codes.Error, "get channel")
		span.RecordError(err)

		return fmt.Errorf("get channel: %w", err)
	}
	defer func(channel *amqp091.Channel) {
		span.AddEvent("closing channel")
		err := channel.Close()
		if err != nil {
			span.SetStatus(codes.Error, "failed to close channel")
			span.RecordError(err)
		}
	}(channel)

	exchangeName, err := mqevent.DeclareEventsTopic(channel)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange")
		span.RecordError(err)

		return fmt.Errorf("declare exchange: %w", err)
	}

	span.AddEvent("publishing event")
	eventType, publishing, err := mqevent.CreatePublishingEvent(ctx, mqevent.EventPublishingPayload{
		Propagator: p.propagator,
		Metadata:   message.Metadata,
		Event:      message.EventMessage,
	})
	if err != nil {
		span.SetStatus(codes.Error, "failed to create publishing event")
		span.RecordError(err)

		return fmt.Errorf("failed to create publishing event: %w", err)
	}

	routingKey := mqevent.GetRoutingKey(mo.Some(eventType))

	err = channel.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false, /* mandatory */
		false, /* immediate */
		publishing,
	)
	if err != nil {
		span.SetStatus(codes.Error, "publish event")
		span.RecordError(err)

		return fmt.Errorf("publish event: %w", err)
	}

	span.AddEvent("published event")
	span.SetStatus(codes.Ok, "event published")

	return nil
}
