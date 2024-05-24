package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
)

type eventRepositoryEnt struct {
	client *ent.Client
}

type EntRepository interface {
	Client() *ent.Client
}

type EntEventRepository interface {
	EventRepository
	EntRepository
}

func NewEventRepositoryEnt(client *ent.Client) EntEventRepository {
	return &eventRepositoryEnt{client: client}
}

func (r *eventRepositoryEnt) Client() *ent.Client {
	return r.client
}

func (r *eventRepositoryEnt) GetEvent(ctx context.Context, id uuid.UUID) (Event, error) {
	eventDao, err := r.client.Event.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return r.transformEvent(ctx, eventDao, false)
}

func (r *eventRepositoryEnt) ListEvents(ctx context.Context, filter EventListFilter) (*EventListResponse, error) {
	// construct query
	query := r.client.Event.Query()
	query = query.Order(ent.Desc(event.FieldID))

	cursor, err := filter.GetCursorUUID()
	if err != nil {
		if !errors.Is(err, ErrNoCursor) {
			return nil, fmt.Errorf("get cursor uuid: %w", err)
		}
	} else {
		query = query.Where(event.IDGT(cursor))
	}

	query = query.Limit(filter.GetLimit())

	// get response
	eventsDao, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	events := make([]Event, 0, len(eventsDao))
	for _, eventDao := range eventsDao {
		eventModel, err := r.transformEvent(ctx, eventDao, true)
		if err != nil {
			return nil, fmt.Errorf("transform event: %w", err)
		}

		events = append(events, eventModel)
	}

	if len(events) == 0 {
		return &EventListResponse{
			Events: []Event{},
			Pagination: PaginationInfo{
				HasNextPage: false,
				StartCursor: "",
				EndCursor:   "",
			},
		}, nil
	}

	return &EventListResponse{
		Events:     events[:len(events)-1], // we ask limit+1, so we need to remove the last one
		Pagination: GeneratePaginationInfo(events, filter.GetLimit()),
	}, nil
}

func (r *eventRepositoryEnt) transformEvent(ctx context.Context, eventDao *ent.Event, brief bool) (Event, error) {
	metadata := Metadata{
		EventID:   eventDao.ID,
		DeviceID:  eventDao.DeviceID,
		EmittedAt: eventDao.CreatedAt,
	}

	switch eventDao.Type {
	case event.TypeMovement:
		if brief {
			return &BriefEvent{
				Metadata: metadata,
				Type:     EventTypeMovement,
			}, nil
		}

		movementDao, err := eventDao.QueryMovements().Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("query movements: %w", err)
		}

		return &MovementEvent{
			Metadata: metadata,
			Movement: Movement{
				MovementID: movementDao.ID,
				Picture:    movementDao.Picture,
			},
		}, nil
	case event.TypeInvaded:
		if brief {
			return &BriefEvent{
				Metadata: metadata,
				Type:     EventTypeInvaded,
			}, nil
		}

		invadersDao, err := eventDao.QueryInvaders().All(ctx)
		if err != nil {
			return nil, fmt.Errorf("query invaders: %w", err)
		}

		invaders := make([]Invader, 0, len(invadersDao))
		for _, invaderDao := range invadersDao {
			invaders = append(invaders, Invader{
				InvaderID:  invaderDao.ID,
				Picture:    invaderDao.Picture,
				Confidence: invaderDao.Confidence,
			})
		}

		return &InvadedEvent{
			Metadata: metadata,
			Invaders: invaders,
		}, nil
	case event.TypeMove:
		if brief {
			return &BriefEvent{
				Metadata: metadata,
				Type:     EventTypeMove,
			}, nil
		}

		moveDao, err := eventDao.QueryMoves().Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("query moves: %w", err)
		}

		return &MoveEvent{
			Metadata: metadata,
			Move: Move{
				Cycle: moveDao.Cycle,
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown event type: %v", eventDao.Type)
}
