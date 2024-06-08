package graph

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	gqlModel "github.com/nkust-monitor-iot-project-2024/central/graph/model"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/samber/lo"
)

func ModelToGql(model models.Event) (*gqlModel.Event, error) {
	eventType, err := EventTypeToGql(model.GetType())
	if err != nil {
		return nil, fmt.Errorf("event type: %w", err)
	}

	details, err := EventDetailsToGql(model)
	if err != nil {
		return nil, fmt.Errorf("event details: %w", err)
	}

	var parentEventID *uuid.UUID
	if parentEventIDResponse, ok := model.GetParentEventID(); ok {
		parentEventID = &parentEventIDResponse
	}

	return &gqlModel.Event{
		EventID:       model.GetEventID(),
		DeviceID:      model.GetDeviceID(),
		Timestamp:     model.GetEmittedAt(),
		ParentEventID: parentEventID,
		ParentEvent:   nil, // resolver handles this
		Type:          eventType,
		Details:       details,
	}, nil
}

func EventListResponseToGql(response *models.EventListResponse) (*gqlModel.EventConnection, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	converted := make([]*gqlModel.EventEdge, 0, len(response.Events))

	for _, model := range response.Events {
		convertedModel, err := ModelToGql(model)
		if err != nil {
			return nil, fmt.Errorf("convert model to gql: %w", err)
		}

		cursor := models.UUIDToCursor(model.GetEventID())
		converted = append(converted, &gqlModel.EventEdge{
			Cursor: cursor,
			Node:   convertedModel,
		})
	}

	return &gqlModel.EventConnection{
		Edges: converted,
		PageInfo: &gqlModel.PageInfo{
			StartCursor:     lo.EmptyableToPtr(response.Pagination.StartCursor),
			EndCursor:       lo.EmptyableToPtr(response.Pagination.EndCursor),
			HasPreviousPage: response.Pagination.HasPreviousPage,
			HasNextPage:     response.Pagination.HasNextPage,
		},
	}, nil
}

func EventTypeFromGql(gql gqlModel.EventType) (models.EventType, error) {
	switch gql {
	case gqlModel.EventTypeMovement:
		return models.EventTypeMovement, nil
	case gqlModel.EventTypeInvaded:
		return models.EventTypeInvaded, nil
	}

	return "", fmt.Errorf("unknown event type: %s", gql)
}

func EventTypeToGql(model models.EventType) (gqlModel.EventType, error) {
	switch model {
	case models.EventTypeMovement:
		return gqlModel.EventTypeMovement, nil
	case models.EventTypeInvaded:
		return gqlModel.EventTypeInvaded, nil
	}

	return "", fmt.Errorf("unknown event type: %s", model)
}

func EventDetailsToGql(model models.Event) (gqlModel.EventDetails, error) {
	switch model := model.(type) {
	case *models.BriefEvent:
		return nil, nil
	case *models.MovementEvent:
		picture, mime := model.Movement.GetPicture()
		encodedPicture := base64.StdEncoding.Strict().EncodeToString(picture)

		return &gqlModel.MovementEvent{
			Movement: &gqlModel.Movement{
				MovementID:     model.Movement.GetMovementID(),
				EncodedPicture: encodedPicture,
				PictureMime:    mime,
			}}, nil
	case *models.InvadedEvent:
		convertedInvadedEvent := make([]*gqlModel.Invader, 0, len(model.GetInvaders()))

		for _, invader := range model.GetInvaders() {
			picture, mime := invader.GetPicture()
			encodedPicture := base64.StdEncoding.Strict().EncodeToString(picture)

			convertedInvadedEvent = append(convertedInvadedEvent, &gqlModel.Invader{
				InvaderID:      invader.GetInvaderID(),
				EncodedPicture: encodedPicture,
				PictureMime:    mime,
			})
		}

		return &gqlModel.InvadedEvent{
			Invaders: convertedInvadedEvent,
		}, nil
	}

	return nil, fmt.Errorf("unknown event type: %s [%T]", model.GetType(), model)
}
