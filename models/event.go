package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

// Event represents an event.
type Event interface {
	// GetEventID returns the unique event ID.
	GetEventID() uuid.UUID
	// GetDeviceID returns the device ID that triggers this event.
	GetDeviceID() string
	// GetEmittedAt returns the time when this event is emitted.
	GetEmittedAt() time.Time
	// GetParentEventID returns the parent event ID if this event is a child event.
	GetParentEventID() (uuid.UUID, bool)
	// GetType returns the type of this event.
	GetType() EventType
}

// BriefEvent represents a brief event (that did not contains anything detailed).
type BriefEvent struct {
	Metadata `json:"metadata"`
	Type     EventType `json:"type"`
}

func (b *BriefEvent) GetType() EventType {
	return b.Type
}

// MovementEvent represents a movement event.
type MovementEvent struct {
	Metadata `json:"metadata"`
	Movement Movement `json:"movement"`
}

func (m *MovementEvent) GetType() EventType {
	return EventTypeMovement
}

func (m *MovementEvent) GetPicture() []byte {
	return m.Movement.GetPicture()
}

// InvadedEvent represents an invaded event.
type InvadedEvent struct {
	Metadata `json:"metadata"`
	Invaders []Invader `json:"invaders"`
}

func (i *InvadedEvent) GetType() EventType {
	return EventTypeInvaded
}

func (i *InvadedEvent) GetInvaders() []Invader {
	return i.Invaders
}

// MoveEvent represents a move event.
type MoveEvent struct {
	Metadata `json:"metadata"`
	Move     Move `json:"move"`
}

func (m *MoveEvent) GetType() EventType {
	return EventTypeMove
}

func (m *MoveEvent) GetCycle() float64 {
	return m.Move.GetCycle()
}

// Metadata represents an event.
type Metadata struct {
	EventID       uuid.UUID            `json:"event_id"`
	DeviceID      string               `json:"device_id"`
	EmittedAt     time.Time            `json:"emitted_at"`
	ParentEventID mo.Option[uuid.UUID] `json:"parent_event_id"`
}

func (e *Metadata) GetEventID() uuid.UUID {
	return e.EventID
}

func (e *Metadata) GetDeviceID() string {
	return e.DeviceID
}

func (e *Metadata) GetEmittedAt() time.Time {
	return e.EmittedAt
}

func (e *Metadata) GetParentEventID() (uuid.UUID, bool) {
	return e.ParentEventID.Get()
}

// Movement represents movement information of an event.
type Movement struct {
	MovementID uuid.UUID `json:"movement_id"`
	Picture    []byte    `json:"picture"`
}

func (m *Movement) GetMovementID() uuid.UUID {
	return m.MovementID
}

func (m *Movement) GetPicture() []byte {
	return m.Picture
}

// Invader represents invaders of an event.
type Invader struct {
	InvaderID  uuid.UUID `json:"invader_id"`
	Picture    []byte    `json:"picture"`
	Confidence float64   `json:"confidence"`
}

func (i *Invader) GetInvaderID() uuid.UUID {
	return i.InvaderID
}

func (i *Invader) GetPicture() []byte {
	return i.Picture
}

func (i *Invader) GetConfidence() float64 {
	return i.Confidence
}

// Move represents a move of an event.
type Move struct {
	MoveID uuid.UUID `json:"move_id"`
	Cycle  float64   `json:"cycle"`
}

func (m *Move) GetMoveID() uuid.UUID {
	return m.MoveID
}

func (m *Move) GetCycle() float64 {
	return m.Cycle
}
