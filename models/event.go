package models

import (
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType string

const (
	EventTypeInvaded  EventType = "invaded"
	EventTypeMovement EventType = "movement"
)

// Event represents an event.
type Event struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	// Metadata is the metadata of the event.
	Metadata EventMetadata `bson:"metadata"`

	// Type is the type of the event.
	Type EventType `bson:"type"`
	// Payload is the payload of the event.
	Payload EventPayload `bson:"payload,omitempty"`

	// CreatedAt is the timestamp when the record was created.
	CreatedAt time.Time `bson:"created_at"`
	// DeletedAt is the timestamp when the record was deleted.
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

// EventPayload represents a typed event payload.
type EventPayload struct {
	// MovementPicture is the picture of the movement.
	MovementPicture *primitive.ObjectID `bson:"movementPicture,omitempty"`
	// MovementDetectedInvader indicates if the movement event is an invasion.
	//
	// If this event has not been handled yet, this field will be nil.
	//
	// This field is only used for the "movement" event.
	MovementDetectedInvader *EventMovementInvaderDetected `bson:"movementDetectedInvader,omitempty"`
	// Invader is the picture of the invader.
	//
	// This field is only used for the "invaded" event.
	Invader *[]EventInvader `bson:"invaded,omitempty"`
}

// EventMovementInvaderDetected represents the details of the movement event.
type EventMovementInvaderDetected struct {
	// Invaded is a boolean that indicates whether the event is an invasion.
	Invaded bool `bson:"invaded"`

	// EventID is the ID to the invaded event.
	EventID *primitive.ObjectID `bson:"invader_id"`
}

// EventMetadata represents the metadata of the event.
type EventMetadata struct {
	// DeviceID is the unique identifier of the device.
	DeviceID string `bson:"device_id"`

	// Timestamp is the timestamp of the event.
	Timestamp time.Time `bson:"timestamp"`
}

// EventInvader represents the invader (picture and confidence) of the event.
type EventInvader struct {
	// Picture is the picture of the invader.
	PictureID primitive.ObjectID `bson:"picture"`

	// Confidence is the confidence of the invader.
	Confidence float64 `bson:"confidence"`
}

func EventMetadataFromProto(pb *eventpb.EventMetadata) EventMetadata {
	return EventMetadata{
		DeviceID:  pb.GetDeviceId(),
		Timestamp: pb.GetTimestamp().AsTime(),
	}
}
