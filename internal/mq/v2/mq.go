// Package mqv2 is the rewritten version of the classical Message Queue packages
// that provides a thin layer of abstraction over the RabbitMQ message queue.
package mqv2

import (
	"context"
	"errors"
	"fmt"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.uber.org/fx"
	"log/slog"
	"time"
)

// FxModule is the fx module for the MessageQueue, which handles the Close of MessageQueue correctly.
var FxModule = fx.Module("amqp-mq-v2", fx.Provide(func(lifecycle fx.Lifecycle, config utils.Config) (*AmqpWrapper, error) {
	messageQueue, err := NewAmqpWrapper(config)
	if err != nil {
		return nil, fmt.Errorf("connect to amqp: %w", err)
	}

	return messageQueue, nil
}))

var ErrRabbitMQURINotSet = errors.New("rabbitmq URI is not set")

// AmqpWrapper is the wrapper for the AMQP message queue.
type AmqpWrapper struct {
	addr string
}

// NewAmqpWrapper creates a new AMQP wrapper.
func NewAmqpWrapper(config utils.Config) (*AmqpWrapper, error) {
	mqAddress := config.String("mq.uri")
	if mqAddress == "" {
		return nil, ErrRabbitMQURINotSet
	}

	return &AmqpWrapper{
		addr: mqAddress,
	}, nil
}

// NewConnection creates a new connection to the AMQP server.
//
// It does not handle the reconnection, and it will return the connection directly.
// If the connection is closed, you should create a new connection.
func (a *AmqpWrapper) NewConnection() (*amqp091.Connection, error) {
	return amqp091.Dial(a.addr)
}

// NewConnectionSupplier creates a new connection supplier to the AMQP server.
//
// The connection supplier will try to connect to the AMQP server and return the connection.
// If the connection is closed, the connection supplier will try to reconnect to the AMQP server until the maximum
// retry count is reached or the context is done.
//
// The connection supplier will return the connection if the connection is successfully connected.
// Otherwise, it will return an error and just close the connection (if the context has closed, it closes the channel directly).
//
// Note that you should close the current connection you get before getting a new connection;
// otherwise, no connection will be provided, and the consumer task will be blocked.
func (a *AmqpWrapper) NewConnectionSupplier(ctx context.Context) chan mo.Result[*amqp091.Connection] {
	connection := make(chan mo.Result[*amqp091.Connection])
	retryCount := 0

	go func() {
		defer close(connection)

		for {
			// If the context is done, we should stop the allocator.
			if ctx.Err() != nil {
				return
			}

			// Connect to the AMQP server.
			conn, err := a.NewConnection()
			if err != nil {
				retryCount += 1

				if retryCount > 5 { // retry 5 times
					select {
					case <-ctx.Done():
						return
					case connection <- mo.Err[*amqp091.Connection](fmt.Errorf("retry 5 times and still failed to connect to the AMQP server: %w", err)):
					}

					return
				}
			}

			// Push the connection for the consumer.
			select {
			case <-ctx.Done():
				return
			case connection <- mo.Ok(conn):
			}

			// Start monitoring for this connection and block until the connection is closed.
			//
			// If the context is done, we should close the connection and stop the allocator.
			// If the connection is closed and the context is not done,
			// we should close the connection and retry.
			select {
			case <-ctx.Done():
				err := conn.CloseDeadline(time.Now().Add(1 * time.Second))
				slog.ErrorContext(ctx, "failed to close the connection", slogext.Error(err))
				return
			case err := <-conn.NotifyClose(make(chan *amqp091.Error)):
				slog.Warn("connection closed", slogext.Error(err))
				continue // create a new connection
			}
		}
	}()

	return connection
}
