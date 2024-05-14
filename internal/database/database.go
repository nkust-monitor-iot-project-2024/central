package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection interface {
	Events() *mongo.Collection
	Fs() *gridfs.Bucket
}

type DataLayer interface {
	CreateMovementEvent(context.Context, *CreateMovementEventRequest) (*CreateMovementEventResponse, error)
	CreateInvadedEvent(context.Context, *CreateInvadedEventRequest) (*CreateInvadedEventResponse, error)

	GetMovementPicture(context.Context, *GetMovementPictureRequest) (*GetMovementPictureResponse, error)
	GetInvaderPicture(context.Context, *GetInvaderPictureRequest) (*GetInvaderPictureResponse, error)

	DeleteEvents(context.Context, *DeleteEventsRequest) ([]primitive.ObjectID, error)
}

type Database interface {
	Collection
	DataLayer
}

type database struct {
	db     *mongo.Database
	logger *slog.Logger
}

func (d *database) Events() *mongo.Collection {
	return d.db.Collection("events")
}

func (d *database) Fs() *gridfs.Bucket {
	bucket, _ := gridfs.NewBucket(d.db)
	return bucket
}

func ConnectByConfig(conf utils.Config, logger *slog.Logger) (Database, error) {
	dbLogger := logger.With(slog.String("service", "_database"))

	uri := conf.String("mongo.uri")
	if uri == "" {
		return nil, errors.New("missing mongo URI (mongo.uri)")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	return &database{
		db:     client.Database("iotmonitor"),
		logger: dbLogger,
	}, nil
}
