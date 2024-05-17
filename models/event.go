package models

import (
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson"
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
	CreatedAt time.Time `bson:"createdAt"`
	// DeletedAt is the timestamp when the record was deleted.
	DeletedAt *time.Time `bson:"deletedAt,omitempty"`
}

var EventJsonSchema = bson.M{
	"description": "A event",
	"bsonType":    "object",
	"required":    bson.A{"_id", "metadata", "type", "createdAt"},
	"properties": bson.M{
		"_id": bson.M{
			"bsonType": "objectId",
		},
		"metadata": EventMetadataJsonSchema,
		"type": bson.M{
			"bsonType": "string",
			"enum":     []string{"invaded", "movement"},
		},
		"payload": EventPayloadJsonSchema,
		"createdAt": bson.M{
			"bsonType": "timestamp",
		},
		"deletedAt": bson.M{
			"bsonType": "timestamp",
		},
	},
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

var EventPayloadJsonSchema = bson.M{
	"description": "The payload of the event",
	"bsonType":    "object",
	"properties": bson.M{
		"movementPicture": bson.M{
			"bsonType": "objectId",
		},
		"movementDetectedInvader": bson.M{
			"bsonType": "object",
		},
		"invader": bson.M{
			"bsonType": "array",
			"items":    EventInvaderJsonSchema,
		},
	},
}

// EventMovementInvaderDetected represents the details of the movement event.
type EventMovementInvaderDetected struct {
	// Invaded is a boolean that indicates whether the event is an invasion.
	Invaded bool `bson:"invaded"`

	// EventID is the ID to the invaded event.
	EventID *primitive.ObjectID `bson:"invaderID"`
}

var EventMovementInvaderDetectedJsonSchema = bson.M{
	"description": "The details of the movement event",
	"bsonType":    "object",
	"required":    bson.A{"invaded"},
	"properties": bson.M{
		"invaded": bson.M{
			"bsonType": "bool",
		},
		"eventID": bson.M{
			"bsonType": "objectId",
		},
	},
}

// EventMetadata represents the metadata of the event.
type EventMetadata struct {
	// DeviceID is the unique identifier of the device.
	DeviceID string `bson:"deviceID"`

	// Timestamp is the timestamp of the event.
	Timestamp time.Time `bson:"timestamp"`
}

var EventMetadataJsonSchema = bson.M{
	"description": "The metadata of the event",
	"bsonType":    "object",
	"required":    bson.A{"deviceID", "timestamp"},
	"properties": bson.M{
		"deviceID": bson.M{
			"bsonType": "string",
		},
		"timestamp": bson.M{
			"bsonType": "timestamp",
		},
	},
	"additionalProperties": true,
}

// EventInvader represents the invader (picture and confidence) of the event.
type EventInvader struct {
	// Picture is the picture of the invader.
	PictureID primitive.ObjectID `bson:"pictureID"`

	// Confidence is the confidence of the invader.
	Confidence float64 `bson:"confidence"`
}

var EventInvaderJsonSchema = bson.M{
	"description": "The invader of the event",
	"bsonType":    "object",
	"required":    bson.A{"picture", "confidence"},
	"properties": bson.M{
		"pictureID": bson.M{
			"bsonType": "objectId",
		},
		"confidence": bson.M{
			"bsonType": "double",
		},
	},
}

func EventMetadataFromProto(pb *eventpb.EventMetadata) EventMetadata {
	return EventMetadata{
		DeviceID:  pb.GetDeviceId(),
		Timestamp: pb.GetTimestamp().AsTime(),
	}
}
