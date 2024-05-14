package database

import (
	"bytes"
	"context"
	"log/slog"
	"path"
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *database) CreateMovementEvent(ctx context.Context, req *CreateMovementEventRequest) (*CreateMovementEventResponse, error) {
	eventID := primitive.NewObjectID()

	// upload the movement picture to the GridFS
	picture := req.MovementPicture

	d.logger.DebugContext(
		ctx, "uploading movement picture",
		slog.Int("size", len(picture)),
	)
	pictureID, err := d.Fs().UploadFromStream(
		getMovementPath(eventID),
		bytes.NewReader(picture),
		options.GridFSUpload().SetMetadata(bson.M{
			"eventID": eventID,
		}),
	)
	if err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to upload movement picture",
			slogext.Error(err),
			slogext.ObjectID("eventID", eventID),
		)
		return nil, err
	}

	event, err := d.Events().InsertOne(ctx, models.Event{
		ID:       eventID,
		Metadata: req.Metadata,
		Type:     models.EventTypeMovement,
		Payload: models.EventPayload{
			MovementPicture: lo.ToPtr(pictureID),
		},
		CreatedAt: time.Now(),
		DeletedAt: nil,
	})
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to insert event", slogext.Error(err))
		return nil, err
	}

	return &CreateMovementEventResponse{
		EventID: event.InsertedID.(primitive.ObjectID),
	}, nil
}

type CreateMovementEventRequest struct {
	Metadata        models.EventMetadata
	MovementPicture []byte
}

type CreateMovementEventResponse struct {
	EventID primitive.ObjectID
}

func (d *database) CreateInvadedEvent(ctx context.Context, req *CreateInvadedEventRequest) (*CreateInvadedEventResponse, error) {
	eventID := primitive.NewObjectID()
	invaders := make([]models.EventInvader, 0, len(req.Invaders))

	// upload the invader to the GridFS
	for _, invader := range req.Invaders {
		invaderID := primitive.NewObjectID()
		picture := invader.Picture
		confidence := invader.Confidence

		d.logger.DebugContext(
			ctx, "uploading invader picture",
			slogext.ObjectID("invaderID", invaderID),
			slog.Int("size", len(picture)),
		)
		err := d.Fs().UploadFromStreamWithID(
			invaderID,
			getInvaderPath(eventID, invaderID),
			bytes.NewReader(picture),
			options.GridFSUpload().SetMetadata(bson.M{
				"eventID":    eventID,
				"confidence": confidence,
			}),
		)
		if err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to upload invader picture",
				slogext.Error(err),
				slogext.ObjectID("eventID", eventID),
			)
			continue
		}

		invaders = append(invaders, models.EventInvader{
			PictureID:  invaderID,
			Confidence: confidence,
		})
	}

	event, err := d.Events().InsertOne(ctx, models.Event{
		ID:       eventID,
		Metadata: req.Metadata,
		Type:     models.EventTypeInvaded,
		Payload: models.EventPayload{
			Invader: lo.ToPtr(invaders),
		},
		CreatedAt: time.Now(),
		DeletedAt: nil,
	})
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to insert event", slogext.Error(err))
		return nil, err
	}

	return &CreateInvadedEventResponse{
		EventID: event.InsertedID.(primitive.ObjectID),
	}, nil
}

type CreateInvadedEventRequest struct {
	Metadata models.EventMetadata
	Invaders []InvaderImageRequest
}

type CreateInvadedEventResponse struct {
	EventID primitive.ObjectID
}

type InvaderImageRequest struct {
	Picture    []byte
	Confidence float64
}

func getMovementPath(eventID primitive.ObjectID) string {
	return path.Join("/events", eventID.Hex(), "movement")
}

func getInvaderPath(eventID, invaderID primitive.ObjectID) string {
	return path.Join("/events", eventID.Hex(), "invaders", invaderID.Hex())
}
