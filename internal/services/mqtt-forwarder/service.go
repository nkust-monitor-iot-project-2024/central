// Package mqtt_forwarder provides the core of the service "mqtt-forwarder",
// which forwards the messages from the MQTT broker to our event exchanges.
//
// MQTT is suitable for IoT devices, while AMQP is suitable for backend services.
// As the events are usually sent by Raspberry Pi devices, we should have a bridge
// to forward the messages from MQTT to AMQP in the correct format.
package mqtt_forwarder

import (
	"context"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

// FxModule is the fx module for the Service that handles the cleanup.
var FxModule = fx.Module(
	"mqtt-forwarder",
	mq.FxModule,
	fx.Provide(fx.Annotate(newFx, fx.As(new(services.Service)))),
	fx.Invoke(services.BootstrapFxService),
)

// Service is the MQTT forwarder service.
type Service struct {
	publisher mq.EventPublisher
	config    utils.Config
}

// newFx provides the constructor for the MQTT forwarder service.
func newFx(publisher mq.MessageQueue, config utils.Config) *Service {
	return New(publisher, config)
}

// New creates a new MQTT forwarder service.
func New(publisher mq.EventPublisher, config utils.Config) *Service {
	return &Service{
		publisher: publisher,
		config:    config,
	}
}

// Run starts the MQTT forwarder service.
func (s *Service) Run(ctx context.Context) error {
	mqttForwarder, err := ConnectMqttReceiver(s.config)
	if err != nil {
		return fmt.Errorf("new mqtt receiver: %w", err)
	}

	amqpPublisher := NewAmqpPublisher(s.publisher)

	convertedMqttMessage := make(chan TraceableEventMessage, 64)
	adapter := NewAdapter(mqttForwarder.GetMqttMessageChannel(), convertedMqttMessage)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		adapter.Run(ctx)
		return nil
	})

	group.Go(func() error {
		amqpPublisher.PublishFromChannel(ctx, convertedMqttMessage)
		return nil
	})

	return group.Wait()
}
