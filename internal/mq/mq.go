// Package mq handles the message (such as Events) from the message queue, and provides
// the high-level interface to operate with the message.
package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
			err := messageQueue.Close()
			return err
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

	// pub/sub separated connections as the AMQP library advised.
	// you should always call getSingletonXXXConnection to get the channel singleton
	// (it handles the connection opening and closing).

	pubConn *amqp091.Connection
	subConn *amqp091.Connection

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

func (mq *amqpMQ) openMQConnection() (*amqp091.Connection, error) {
	if mq.mqAddress == "" {
		return nil, errMQAddressNotSet
	}

	conn, err := amqp091.Dial(mq.mqAddress)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	return conn, nil
}

func (mq *amqpMQ) getSingletonPublishConnection() (*amqp091.Connection, error) {
	if mq.pubConn == nil || mq.pubConn.IsClosed() {
		conn, err := mq.openMQConnection()
		if err != nil {
			return nil, fmt.Errorf("open publish connection: %w", err)
		}

		mq.pubConn = conn
	}

	return mq.pubConn, nil
}

func (mq *amqpMQ) getSingletonSubscribeConnection() (*amqp091.Connection, error) {
	if mq.subConn == nil || mq.subConn.IsClosed() {
		conn, err := mq.openMQConnection()
		if err != nil {
			return nil, fmt.Errorf("open subscribe connection: %w", err)
		}

		mq.subConn = conn
	}

	return mq.subConn, nil
}

// Close closes the AMQP message queue.
//
// It closes the publishing and subscribing connections for each in 1 second.
func (mq *amqpMQ) Close() (closeErr error) {
	if mq.pubConn != nil {
		if err := mq.pubConn.CloseDeadline(time.Now().Add(1 * time.Second)); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("closech publish connection: %w", err))
		}
	}

	if mq.subConn != nil {
		if err := mq.subConn.CloseDeadline(time.Now().Add(1 * time.Second)); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("closech subscribe connection: %w", err))
		}
	}

	return
}
