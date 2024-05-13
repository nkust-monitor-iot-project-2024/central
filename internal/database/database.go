package database

import (
	"context"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection interface {
	Events() *mongo.Collection
	Fs() *gridfs.Bucket
}

type wrapper struct {
	db *mongo.Database
}

func (d *wrapper) Events() *mongo.Collection {
	return d.db.Collection("events")
}

func (d *wrapper) Fs() *gridfs.Bucket {
	bucket, _ := gridfs.NewBucket(d.db)
	return bucket
}

func ConnectByConfig(conf utils.Config) (Collection, error) {
	uri := conf.String("mongo.uri")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	return &wrapper{db: client.Database("iotmonitor")}, nil
}
