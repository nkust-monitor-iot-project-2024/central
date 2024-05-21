package models

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeInvaded  EventType = "invaded"
	EventTypeMovement EventType = "movement"
)

type EventRepository interface {
	Get(eventID uuid.UUID) (*Event, error)
}

type MovementRepository interface {
	GetByEvent(eventID uuid.UUID) (*Movement, error)
}

type InvaderRepository interface {
	ListByEvent(eventID uuid.UUID) ([]*Invader, error)
}

// Event represents an event.
type Event struct {
	EventID   uuid.UUID `json:"event_id"`
	Type      EventType `json:"type"`
	EmittedAt time.Time `json:"emitted_at"`
}

func (e *Event) GetEventID() uuid.UUID {
	return e.EventID
}

func (e *Event) GetType() EventType {
	return e.Type
}

func (e *Event) GetEmittedAt() time.Time {
	return e.EmittedAt
}

func (e *Event) RetrieveMovement(movementRepo MovementRepository) (*Movement, error) {
	return movementRepo.GetByEvent(e.GetEventID())
}

func (e *Event) RetrieveInvaded(invaderRepo InvaderRepository) ([]*Invader, error) {
	return invaderRepo.ListByEvent(e.GetEventID())
}

// Movement represents movement information of an event.
type Movement struct {
	EventID uuid.UUID `json:"event_id"`
	Picture []byte    `json:"picture"`
}

func (m *Movement) GetEventID() uuid.UUID {
	return m.EventID
}

func (m *Movement) GetPicture() []byte {
	return m.Picture
}

func (m *Movement) RetrieveEvent(eventRepo EventRepository) (*Event, error) {
	return eventRepo.Get(m.GetEventID())
}

// Invader represents invaders of an event.
type Invader struct {
	EventID    uuid.UUID `json:"event_id"`
	Picture    []byte    `json:"picture"`
	Confidence float64   `json:"confidence"`
}

func (i *Invader) GetEventID() uuid.UUID {
	return i.EventID
}

func (i *Invader) GetPicture() []byte {
	return i.Picture
}

func (i *Invader) GetConfidence() float64 {
	return i.Confidence
}
