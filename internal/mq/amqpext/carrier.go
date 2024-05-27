// Package amqpext provides the extension (like instrumentation) of the AMQP library.
package amqpext

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/propagation"
)

type HeaderSupplier struct {
	header amqp091.Table
}

func NewHeaderSupplier(header amqp091.Table) *HeaderSupplier {
	return &HeaderSupplier{
		header: header,
	}
}

func (h *HeaderSupplier) Get(key string) string {
	switch v := h.header[key].(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (h *HeaderSupplier) Set(key string, value string) {
	h.header[key] = value
}

func (h *HeaderSupplier) Keys() []string {
	return lo.MapToSlice(h.header, func(key string, _ any) string {
		return key
	})
}

var _ propagation.TextMapCarrier = &HeaderSupplier{}
