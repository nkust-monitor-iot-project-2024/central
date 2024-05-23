package mq

import (
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

const name = "mq"

var (
	propagator = otel.GetTextMapPropagator()
	tracer     = otel.Tracer(name)
	logger     = utils.NewLogger(name)
)

type MessageQueue interface {
	EventSubscriber
}

type amqpMQ struct {
	channel *amqp091.Channel
}

func ConnectAmqp(config utils.Config) (MessageQueue, error) {
	amqpAddress := config.String("mq.address")
	if amqpAddress == "" {
		panic("rabbitmq address is not set")
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
		channel: ch,
	}, nil
}
