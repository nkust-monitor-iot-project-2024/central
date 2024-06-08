// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type EventDetails interface {
	IsEventDetails()
}

type Event struct {
	EventID       uuid.UUID    `json:"eventID"`
	DeviceID      string       `json:"deviceID"`
	Timestamp     time.Time    `json:"timestamp"`
	ParentEventID *uuid.UUID   `json:"parentEventID,omitempty"`
	ParentEvent   *Event       `json:"parentEvent,omitempty"`
	Type          EventType    `json:"type"`
	Details       EventDetails `json:"details,omitempty"`
}

type EventConnection struct {
	Edges    []*EventEdge `json:"edges"`
	PageInfo *PageInfo    `json:"pageInfo"`
}

type EventEdge struct {
	Cursor string `json:"cursor"`
	Node   *Event `json:"node"`
}

type InvadedEvent struct {
	Invaders []*Invader `json:"invaders"`
}

func (InvadedEvent) IsEventDetails() {}

type Invader struct {
	InvaderID      uuid.UUID `json:"invaderID"`
	EncodedPicture string    `json:"encodedPicture"`
	PictureMime    string    `json:"pictureMime"`
	Confidence     float64   `json:"confidence"`
}

type Movement struct {
	MovementID     uuid.UUID `json:"movementID"`
	EncodedPicture string    `json:"encodedPicture"`
	PictureMime    string    `json:"pictureMime"`
}

type MovementEvent struct {
	Movement *Movement `json:"movement"`
}

func (MovementEvent) IsEventDetails() {}

type PageInfo struct {
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	HasNextPage     bool    `json:"hasNextPage"`
}

type Query struct {
}

type EventType string

const (
	EventTypeMovement EventType = "MOVEMENT"
	EventTypeInvaded  EventType = "INVADED"
)

var AllEventType = []EventType{
	EventTypeMovement,
	EventTypeInvaded,
}

func (e EventType) IsValid() bool {
	switch e {
	case EventTypeMovement, EventTypeInvaded:
		return true
	}
	return false
}

func (e EventType) String() string {
	return string(e)
}

func (e *EventType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EventType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventType", str)
	}
	return nil
}

func (e EventType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
