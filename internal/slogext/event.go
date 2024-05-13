package slogext

import (
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func ObjectID(key string, id primitive.ObjectID) slog.Attr {
	return slog.String(key, id.Hex())
}

func EventMetadata(metadata models.EventMetadata) slog.Attr {
	return slog.Group(
		"event_metadata",
		slog.String("device_id", metadata.DeviceID),
		slog.Time("timestamp", metadata.Timestamp),
	)
}

func EventMetadataPb(metadatapb *eventpb.EventMetadata) slog.Attr {
	return EventMetadata(models.EventMetadataFromProto(metadatapb))
}
