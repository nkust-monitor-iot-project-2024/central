package mq

import (
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

var FxModule = fx.Module("amqp-mq", fx.Provide(ConnectAmqp))

type MessageQueue interface {
	EventSubscriber
}

type amqpMQ struct {
	channel *amqp091.Channel

	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	logger     *slog.Logger
}

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
