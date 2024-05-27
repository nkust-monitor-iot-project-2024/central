// Package mq handles the message (such as Events) from the message queue, and provides
// the high-level interface to operate with the message.
package mq

import (
	"context"
	"errors"
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

	EventPublisher

	// Close closes the message queue.
	//
	// If you get this instance from FxModule, you should never call it since FxModule has handled it.
	Close() error
}

// amqpMQ is the AMQP-based implementation of the MessageQueue.
type amqpMQ struct {
	mqAddress string

	// pub/sub separated channels as the AMQP library advised.
	// you should always call getPubChan to get the channel singleton
	// (it handles the connection opening and closing).

	pubChan *amqp091.Channel
	subChan *amqp091.Channel

	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	logger     *slog.Logger
}

var errMQAddressNotSet = errors.New("rabbitmq address is not set")

// ConnectAmqp connects to the AMQP message queue.
//
// The config "mq.address" specifies the address of the AMQP server.
func ConnectAmqp(config utils.Config) (MessageQueue, error) {
	const name = "amqpMQ"

	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	mqAddress := config.String("mq.address")
	if mqAddress == "" {
		return nil, errMQAddressNotSet
	}

	return &amqpMQ{
		mqAddress:  config.String("mq.address"),
		propagator: propagator,
		tracer:     tracer,
		logger:     logger,
	}, nil
}

func (mq *amqpMQ) openMQConnection() (*amqp091.Channel, error) {
	if mq.mqAddress == "" {
		return nil, errMQAddressNotSet
	}

	conn, err := amqp091.Dial(mq.mqAddress)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	return ch, nil
}

func (mq *amqpMQ) getPubChan() (*amqp091.Channel, error) {
	if mq.pubChan == nil {
		ch, err := mq.openMQConnection()
		if err != nil {
			return nil, fmt.Errorf("open connection: %w", err)
		}

		mq.pubChan = ch
	}

	return mq.pubChan, nil
}

func (mq *amqpMQ) getSubChan() (*amqp091.Channel, error) {
	if mq.subChan == nil {
		ch, err := mq.openMQConnection()
		if err != nil {
			return nil, fmt.Errorf("open connection: %w", err)
		}

		mq.subChan = ch
	}

	return mq.subChan, nil
}

// Close closes the AMQP message queue.
func (mq *amqpMQ) Close() (err error) {
	if mq.pubChan != nil {
		if err := mq.pubChan.Close(); err != nil {
			err = errors.Join(err, fmt.Errorf("close publish channel: %w", err))
		}
	}

	if mq.subChan != nil {
		if err := mq.subChan.Close(); err != nil {
			err = errors.Join(err, fmt.Errorf("close subscribe channel: %w", err))
		}
	}

	return
}
