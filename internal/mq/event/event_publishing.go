package mqevent

import (
	"context"
	"fmt"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"
)

// EventPublishingPayload is the payload for the event publishing.
type EventPublishingPayload struct {
	// Propagator is the propagator used to inject the context into the message headers.
	Propagator propagation.TextMapPropagator
	// Metadata is the metadata of the event.
	Metadata models.Metadata
	// Event is the event message to be published.
	Event *eventpb.EventMessage
}

// CreatePublishingEvent creates a new AMQP publishing message from the given event message.
func CreatePublishingEvent(ctx context.Context, payload EventPublishingPayload) (eventType models.EventType, publishing amqp091.Publishing, err error) {
	switch payload.Event.GetEvent().(type) {
	case *eventpb.EventMessage_MovementInfo:
		eventType = models.EventTypeMovement
	case *eventpb.EventMessage_InvadedInfo:
		eventType = models.EventTypeInvaded
	default:
		return eventType, amqp091.Publishing{}, fmt.Errorf("unknown event type: %T", payload.Event.GetEvent())
	}

	header := amqp091.Table{}

	// Inject the context into the header, so we can trace the span across the system boundaries.
	if payload.Propagator != nil {
		payload.Propagator.Inject(ctx, amqpext.NewHeaderSupplier(header))
	}

	if parentEventID, ok := payload.Metadata.GetParentEventID(); ok {
		header["parent_event_id"] = parentEventID.String()
	}

	marshalledBody, err := proto.Marshal(payload.Event)
	if err != nil {
		return eventType, amqp091.Publishing{}, fmt.Errorf("marshal event: %w", err)
	}

	return eventType, amqp091.Publishing{
		ContentType: "application/x-google-protobuf",
		MessageId:   payload.Metadata.GetEventID().String(),
		Timestamp:   payload.Metadata.GetEmittedAt(),
		Type:        "eventpb.EventMessage",
		AppId:       payload.Metadata.GetDeviceID(),
		Headers:     header,
		Body:        marshalledBody,
	}, nil
}
