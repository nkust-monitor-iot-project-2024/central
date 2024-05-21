package event

// Channel is a generic interface for event channels.
type Channel interface {
	// Subscribe subscribes to the channel.
	Subscribe() <-chan *Event
}
