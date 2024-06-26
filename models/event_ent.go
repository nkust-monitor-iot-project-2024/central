package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/samber/mo"
	"go.uber.org/fx"
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

var EventRepositoryEntFx = fx.Module("event-repository-ent", database.EntFx, fx.Provide(NewEventRepositoryEnt))

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
	paginationFilter := filter.Pagination.OrElse(ForwardPagination{})

	query := r.client.Event.Query()
	query = query.Order(ent.Desc(event.FieldID))

	if eventType, ok := filter.EventType.Get(); ok {
		query = query.Where(event.TypeEQ(event.Type(eventType)))
	}

	newerElementQuery := query.Clone()
	olderElementQuery := query.Clone()

	switch pagination := paginationFilter.(type) {
	case ForwardPagination:
		cursor, err := pagination.GetAfterUUID()
		switch {
		case err == nil:
			query = query.Where(event.IDLT(cursor))
		case errors.Is(err, ErrNoCursor): // no cursor, do nothing
		default:
			return nil, fmt.Errorf("get after uuid: %w", err)
		}

		query = query.Limit(pagination.GetFirst())
	case BackwardPagination:
		cursor, err := pagination.GetBeforeUUID()

		switch {
		case err == nil:
			query = query.Where(event.IDGT(cursor))
		case errors.Is(err, ErrNoCursor): // no cursor, do nothing
		default:
			return nil, fmt.Errorf("get before uuid: %w", err)
		}

		query = query.Limit(pagination.GetLast())
	}

	// get response
	eventsDao, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	hasPrevPage, hasNextPage, err := func() (bool, bool, error) {
		if len(eventsDao) == 0 {
			// there may be previous page, but definitively no next page
			return true, false, nil
		}

		hasNewerElement, err := newerElementQuery.
			Where(event.IDGT(eventsDao[0].ID)).
			Exist(ctx)
		if err != nil {
			return false, false, fmt.Errorf("check newer element: %w", err)
		}
		hasOlderElement, err := olderElementQuery.
			Where(event.IDLT(eventsDao[len(eventsDao)-1].ID)).
			Exist(ctx)
		if err != nil {
			return false, false, fmt.Errorf("check older element: %w", err)
		}

		switch paginationFilter.(type) {
		case BackwardPagination:
			return hasOlderElement, hasNewerElement, nil
		case ForwardPagination:
			return hasNewerElement, hasOlderElement, nil
		}

		return false, false, fmt.Errorf("unknown pagination type: %T", paginationFilter)
	}()
	if err != nil {
		return nil, fmt.Errorf("check pagination: %w", err)
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
			Events: events,
			Pagination: PaginationInfo{
				HasPreviousPage: hasPrevPage,
				HasNextPage:     hasNextPage,
				StartCursor:     "",
				EndCursor:       "",
			},
		}, nil
	}

	return &EventListResponse{
		Events: events,
		Pagination: PaginationInfo{
			HasPreviousPage: hasPrevPage,
			HasNextPage:     hasNextPage,
			StartCursor:     UUIDToCursor(eventsDao[0].ID),
			EndCursor:       UUIDToCursor(eventsDao[len(eventsDao)-1].ID),
		},
	}, nil
}

func (r *eventRepositoryEnt) transformEvent(ctx context.Context, eventDao *ent.Event, brief bool) (Event, error) {
	metadata := Metadata{
		EventID:       eventDao.ID,
		DeviceID:      eventDao.DeviceID,
		EmittedAt:     eventDao.CreatedAt,
		ParentEventID: mo.PointerToOption(eventDao.ParentEventID),
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
				MovementID:  movementDao.ID,
				Picture:     movementDao.Picture,
				PictureMime: movementDao.PictureMime,
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
				InvaderID:   invaderDao.ID,
				Picture:     invaderDao.Picture,
				PictureMime: invaderDao.PictureMime,
				Confidence:  invaderDao.Confidence,
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
