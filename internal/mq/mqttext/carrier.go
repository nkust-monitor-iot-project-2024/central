// Package mqttext provides the extension (like instrumentation) of the MQTT library.
package mqttext

import (
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/propagation"
)

type HeaderSupplier struct {
	header paho.UserProperties
}

func NewHeaderSupplier(header paho.UserProperties) *HeaderSupplier {
	return &HeaderSupplier{
		header: header,
	}
}

func (h *HeaderSupplier) Get(key string) string {
	return h.header.Get(key)
}

func (h *HeaderSupplier) Set(key string, value string) {
	// unsupported; noop
}

func (h *HeaderSupplier) Keys() []string {
	return lo.Map(h.header.ToPacketProperties(), func(k packets.User, _ int) string {
		return k.Key
	})
}

var _ propagation.TextMapCarrier = &HeaderSupplier{}
