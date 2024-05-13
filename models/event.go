package models

import (
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
)

// EventMetadata represents the metadata of the event.
type EventMetadata struct {
	// DeviceID is the unique identifier of the device.
	DeviceID string

	// Timestamp is the timestamp of the event.
	Timestamp time.Time
}

func EventMetadataFromProto(pb *eventpb.EventMetadata) EventMetadata {
	return EventMetadata{
		DeviceID:  pb.GetDeviceId(),
		Timestamp: pb.GetTimestamp().AsTime(),
	}
}
