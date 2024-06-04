package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
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

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			conn, err := mq.ConnectAmqp(config)
			if err != nil {
				panic(fmt.Errorf("connect to rabbitmq: %w", err))
			}
			defer func() {
				_ = conn.Close()
			}()

			for j, body := range bodiesToSend {
				err := conn.PublishEvent(ctx, models.Metadata{
					EventID:       uuid.Must(uuid.NewV7()),
					DeviceID:      "central/example/emit-event",
					EmittedAt:     time.Now(),
					ParentEventID: mo.None[uuid.UUID](),
				}, body)
				if err != nil {
					log.Println(fmt.Errorf("publish event: %w", err))
				}

				fmt.Println("Event emitted", i, j)
			}
		}()
	}

	wg.Wait()
}
