package slogext

import (
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
)

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
