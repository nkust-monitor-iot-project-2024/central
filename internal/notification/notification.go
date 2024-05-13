package notification

// NotificationCode represents the notification code.
//
// You can freely define the notification code as you like.
// Make sure the receiver of the notification knows how to handle it.
type NotificationCode string

// Receiver represents the receiver of the notification.
type Receiver struct {
	// ID is the unique identifier of the receiver.
	ID string
}

// Notification represents the notification.
type Notification struct {
	// Code is the notification code.
	Code NotificationCode
}

// Notifier represents the notifier.
type Notifier interface {
	// Notify sends the notification to the receiver.
	Notify(receiver Receiver, notification Notification) error
}
