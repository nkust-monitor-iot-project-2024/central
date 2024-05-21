package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
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

var (
	ErrEventDeleted = errors.New("event is deleted")
)

// EventProtoResolver is a resolver for the event proto.
//
// The resolver is used to resolve the picture ID to the actual picture.
// If the resolver is nil, the picture will be nil.
type EventProtoResolver struct {
	InvaderPictureResolver  func(primitive.ObjectID) ([]byte, error)
	MovementPictureResolver func(primitive.ObjectID) ([]byte, error)
}

func (e Event) ToProto(resolver EventProtoResolver) (*eventpb.Event, error) {
	if e.DeletedAt != nil {
		return nil, ErrEventDeleted
	}

	metadata, err := e.Metadata.ToProto()
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}

	switch e.Type {
	case EventTypeInvaded:
		if e.Payload.Invader == nil {
			return nil, errors.New("invader is not in payload")
		}

		invaders := make([]*eventpb.Invader, len(*e.Payload.Invader))
		for i, invader := range *e.Payload.Invader {
			pb, err := invader.ToProto(resolver.InvaderPictureResolver)
			if err != nil {
				return nil, fmt.Errorf("invader[%d]: %w", i, err)
			}
			invaders[i] = pb
		}

		return &eventpb.Event{
			Id:        e.ID.Hex(),
			CreatedAt: timestamppb.New(e.CreatedAt),
			Event: &eventpb.Event_EventInvaded{
				EventInvaded: &eventpb.EventInvaded{
					Metadata: metadata,
					Invaders: invaders,
				},
			},
		}, nil
	case EventTypeMovement:
		if e.Payload.MovementPicture == nil {
			return nil, errors.New("movementPicture is not in payload")
		}

		var resolvedPicture []byte
		if resolver.MovementPictureResolver != nil {
			var err error
			resolvedPicture, err = resolver.MovementPictureResolver(*e.Payload.MovementPicture)
			if err != nil {
				return nil, fmt.Errorf("resolve movement picture: %w", err)
			}
		}

		return &eventpb.Event{
			Id:        e.ID.Hex(),
			CreatedAt: timestamppb.New(e.CreatedAt),
			Event: &eventpb.Event_EventMovement{
				EventMovement: &eventpb.EventMovement{
					Metadata: metadata,
					Picture:  resolvedPicture,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", e.Type)
	}
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
		"movementDetectedInvader": EventMovementInvaderDetectedJsonSchema,
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

func (m EventMetadata) ToProto() (*eventpb.EventMetadata, error) {
	return &eventpb.EventMetadata{
		DeviceId:  m.DeviceID,
		Timestamp: timestamppb.New(m.Timestamp),
	}, nil
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

// ToProto converts the EventInvader to the protobuf representation.
//
// The resolver function is used to resolve the picture ID to the actual picture.
// If the resolver is nil, the picture will be nil.
func (e EventInvader) ToProto(resolver func(primitive.ObjectID) ([]byte, error)) (*eventpb.Invader, error) {
	var picture []byte
	if resolver != nil {
		var err error
		picture, err = resolver(e.PictureID)
		if err != nil {
			return nil, fmt.Errorf("resolve picture: %w", err)
		}
	}

	return &eventpb.Invader{
		Picture:    picture,
		Confidence: float32(e.Confidence),
	}, nil
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
