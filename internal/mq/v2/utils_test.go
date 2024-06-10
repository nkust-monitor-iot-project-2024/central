package mqv2_test

import (
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestGetDeliveryCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  amqp091.Table
		want int64
	}{
		{
			name: "no x-delivery-count, expect 0",
			arg:  amqp091.Table{},
			want: 0,
		},
		{
			name: "x-delivery-count is 1 [int64], expect 1",
			arg:  amqp091.Table{"x-delivery-count": int64(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [int], expect 1",
			arg:  amqp091.Table{"x-delivery-count": 1},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [int8], expect 1",
			arg:  amqp091.Table{"x-delivery-count": int8(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [int16], expect 1",
			arg:  amqp091.Table{"x-delivery-count": int16(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [int32], expect 1",
			arg:  amqp091.Table{"x-delivery-count": int32(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [float32], expect 1",
			arg:  amqp091.Table{"x-delivery-count": float32(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is 1 [float64], expect 1",
			arg:  amqp091.Table{"x-delivery-count": float64(1)},
			want: 1,
		},
		{
			name: "x-delivery-count is string, expect -1 (error)",
			arg:  amqp091.Table{"x-delivery-count": "1"},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mqv2.GetDeliveryCount(amqp091.Delivery{
				Headers: tt.arg,
			}); got != tt.want {
				t.Errorf("GetDeliveryCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDeliveryOverRequeueLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  amqp091.Table
		want bool
	}{
		{
			name: "no x-delivery-count, expect false",
			arg:  amqp091.Table{},
			want: false,
		},
		{
			name: "x-delivery-count is 0, expect false",
			arg:  amqp091.Table{"x-delivery-count": int64(0)},
			want: false,
		},
		{
			name: "x-delivery-count is 1, expect false",
			arg:  amqp091.Table{"x-delivery-count": int64(1)},
			want: false,
		},
		{
			name: "x-delivery-count is 3 (actual count = 4), expect true",
			arg:  amqp091.Table{"x-delivery-count": int64(3)},
			want: true,
		},
		{
			name: "x-delivery-count is 4 (actual count = 5), expect true",
			arg:  amqp091.Table{"x-delivery-count": int64(4)},
			want: true,
		},
		{
			name: "x-delivery-count is string, expect true",
			arg:  amqp091.Table{"x-delivery-count": "4"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mqv2.IsDeliveryOverRequeueLimit(amqp091.Delivery{
				Headers: tt.arg,
			}); got != tt.want {
				t.Errorf("IsDeliveryOverRequeueLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
