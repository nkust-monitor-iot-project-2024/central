package mqv2

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

// QueueDeclareArgs is the arguments for declaring a queue.
var QueueDeclareArgs = amqp091.Table{
	"x-queue-type": "quorum",
}

// GetDeliveryCount gets the delivery count of the message.
//
// -1 means the message is invalid.
func GetDeliveryCount(delivery amqp091.Delivery) int64 {
	// If the message is resent over 5 times, we throw it.
	if v, ok := delivery.Headers["x-delivery-count"]; ok {
		var deliveryCount int64

		switch value := v.(type) {
		case int:
			deliveryCount = int64(value)
		case int8:
			deliveryCount = int64(value)
		case int16:
			deliveryCount = int64(value)
		case int32:
			deliveryCount = int64(value)
		case int64:
			deliveryCount = value
		case float32:
			deliveryCount = int64(value)
		case float64:
			deliveryCount = int64(value)
		default:
			slog.Warn("invalid x-delivery-count",
				slog.String("type", fmt.Sprintf("%T", v)),
				slog.String("value", fmt.Sprintf("%v", v)))

			return -1
		}

		return deliveryCount
	}

	return 0
}

// IsDeliveryOverRequeueLimit checks if the delivery should be throws.
//
// The delivery should be throws if the message is resent over three times,
// or the message is invalid.
func IsDeliveryOverRequeueLimit(delivery amqp091.Delivery) bool {
	count := GetDeliveryCount(delivery)
	return count >= 3 || count == -1
}
