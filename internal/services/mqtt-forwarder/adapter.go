package mqtt_forwarder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type Adapter struct {
	rawMqttMessageChannel <-chan TraceableMqttPublish

	tracer trace.Tracer
}

type TraceableEventMessage struct {
	Metadata     models.Metadata
	EventMessage *eventpb.EventMessage

	SpanContext trace.SpanContext
}

func NewAdapter(rawMqttMessage <-chan TraceableMqttPublish) *Adapter {
	return &Adapter{
		rawMqttMessageChannel: rawMqttMessage,
		tracer:                otel.GetTracerProvider().Tracer("mqtt-forwarder"),
	}
}

func (a *Adapter) StartConvert(parentCtx context.Context) <-chan TraceableEventMessage {
	convertedEventMessage := make(chan TraceableEventMessage, 64)

	go func() {
		defer close(convertedEventMessage)
		for {
			select {
			case <-parentCtx.Done():
				return
			case rawMqttMessage := <-a.rawMqttMessageChannel:
				if parentCtx.Err() != nil {
					return
				}

				ctx := trace.ContextWithSpanContext(parentCtx, rawMqttMessage.SpanContext)
				message, err := a.handleMessage(ctx, rawMqttMessage)
				if err != nil {
					continue
				}

				convertedEventMessage <- message
			}
		}
	}()

	return convertedEventMessage
}

func (a *Adapter) handleMessage(ctx context.Context, rawMqttMessage TraceableMqttPublish) (TraceableEventMessage, error) {
	_, span := a.tracer.Start(
		ctx, "mqtt-forwarder/adapter/handle_message",
		trace.WithAttributes(
			attribute.String("mqtt.message.topic", rawMqttMessage.Topic),
			attribute.String("mqtt.message.content_type", rawMqttMessage.Properties.ContentType)))
	defer span.End()

	span.AddEvent("check if content type is supported")
	if rawMqttMessage.Properties.ContentType != "application/x-google-protobuf" {
		span.SetStatus(codes.Error, "unsupported content type")
		span.AddEvent("received content type", trace.WithAttributes(
			attribute.String("content_type", rawMqttMessage.Properties.ContentType)))

		return TraceableEventMessage{}, fmt.Errorf("unsupported content type: %s", rawMqttMessage.Properties.ContentType)
	}

	span.AddEvent("extract metadata from MQTT message")
	eventID := rawMqttMessage.Properties.User.Get("event_id")
	if eventID == "" {
		span.SetStatus(codes.Error, "event_id is missing")

		return TraceableEventMessage{}, errors.New("event_id is missing")
	}
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		span.SetStatus(codes.Error, "parse event_id")
		span.RecordError(err)

		return TraceableEventMessage{}, fmt.Errorf("parse event_id: %w", err)
	}

	deviceID := rawMqttMessage.Properties.User.Get("device_id")
	if deviceID == "" {
		span.SetStatus(codes.Error, "device_id is missing")

		return TraceableEventMessage{}, fmt.Errorf("device_id is missing")
	}

	emittedAt, err := time.Parse(time.RFC3339Nano, rawMqttMessage.Properties.User.Get("emitted_at"))
	if err != nil {
		span.SetStatus(codes.Error, "parse emitted_at")
		span.RecordError(err)

		return TraceableEventMessage{}, fmt.Errorf("parse emitted_at: %w", err)
	}

	parentEventUUID := mo.None[uuid.UUID]()
	parentEventID := rawMqttMessage.Properties.User.Get("parent_event_id")
	if parentEventID != "" {
		span.AddEvent("parsing parent_event_id",
			trace.WithAttributes(attribute.String("parent_event_id", parentEventID)))

		parentEventUUIDValue, err := uuid.Parse(parentEventID)
		if err != nil {
			span.SetStatus(codes.Error, "parse parent_event_id")
			span.RecordError(err)

			return TraceableEventMessage{}, fmt.Errorf("parse parent_event_id: %w", err)
		}

		parentEventUUID = mo.Some[uuid.UUID](parentEventUUIDValue)
	}

	metadata := models.Metadata{
		EventID:       eventUUID,
		DeviceID:      deviceID,
		EmittedAt:     emittedAt,
		ParentEventID: parentEventUUID,
	}

	span.AddEvent("unmarshal event message")
	eventMessage := &eventpb.EventMessage{}
	if err := proto.Unmarshal(rawMqttMessage.Payload, eventMessage); err != nil {
		return TraceableEventMessage{}, fmt.Errorf("unmarshal event message: %w", err)
	}

	span.AddEvent("message constructed")
	span.SetStatus(codes.Ok, "message is successfully handled")

	return TraceableEventMessage{
		Metadata:     metadata,
		EventMessage: eventMessage,
		SpanContext:  span.SpanContext(),
	}, nil
}
