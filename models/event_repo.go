package models

import (
	"context"
	"encoding/hex"
	"fmt"

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
)

type EventRepository interface {
	GetEvent(ctx context.Context, id uuid.UUID) (Event, error)
	ListEvents(ctx context.Context, filter EventListFilter) (*EventListResponse, error)
}

type EventListFilter struct {
	Limit     int                  `json:"limit"`
	Cursor    string               `json:"cursor"`
	EventType mo.Option[EventType] `json:"event_type"`
}

func (f *EventListFilter) GetLimit() int {
	const defaultLimit = 25
	const maxLimit = 50

	if f.Limit <= 0 {
		return defaultLimit
	}

	if f.Limit > 50 {
		return maxLimit
	}

	return f.Limit
}

func (f *EventListFilter) GetCursor() (string, bool) {
	return f.Cursor, f.Cursor != ""
}

func (f *EventListFilter) GetEventType() mo.Option[EventType] {
	return f.EventType
}

type EventListResponse struct {
	Events     []Event
	Pagination PaginationInfo
}

type PaginationInfo struct {
	HasNextPage bool `json:"has_next_page"`

	StartCursor string `json:"start_cursor"`
	EndCursor   string `json:"end_cursor"`
}

// GeneratePaginationInfo generates pagination info from a list of events.
func GeneratePaginationInfo[T Event](events []T, limit int) PaginationInfo {
	// If there are no events, there is no next page.
	if len(events) == 0 {
		return PaginationInfo{
			HasNextPage: false,
			StartCursor: "",
			EndCursor:   "",
		}
	}

	// "events" = [ (events to print), [(next event)] ]
	// If (next event) is presented, there is a next page.
	hasNextPage := len(events) > limit

	startCursor := UUIDToCursor(events[0].GetEventID())

	// finalEvent
	// [ (events to print), (events to print), [(next event)] ]
	//   ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ / limit       / len-1
	// or
	// [ (events to print)    				                  ]
	//   ~~~~~~~~~~~~~~~~~/len-1              /limit
	finalEvent := events[min(limit-1, len(events)-1)]
	endCursor := UUIDToCursor(finalEvent.GetEventID())

	return PaginationInfo{
		HasNextPage: hasNextPage,
		StartCursor: startCursor,
		EndCursor:   endCursor,
	}
}

func UUIDToCursor(id uuid.UUID) string {
	return hex.EncodeToString(id[:])
}

func CursorToUUID(cursor string) (uuid.UUID, error) {
	decodedUuidBytes, err := hex.DecodeString(cursor)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("decode cursor: %w", err)
	}

	decodedUuid, err := uuid.FromBytes(decodedUuidBytes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("decode cursor: %w", err)
	}

	return decodedUuid, nil
}
