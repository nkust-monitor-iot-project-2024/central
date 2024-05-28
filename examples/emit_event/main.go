package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config := utils.NewConfig()

	amqpAddress := config.String("mq.address")
	if amqpAddress == "" {
		panic("rabbitmq address is not set")
	}

	var bodiesToSend [][]byte

	for i := 1; i <= 7; i++ {
		picture, err := os.ReadFile(fmt.Sprintf("./examples/emit_event/%d.jpg", i))
		if err != nil {
			panic(fmt.Errorf("read picture: %w", err))
		}

		marshalledBody, err := protojson.Marshal(&eventpb.EventMessage{
			Event: &eventpb.EventMessage_MovementInfo{
				MovementInfo: &eventpb.MovementInfo{
					Picture: picture,
				},
			},
		})
		if err != nil {
			panic(fmt.Errorf("marshal event message: %w", err))
		}

		bodiesToSend = append(bodiesToSend, marshalledBody)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := amqp091.Dial(amqpAddress)
			if err != nil {
				panic(fmt.Errorf("connect to rabbitmq: %w", err))
			}
			defer func() {
				_ = conn.Close()
			}()

			ch, err := conn.Channel()
			if err != nil {
				panic(fmt.Errorf("open channel: %w", err))
			}
			defer func(ch *amqp091.Channel) {
				_ = ch.Close()
			}(ch)

			err = ch.ExchangeDeclare(
				"events_topic",
				"topic",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				panic(fmt.Errorf("declare exchange: %w", err))
			}

			for v, body := range bodiesToSend {
				err = ch.Publish("events_topic", "event.v1."+string(models.EventTypeMovement), false, false, amqp091.Publishing{
					Headers:       nil,
					ContentType:   "application/json",
					DeliveryMode:  0,
					Priority:      0,
					CorrelationId: "",
					ReplyTo:       "",
					Expiration:    "",
					MessageId:     uuid.New().String(),
					Timestamp:     time.Now(),
					Type:          "eventpb.EventMessage",
					UserId:        "",
					AppId:         "central/example/emit-event",
					Body:          body,
				})
				if err != nil {
					log.Println(fmt.Errorf("publish message: %w", err))
				}
				fmt.Println("Event emitted", i, v)
			}
		}()
	}

	wg.Wait()
}
