package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := utils.NewConfig()

	var bodiesToSend []*eventpb.EventMessage

	files, err := filepath.Glob("./examples/emit_event/assets/*")
	if err != nil {
		panic(fmt.Errorf("glob pictures: %w", err))
	}

	serverUrlRaw := config.String("mq.mqtt.uri")
	if serverUrlRaw == "" {
		panic("missing configuration: mq.mqtt.uri")
	}

	serverUrl, err := url.Parse(serverUrlRaw)
	if err != nil {
		panic(fmt.Errorf("parse server URL %q: %w", serverUrl, err))
	}

	for _, file := range files {
		picture, err := os.ReadFile(file)
		if err != nil {
			panic(fmt.Errorf("read picture: %w", err))
		}

		bodiesToSend = append(bodiesToSend, &eventpb.EventMessage{
			Event: &eventpb.EventMessage_MovementInfo{
				MovementInfo: &eventpb.MovementInfo{
					Picture:     picture,
					PictureMime: "image/jpeg",
				},
			},
		})
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{serverUrl},
		KeepAlive:  20,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			slog.Info("mqtt connection up")
		},
		OnConnectError: func(err error) {
			slog.Error("error whilst attempting connection", slogext.Error(err))
		},
		ClientConfig: paho.ClientConfig{
			ClientID:      "central/example/emit_mqtt_event",
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
		panic(fmt.Errorf("create connection: %w", err))
	}
	if err := c.AwaitConnection(ctx); err != nil {
		cancel() // failed; cancel the context
		panic(fmt.Errorf("await connection: %w", err))
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, body := range bodiesToSend {
		group.Go(func() error {
			marshalledBody, err := proto.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal body: %w", err)
			}

			publishing, err := c.Publish(ctx, &paho.Publish{
				QoS:   0,
				Topic: "iot/events/v1/movement",
				Properties: &paho.PublishProperties{
					ContentType: "application/x-google-protobuf",
					User: paho.UserPropertiesFromPacketUser([]packets.User{
						{
							Key:   "event_id",
							Value: uuid.Must(uuid.NewV7()).String(),
						},
						{
							Key:   "device_id",
							Value: "central/example/emit-event",
						},
						{
							Key:   "emitted_at",
							Value: time.Now().Format(time.RFC3339Nano),
						},
					}),
				},
				Payload: marshalledBody,
			})
			if err != nil {
				return fmt.Errorf("publish: %w", err)
			}

			slog.Info("event published", slog.Any("publishing", publishing))
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		panic(err)
	}

	if err := c.Disconnect(ctx); err != nil {
		panic(fmt.Errorf("disconnect: %w", err))
	}
}
