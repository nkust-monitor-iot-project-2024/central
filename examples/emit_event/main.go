package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
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

	conn, err := amqp091.Dial(amqpAddress)
	if err != nil {
		panic(fmt.Errorf("connect to rabbitmq: %w", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Errorf("open channel: %w", err))
	}

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

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("declare queue: %w", err))
	}

	err = ch.QueueBind(
		queue.Name,
		"event.v1.movement",
		"events_topic",
		false,
		nil)
	if err != nil {
		panic(fmt.Errorf("bind queue: %w", err))
	}

	marshalledBody, err := protojson.Marshal(&eventpb.EventMessage{
		Event: &eventpb.EventMessage_MovementInfo{
			MovementInfo: &eventpb.MovementInfo{
				Picture: []byte{},
			},
		},
	})
	if err != nil {
		panic(fmt.Errorf("marshal event message: %w", err))
	}

	err = ch.Publish("events_topic", "event.v1.movement", false, false, amqp091.Publishing{
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
		Body:          marshalledBody,
	})
	if err != nil {
		panic(fmt.Errorf("publish message: %w", err))
	}

	fmt.Println("Event emitted")
}
