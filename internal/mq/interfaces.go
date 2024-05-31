package mq

import "errors"

type SubscribeResponse[T any] struct {
	// DeliveryChan is the channel for the delivery.
	DeliveryChan <-chan T

	// Cleanup is the cleanup function that can used to clean up the subscription.
	//
	// If the error happened, you should also call this Cleanup function.
	Cleanup func() error

	// ClosedChan is the channel that will send the error (or nil if the channel
	// is closed normally) if the subscription is closed.
	ClosedChan <-chan error
}

// ErrNotRejectable is the error indicating the delivery is not rejectable.
var ErrNotRejectable = errors.New("the delivery is not rejectable")

// Rejectable provides an interface, allow the consumers rejecting the message.
type Rejectable interface {
	// Reject rejects the message.
	//
	// If requeue is true, the message will be requeued.
	Reject(requeue bool) error
}

// Reject is a helper function to reject the message.
func Reject(delivery any, requeue bool) error {
	if rejectableDelivery, ok := delivery.(Rejectable); ok {
		return rejectableDelivery.Reject(requeue)
	}

	return ErrNotRejectable
}

// ErrNotAckable is the error indicating the delivery is not ackable.
var ErrNotAckable = errors.New("the delivery is not ackable")

// Ackable provides an interface, allow the consumers acknowledging the message.
type Ackable interface {
	// Ack acknowledges the message.
	Ack() error
}

// Ack is a helper function to acknowledge the message.
func Ack(delivery any) error {
	if ackableDelivery, ok := delivery.(Ackable); ok {
		return ackableDelivery.Ack()
	}

	return ErrNotAckable
}
