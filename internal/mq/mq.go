// Package mq handles the message (such as Events) from the message queue, and provides
// the high-level interface to operate with the message.
package mq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// FxModule is the fx module for the MessageQueue, which handles the Close of MessageQueue correctly.
var FxModule = fx.Module("amqp-mq", fx.Provide(func(lifecycle fx.Lifecycle, config utils.Config) (MessageQueue, error) {
	messageQueue, err := ConnectAmqp(config)
	if err != nil {
		return nil, fmt.Errorf("connect to amqp: %w", err)
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return messageQueue.Close()
		},
	})

	return messageQueue, nil
}))

// MessageQueue is the interface for the message queue.
//
// It provides the high-level interface (such as EventSubscriber) to operate with the message queue.
type MessageQueue interface {
	EventSubscriber
	MovementEventSubscriber

	// Close closes the message queue.
	//
	// If you get this instance from FxModule, you should never call it since FxModule has handled it.
	Close() error
}

// amqpMQ is the AMQP-based implementation of the MessageQueue.
type amqpMQ struct {
	channel *amqp091.Channel

	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	logger     *slog.Logger
}

// ConnectAmqp connects to the AMQP message queue.
//
// The config "mq.address" specifies the address of the AMQP server.
func ConnectAmqp(config utils.Config) (MessageQueue, error) {
	const name = "amqpMQ"

	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	amqpAddress := config.String("mq.address")
	if amqpAddress == "" {
		return nil, fmt.Errorf("rabbitmq address is not set")
	}

	conn, err := amqp091.Dial(amqpAddress)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	return &amqpMQ{
		channel:    ch,
		propagator: propagator,
		tracer:     tracer,
		logger:     logger,
	}, nil
}

// Close closes the AMQP message queue.
func (mq *amqpMQ) Close() error {
	if err := mq.channel.Close(); err != nil {
		return fmt.Errorf("close channel: %w", err)
	}

	return nil
}
