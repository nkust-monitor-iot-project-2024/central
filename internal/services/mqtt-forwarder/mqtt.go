package mqtt_forwarder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/mqttext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MqttReceiver is a receiver for MQTT messages.
type MqttReceiver struct {
	ctx    context.Context
	cancel context.CancelFunc

	connection *autopaho.ConnectionManager

	mqttMessageChannel chan TraceableMqttPublish

	tracer trace.Tracer
}

// TraceableMqttPublish is a wrapper for paho.Publish with a SpanContext.
type TraceableMqttPublish struct {
	*paho.Publish
	SpanContext trace.SpanContext
}

// ConnectMqttReceiver creates a new MQTT receiver and connects to the MQTT broker.
//
// It maintains a context.Context internally. To stop the MQTT receiver, call Close.
func ConnectMqttReceiver(config utils.Config) (*MqttReceiver, error) {
	const topic = "iot/events/v1/#"

	tracer := otel.GetTracerProvider().Tracer("mqtt-receiver")
	propagator := otel.GetTextMapPropagator()

	serverUrlRaw := config.String("mq.mqtt.uri")
	if serverUrlRaw == "" {
		return nil, errors.New("missing configuration: mq.mqtt.uri")
	}

	serverUrl, err := url.Parse(serverUrlRaw)
	if err != nil {
		return nil, fmt.Errorf("parse server URL %q: %w", serverUrl, err)
	}

	mqttMessageChannel := make(chan TraceableMqttPublish, 64)

	ctx, cancel := context.WithCancel(context.Background())
	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{serverUrl},
		KeepAlive:  20,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			if _, err := cm.Subscribe(ctx, &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: topic, QoS: 0},
				},
			}); err != nil {
				slog.Error("failed to subscribe", slogext.Error(err))
			}
			slog.Info("mqtt subscription made")
		},
		OnConnectError: func(err error) {
			slog.Error("error whilst attempting connection", slogext.Error(err))
		},

		ClientConfig: paho.ClientConfig{
			ClientID: "central/mqtt-receiver",
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					ctx = propagator.Extract(ctx, mqttext.NewHeaderSupplier(pr.Packet.Properties.User))
					ctx, span := tracer.Start(ctx, "mqtt-forwarder/mqtt/on_publish_received")
					defer span.End()

					if ctx.Err() != nil {
						return false, nil
					}

					span.AddEvent("received message", trace.WithAttributes(
						attribute.Int("packetId", int(pr.Packet.PacketID)),
						attribute.String("topic", pr.Packet.Topic),
						attribute.Int("qos", int(pr.Packet.QoS)),
					))
					span.SetAttributes(
						attribute.Int("packetId", int(pr.Packet.PacketID)),
						attribute.String("topic", pr.Packet.Topic),
					)

					mqttMessageChannel <- TraceableMqttPublish{
						Publish:     pr.Packet,
						SpanContext: span.SpanContext(),
					}

					return true, nil
				}},
			OnClientError: func(err error) { slog.Error("client error", slogext.Error(err)) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					slog.Error("server requested disconnect", slog.String("reason", d.Properties.ReasonString))
				} else {
					slog.Error("server requested disconnect",
						slog.String("reason", d.Properties.ReasonString),
						slog.Int("reason_code", int(d.ReasonCode)))
				}
			},
		},
	}
	c, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		cancel() // failed; cancel the context
		return nil, fmt.Errorf("create connection: %w", err)
	}
	if err := c.AwaitConnection(ctx); err != nil {
		cancel() // failed; cancel the context
		return nil, fmt.Errorf("await connection: %w", err)
	}

	return &MqttReceiver{
		ctx:                ctx,
		cancel:             cancel,
		connection:         c,
		mqttMessageChannel: mqttMessageChannel,
		tracer:             tracer,
	}, nil
}

// Close stops the MQTT receiver.
func (r *MqttReceiver) Close(ctx context.Context) error {
	r.cancel()
	close(r.mqttMessageChannel)
	return r.connection.Disconnect(ctx)
}

// GetMqttMessageChannel returns the channel for MQTT messages.
//
// The channel is closed when the receiver is closed.
func (r *MqttReceiver) GetMqttMessageChannel() <-chan TraceableMqttPublish {
	return r.mqttMessageChannel
}
