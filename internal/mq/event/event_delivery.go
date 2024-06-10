package mqevent

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq/amqpext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"
)

// AmqpEventMessageDelivery is a wrapper around amqp091.Delivery that includes a function
// to extract the metadata and the body from Delivery, and provide the span extractor of
// the Delivery.
type AmqpEventMessageDelivery struct {
	amqp091.Delivery
}

func (w *AmqpEventMessageDelivery) Extract(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	supplier := amqpext.NewHeaderSupplier(w.Headers)

	return propagator.Extract(ctx, supplier)
}

func (w *AmqpEventMessageDelivery) Metadata() (models.Metadata, error) {
	eventID := w.Delivery.MessageId
	if eventID == "" {
		return models.Metadata{}, errors.New("missing or invalid MessageId (-> event_id)")
	}
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("parse event_id: %w", err)
	}

	deviceID := w.Delivery.AppId
	if deviceID == "" {
		return models.Metadata{}, errors.New("missing or invalid AppId (-> device_id)")
	}

	eventTs := w.Delivery.Timestamp

	metadata := models.Metadata{
		EventID:       eventUUID,
		DeviceID:      deviceID,
		EmittedAt:     eventTs,
		ParentEventID: mo.None[uuid.UUID](),
	}

	if parentEventID, ok := w.Headers["parent_event_id"]; ok {
		parentEventUUID, err := uuid.Parse(parentEventID.(string))
		if err != nil {
			return models.Metadata{}, fmt.Errorf("parse parent_event_id: %w", err)
		}

		metadata.ParentEventID = mo.Some[uuid.UUID](parentEventUUID)
	}

	return metadata, nil
}

func (w *AmqpEventMessageDelivery) Body() (*eventpb.EventMessage, error) {
	if w.Delivery.ContentType != "application/x-google-protobuf" || w.Delivery.Type != "eventpb.EventMessage" {
		return nil, errors.New("invalid header")
	}

	event := &eventpb.EventMessage{}

	if err := proto.Unmarshal(w.Delivery.Body, event); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return event, nil
}

func (w *AmqpEventMessageDelivery) Reject(requeue bool) error {
	return w.Delivery.Reject(requeue)
}

func (w *AmqpEventMessageDelivery) Ack() error {
	return w.Delivery.Ack(false)
}
