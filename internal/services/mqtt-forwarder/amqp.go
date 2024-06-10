package mqtt_forwarder

import (
	"context"
	"errors"
	"fmt"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// AmqpPublisher sends the message to the AMQP broker.
type AmqpPublisher struct {
	connectionSupplier chan mo.Result[*amqp091.Connection]

	connection mo.Option[*amqp091.Connection]

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewAmqpPublisher creates a new AMQP publisher.
func NewAmqpPublisher(ctx context.Context, amqp mqv2.AmqpWrapper) *AmqpPublisher {
	supplier := amqp.NewConnectionSupplier(ctx)

	return &AmqpPublisher{
		connectionSupplier: supplier,
		connection:         mo.None[*amqp091.Connection](),
		tracer:             otel.GetTracerProvider().Tracer("amqp-publisher"),
		propagator:         otel.GetTextMapPropagator(),
	}
}

// getConnection checks if the connection is already established and returns it.
// If the connection is not established, it requests for a new connection.
func (p *AmqpPublisher) getConnection() (*amqp091.Connection, error) {
	existingConnection, ok := p.connection.Get()
	if ok && !existingConnection.IsClosed() {
		return existingConnection, nil
	}

	connectionResult := <-p.connectionSupplier
	newConnection, err := connectionResult.Get()
	if err != nil {
		return nil, fmt.Errorf("connection: %w", err)
	}

	p.connection = mo.Some[*amqp091.Connection](newConnection)
	return newConnection, nil
}

// Publish sends the TraceableEventMessage to the AMQP broker.
func (p *AmqpPublisher) Publish(ctx context.Context, message TraceableEventMessage) error {
	ctx = trace.ContextWithSpanContext(ctx, message.SpanContext)
	ctx, span := p.tracer.Start(ctx, "mqtt-forwarder/amqp-publisher/publish")
	defer span.End()

	span.AddEvent("prepare AMQP connection and exchange")

	connection, err := p.getConnection()
	if err != nil {
		span.SetStatus(codes.Error, "get connection")
		span.RecordError(err)

		return fmt.Errorf("get connection: %w", err)
	}

	channel, err := connection.Channel()
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
		if errors.Is(err, context.Canceled) {
			return err
		}

		span.SetStatus(codes.Error, "publish event")
		span.RecordError(err)

		return fmt.Errorf("publish event: %w", err)
	}
	span.AddEvent("published event")

	return nil
}
