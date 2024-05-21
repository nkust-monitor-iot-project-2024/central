package database

import (
	"context"
	"fmt"
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

func (d *database) DeleteEvents(ctx context.Context, req *DeleteEventsRequest) (*DeleteEventsResponse, error) {
	// Find the event to delete
	cursor, err := d.Events().Find(ctx, withoutDeleted(req.Filter))
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to find event", slogext.Error(err))
		return nil, fmt.Errorf("find event: %w", err)
	}

	var deletedEvents []primitive.ObjectID

	for cursor.Next(ctx) {
		var event models.Metadata

		if err := cursor.Decode(&event); err != nil {
			d.logger.ErrorContext(ctx, "failed to decode event", slogext.Error(err))
			return nil, fmt.Errorf("decode event: %w", err)
		}

		d.logger.DebugContext(ctx, "deleting event", slogext.ObjectID("eventID", event.ID))

		// Find the blob in the event, and delete it.
		fsCursor, err := d.Fs().FindContext(ctx, bson.M{"metadata.eventID": event.ID})
		if err != nil {
			d.logger.ErrorContext(ctx, "failed to find corresponding blob", slogext.Error(err), slogext.ObjectID("eventID", event.ID))
			return nil, fmt.Errorf("find corresponding blob: %w", err)
		}

		for fsCursor.Next(ctx) {
			var result gridfs.File

			if err := fsCursor.Decode(&result); err != nil {
				d.logger.ErrorContext(ctx, "failed to decode blob", slogext.Error(err))
				return nil, fmt.Errorf("decode blob: %w", err)
			}

			if err := d.Fs().Delete(result.ID); err != nil {
				d.logger.ErrorContext(ctx, "failed to delete blob", slogext.Error(err))
				return nil, fmt.Errorf("delete blob: %w", err)
			}
		}

		// Delete the event
		if _, err := d.Events().UpdateByID(ctx, event.ID, bson.M{
			"$set": bson.M{"deletedAt": time.Now()},
		}); err != nil {
			d.logger.ErrorContext(ctx, "failed to soft-delete event", slogext.Error(err))
			return nil, fmt.Errorf("soft-delete event: %w", err)
		}

		deletedEvents = append(deletedEvents, event.ID)
	}

	return &DeleteEventsResponse{
		DeletedEventID: deletedEvents,
	}, nil
}

type DeleteEventsRequest struct {
	Filter any
}

type DeleteEventsResponse struct {
	DeletedEventID []primitive.ObjectID
}
