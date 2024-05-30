package mq

import (
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
)

// declareEventsTopic declares the events topic exchange.
func declareEventsTopic(channel *amqp091.Channel) (string, error) {
	const exchangeName = "events_topic"

	return exchangeName, channel.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

// declareDurableQueueToTopic declares a durable queue to the events topic exchange.
//
// It returns the queue name.
func declareDurableQueueToTopic(channel *amqp091.Channel, topicName string, eventType mo.Option[models.EventType]) (queue amqp091.Queue, prefetchCount int, err error) {
	const prefetchCountQos = 64
	const prefetchSizeQos = 0

	key := getMessageKey(eventType)

	queueName := func() string {
		if eventTypeValue, ok := eventType.Get(); ok {
			return "iot-monitoring-" + string(eventTypeValue) + "-events"
		}

		return "iot-monitoring-all-events"
	}()

	queue, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return queue, 0, err
	}

	err = channel.QueueBind(
		queue.Name,
		key,
		topicName,
		false,
		nil,
	)
	if err != nil {
		return queue, 0, err
	}

	if err := channel.Qos(prefetchCountQos, prefetchSizeQos, false); err != nil {
		return queue, 0, fmt.Errorf("set QoS: %w", err)
	}

	return queue, prefetchCountQos, nil
}

// getMessageKey returns the routing key for the event type.
//
// The routing key is in the format of "event.v1.<event_type>".
// If the event type is empty, it will return "event.v1.*".
func getMessageKey(eventType mo.Option[models.EventType]) string {
	if eventTypeValue, ok := eventType.Get(); ok {
		return "event.v1." + string(eventTypeValue)
	}

	return "event.v1.*"
}
