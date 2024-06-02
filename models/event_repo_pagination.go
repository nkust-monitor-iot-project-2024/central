package models

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var ErrNoCursor = errors.New("no cursor")

type Pagination interface {
	IsPagination()
}

type ForwardPagination struct {
	First int    `json:"first"`
	After string `json:"after"`
}

func (f ForwardPagination) IsPagination() {}

func (f ForwardPagination) GetFirst() int {
	return determineLimit(f.First)
}

func (f ForwardPagination) GetAfterUUID() (uuid.UUID, error) {
	return parseCursor(f.After)
}

type BackwardPagination struct {
	Last   int    `json:"last"`
	Before string `json:"before"`
}

func (f BackwardPagination) IsPagination() {}

func (f BackwardPagination) GetLast() int {
	return determineLimit(f.Last)
}

func (f BackwardPagination) GetBeforeUUID() (uuid.UUID, error) {
	return parseCursor(f.Before)
}

func determineLimit(limit int) int {
	if limit == 0 {
		return 25
	}

	return min(limit, 50)
}

func parseCursor(cursor string) (uuid.UUID, error) {
	if cursor == "" {
		return uuid.UUID{}, ErrNoCursor
	}

	return CursorToUUID(cursor)
}

func (f *EventListFilter) GetEventType() mo.Option[EventType] {
	return f.EventType
}

type EventListResponse struct {
	Events     []Event
	Pagination PaginationInfo
}

type PaginationInfo struct {
	HasPreviousPage bool `json:"has_previous_page"`
	HasNextPage     bool `json:"has_next_page"`

	StartCursor string `json:"start_cursor"`
	EndCursor   string `json:"end_cursor"`
}

// PaginationContext is the abstract logic of pagination.
//
// It gives some questions that you should determine from your database,
// like "is there n+1 element?"
// Then, you can generate pagination info from the list of events.
type PaginationContext struct {
	AnyElementBeforeCursor bool
	AnyElementAfterCursor  bool
	Limit                  int
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
