package mq

import (
	"context"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/encoding/protojson"
)

type EventSubscriber interface {
	SubscribeEvent(ctx context.Context) (<-chan *eventpb.EventMessage, error)
}

type TypedDelivery[T any] struct {
	amqp091.Delivery

	Body T
}

func (mq *amqpMQ) SubscribeEvent(parentCtx context.Context) (<-chan *eventpb.EventMessage, error) {
	ctx, span := tracer.Start(parentCtx, "mq.SubscribeEvent")
	defer span.End()

	span.AddEvent("declare exchange of AMQP")
	err := mq.channel.ExchangeDeclare(
		"events_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "declare exchange failed")
		span.RecordError(err)

		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	span.AddEvent("declared exchange of AMQP")

	span.AddEvent("declare queue of AMQP")
	queue, err := mq.channel.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "declare queue failed")
		span.RecordError(err)

		return nil, fmt.Errorf("declare queue: %w", err)
	}
	span.AddEvent("declared queue of AMQP")

	span.AddEvent("bind queue to exchange of AMQP")
	err = mq.channel.QueueBind(
		queue.Name,
		"event.v1.*",
		"events_topic",
		false,
		nil)
	if err != nil {
		span.SetStatus(codes.Error, "bind queue failed")
		span.RecordError(err)

		return nil, fmt.Errorf("bind queue: %w", err)
	}
	span.AddEvent("bound queue to exchange of AMQP")

	span.AddEvent("prepare raw message channel from AMQP")
	rawMessageCh, err := mq.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		span.SetStatus(codes.Error, "consume failed")
		span.RecordError(err)

		return nil, fmt.Errorf("consume: %w", err)
	}
	span.AddEvent("prepared the raw channel")

	eventsCh := make(chan *eventpb.EventMessage)
	go func() {
		defer close(eventsCh)

		// find the request id; if none, we set one.

		for rawMessage := range rawMessageCh {
			var event *eventpb.EventMessage

			if err := protojson.Unmarshal(rawMessage.Body, event); err != nil {
				logger.WarnContext(parentCtx, "unmarshal failed – ignoring", slogext.Error(err))
				continue
			}

			select {
			case <-ctx.Done():
				// close the channel if the context is done
				err := mq.channel.Cancel("", false)
				if err != nil {
					logger.ErrorContext(parentCtx, "cancel consume failed", slogext.Error(err))
				}
				return
			default:
				eventsCh <- event
			}
		}
	}()

	return eventsCh, nil
}
