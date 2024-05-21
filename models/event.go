package models

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an event.
type Event interface {
	GetEventID() uuid.UUID
	GetEmittedAt() time.Time
}

// MovementEvent represents a movement event.
type MovementEvent struct {
	Metadata Metadata `json:"metadata"`
	Movement Movement `json:"movement"`
}

func (m *MovementEvent) GetEventID() uuid.UUID {
	return m.Metadata.GetEventID()
}

func (m *MovementEvent) GetEmittedAt() time.Time {
	return m.Metadata.GetEmittedAt()
}

func (m *MovementEvent) GetPicture() []byte {
	return m.Movement.GetPicture()
}

// InvadedEvent represents an invaded event.
type InvadedEvent struct {
	Metadata Metadata  `json:"metadata"`
	Invaders []Invader `json:"invaders"`
}

func (i *InvadedEvent) GetEventID() uuid.UUID {
	return i.Metadata.GetEventID()
}

func (i *InvadedEvent) GetEmittedAt() time.Time {
	return i.Metadata.GetEmittedAt()
}

func (i *InvadedEvent) GetInvaders() []Invader {
	return i.Invaders
}

// Metadata represents an event.
type Metadata struct {
	EventID   uuid.UUID `json:"event_id"`
	EmittedAt time.Time `json:"emitted_at"`
}

func (e *Metadata) GetEventID() uuid.UUID {
	return e.EventID
}

func (e *Metadata) GetEmittedAt() time.Time {
	return e.EmittedAt
}

// Movement represents movement information of an event.
type Movement struct {
	Picture []byte `json:"picture"`
}

func (m *Movement) GetPicture() []byte {
	return m.Picture
}

// Invader represents invaders of an event.
type Invader struct {
	Picture    []byte  `json:"picture"`
	Confidence float64 `json:"confidence"`
}

func (i *Invader) GetPicture() []byte {
	return i.Picture
}

func (i *Invader) GetConfidence() float64 {
	return i.Confidence
}
