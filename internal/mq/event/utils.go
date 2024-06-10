// Package mqevent includes some utilities to handle the Event messages.
package mqevent

import (
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
)

// DeclareEventsTopic declares the events topic exchange.
func DeclareEventsTopic(channel *amqp091.Channel) (string, error) {
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

// GetRoutingKey returns the routing key for the event type.
//
// The routing key is in the format of "event.v1.<event_type>".
// If the event type is empty, it will return "event.v1.*".
func GetRoutingKey(eventType mo.Option[models.EventType]) string {
	if eventTypeValue, ok := eventType.Get(); ok {
		return "event.v1." + string(eventTypeValue)
	}

	return "event.v1.*"
}
