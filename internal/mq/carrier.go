package mq

import (
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

// MessageCarrier carries the request information across the boundary of services.
type MessageCarrier struct {
	Header amqp091.Table
}

func (m *MessageCarrier) Get(key string) string {
	if v, ok := m.Header[key].(string); ok {
		return v
	}

	if key == "request_id" {
		logger.Warn("request_id is not provided!")
	}

	return ""
}

func (m *MessageCarrier) Set(key string, value string) {
	m.Header[key] = value
}

func (m *MessageCarrier) Keys() []string {
	return []string{
		"request_id",
	}
}

func NewMessageCarrier(header amqp091.Table) *MessageCarrier {
	return &MessageCarrier{
		Header: header,
	}
}

var _ propagation.TextMapCarrier = &MessageCarrier{}
