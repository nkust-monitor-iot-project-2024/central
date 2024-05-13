package models

import (
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event.
type Event struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	// Metadata is the metadata of the event.
	Metadata EventMetadata `bson:"metadata"`

	// Type is the type of the event.
	Type string `bson:"type"`

	// Payload is the payload of the event.
	Payload EventPayload `bson:"payload,omitempty"`
}

// EventPayload represents a typed event payload.
type EventPayload struct {
	// InvaderPicture is the picture of the invader.
	//
	// This field is only used for the "invaded" event.
	InvaderPicture []primitive.ObjectID `bson:"invaded,omitempty"`
}

// EventMetadata represents the metadata of the event.
type EventMetadata struct {
	// DeviceID is the unique identifier of the device.
	DeviceID string `bson:"device_id"`

	// Timestamp is the timestamp of the event.
	Timestamp time.Time `bson:"timestamp"`
}

func EventMetadataFromProto(pb *eventpb.EventMetadata) EventMetadata {
	return EventMetadata{
		DeviceID:  pb.GetDeviceId(),
		Timestamp: pb.GetTimestamp().AsTime(),
	}
}
