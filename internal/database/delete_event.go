package database

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (d *database) DeleteEvents(ctx context.Context, req *DeleteEventsRequest) ([]primitive.ObjectID, error) {
	// Find the event to delete
	cursor, err := d.Events().Find(ctx, req.Match)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to find event", slogext.Error(err))
		return primitive.NilObjectID, err
	}

	event, err := d.Events().FindOne(ctx, req.Match)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to find event", slogext.Error(err))
		return primitive.NilObjectID, err

	}

	// upload the movement picture to the GridFS
	picture := req.MovementPicture

	d.logger.DebugContext(
		ctx, "uploading movement picture",
		slog.Int("size", len(picture)),
	)
	pictureID, err := d.Fs().UploadFromStream(
		getMovementPath(eventID),
		bytes.NewReader(picture),
	)
	if err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to upload movement picture",
			slogext.Error(err),
			slogext.ObjectID("eventID", eventID),
		)
		return primitive.NilObjectID, err
	}

	event, err := d.Events().InsertOne(ctx, models.Event{
		ID:       eventID,
		Metadata: req.Metadata,
		Type:     models.EventTypeMovement,
		Payload: models.EventPayload{
			MovementPicture: lo.ToPtr(pictureID),
		},
	})
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to insert event", slogext.Error(err))
		return primitive.NilObjectID, err
	}

	return event.InsertedID.(primitive.ObjectID), nil
}

type DeleteEventsRequest struct {
	Match any
}
