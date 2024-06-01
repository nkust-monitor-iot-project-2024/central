package models

import (
	"context"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/samber/mo"
)

// EventType represents an event type.
type EventType event.Type

const (
	// EventTypeMovement is the type of movement event.
	EventTypeMovement = EventType(event.TypeMovement)

	// EventTypeInvaded is the type of invaded event.
	EventTypeInvaded = EventType(event.TypeInvaded)

	// EventTypeMove is the type of move event.
	EventTypeMove = EventType(event.TypeMove)
)

type EventRepository interface {
	GetEvent(ctx context.Context, id uuid.UUID) (Event, error)
	ListEvents(ctx context.Context, filter EventListFilter) (*EventListResponse, error)
}

type EventListFilter struct {
	Pagination mo.Option[Pagination]
	EventType  mo.Option[EventType]
}
