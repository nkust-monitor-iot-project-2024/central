package models

import (
	"github.com/samber/mo"
)

// EventType represents an event type.
type EventType string

const (
	// EventMovementType is the type of movement event.
	EventMovementType EventType = "movement"

	// EventInvadedType is the type of invaded event.
	EventInvadedType EventType = "invaded"
)

type EventRepository interface {
	GetEvent(id string) (*Event, error)
	ListEvents(eventType mo.Option[EventType]) ([]Event, error)
}
