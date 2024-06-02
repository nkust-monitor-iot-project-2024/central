package mqtt_forwarder

import (
	"context"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AmqpPublisher sends the message to the AMQP broker.
type AmqpPublisher struct {
	publisher mq.EventPublisher

	tracer trace.Tracer
}

// NewAmqpPublisher creates a new AMQP publisher.
func NewAmqpPublisher(publisher mq.EventPublisher) *AmqpPublisher {
	return &AmqpPublisher{
		publisher: publisher,
		tracer:    otel.GetTracerProvider().Tracer("amqp-publisher"),
	}
}

// Publish sends the TraceableEventMessage to the AMQP broker.
func (p *AmqpPublisher) Publish(ctx context.Context, message TraceableEventMessage) error {
	ctx = trace.ContextWithSpanContext(ctx, message.SpanContext)
	ctx, span := p.tracer.Start(ctx, "mqtt-forwarder/amqp-publisher/publish")
	defer span.End()

	err := p.publisher.PublishEvent(ctx, message.Metadata, message.EventMessage)
	if err != nil {
		span.SetStatus(codes.Error, "publish event")
		span.RecordError(err)

		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

// PublishFromChannel sends the TraceableEventMessage from the channel to the AMQP broker.
func (p *AmqpPublisher) PublishFromChannel(ctx context.Context, messageChannel <-chan TraceableEventMessage) {
	for message := range messageChannel {
		if ctx.Err() != nil {
			return
		}

		if err := p.Publish(ctx, message); err != nil {
			continue
		}
	}
}
