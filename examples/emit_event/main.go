package main

import (
	"context"
	"fmt"
	mqevent "github.com/nkust-monitor-iot-project-2024/central/internal/mq/event"
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/mo"
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

	amqp, err := mqv2.NewAmqpWrapper(config)
	if err != nil {
		panic(err)
	}

	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < 10; i++ {
		group.Go(func() error {
			connection, err := amqp.NewConnection()
			if err != nil {
				return fmt.Errorf("new connection: %w", err)
			}

			channel, err := connection.Channel()
			if err != nil {
				return fmt.Errorf("get channel: %w", err)
			}
			defer func(channel *amqp091.Channel) {
				_ = channel.Close()
			}(channel)

			exchangeName, err := mqevent.DeclareEventsTopic(channel)
			if err != nil {
				return fmt.Errorf("declare exchange: %w", err)
			}

			for _, body := range bodiesToSend {
				eventType, publishing, err := mqevent.CreatePublishingEvent(ctx, mqevent.EventPublishingPayload{
					Propagator: nil,
					Metadata: models.Metadata{
						EventID:       uuid.Must(uuid.NewV7()),
						DeviceID:      "central/example/emit-event",
						EmittedAt:     time.Now(),
						ParentEventID: mo.None[uuid.UUID](),
					},
					Event: body,
				})
				if err != nil {
					return fmt.Errorf("create publishing event: %w", err)
				}

				routingKey := mqevent.GetRoutingKey(mo.Some(eventType))

				err = channel.PublishWithContext(
					ctx,
					exchangeName,
					routingKey,
					false, /* mandatory */
					false, /* immediate */
					publishing,
				)
				if err != nil {
					return fmt.Errorf("publish event: %w", err)
				}
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		panic(fmt.Errorf("emit event: %w", err))
	}
}
