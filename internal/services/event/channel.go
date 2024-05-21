package event

import "github.com/nkust-monitor-iot-project-2024/central/models"

// Channel is a generic interface for event channels.
type Channel interface {
	// Subscribe subscribes to the event channel
	// and returns a channel of all the non-ACKed events.
	Subscribe() <-chan *models.Event
}
