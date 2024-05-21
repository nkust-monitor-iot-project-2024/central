package event

import (
	"github.com/nkust-monitor-iot-project-2024/central/protos/eventpb"
)

// Channel is a generic interface for event channels.
type Channel interface {
	// Subscribe subscribes to the event channel
	// and returns a channel of all the non-ACKed events.
	Subscribe() <-chan *eventpb.MovementEventMessage
}
