package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.mongodb.org/mongo-driver/bson"
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

	DeleteEvents(context.Context, *DeleteEventsRequest) (*DeleteEventsResponse, error)
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

func ConnectAndMigrateByConfig(conf utils.Config, logger *slog.Logger) (Database, error) {
	dbLogger := logger.With(slog.String("service", "_database"))

	uri := conf.String("mongo.uri")
	if uri == "" {
		return nil, errors.New("missing mongo URI (mongo.uri)")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	db := client.Database("iotmonitor")

	// update JSON schema of the events collection
	err = func() error {
		collsJsonSchema := map[string]any{
			"events": models.EventJsonSchema,
		}
		existingCollections, err := db.ListCollectionNames(context.Background(), bson.M{})
		if err != nil {
			return fmt.Errorf("list collections: %w", err)
		}
		for coll, jsonSchema := range collsJsonSchema {
			if slices.Contains(existingCollections, coll) {
				// collMod
				result := db.RunCommand(
					context.Background(),
					bson.D{
						{Key: "collMod", Value: coll},
						{Key: "validator", Value: bson.M{"$jsonSchema": jsonSchema}},
					},
				)
				if result.Err() != nil {
					return fmt.Errorf("update JSON schema: %w", result.Err())
				}
			} else {
				err := db.CreateCollection(
					context.Background(), coll,
					options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema}),
				)
				if err != nil {
					return fmt.Errorf("create collection: %w", err)
				}
			}
		}

		return nil
	}()
	if err != nil {
		logger.Warn("failed to migrate database", slogext.Error(err))
	}

	return &database{
		db:     client.Database("iotmonitor"),
		logger: dbLogger,
	}, nil
}
