package mq

import (
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

// MessageCarrier carries the request information across the boundary of services.
type MessageCarrier struct {
	Delivery amqp091.Delivery
}

func (m *MessageCarrier) Get(key string) string {
	switch key {
	case "event_id":
		return m.Delivery.MessageId
	case "device_id":
		return m.Delivery.AppId
	default:
		if m.Delivery.Headers != nil {
			if value, ok := m.Delivery.Headers[key]; ok {
				return value.(string)
			}
		}
	}

	return ""
}

func (m *MessageCarrier) Set(key string, value string) {
	switch key {
	case "event_id":
		m.Delivery.MessageId = value
	case "device_id":
		m.Delivery.AppId = value
	default:
		if m.Delivery.Headers == nil {
			m.Delivery.Headers = make(amqp091.Table)
		}

		m.Delivery.Headers[key] = value
	}
}

func (m *MessageCarrier) Keys() []string {
	return []string{
		"event_id",
		"device_id",
	}
}

func NewMessageHeaderCarrier(delivery amqp091.Delivery) *MessageCarrier {
	return &MessageCarrier{
		Delivery: delivery,
	}
}

var _ propagation.TextMapCarrier = &MessageCarrier{}
