package mqevent

import (
	"testing"

	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/samber/mo"
)

func TestGetMessageKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  mo.Option[models.EventType]
		want string
	}{
		{
			name: "empty",
			arg:  mo.None[models.EventType](),
			want: "event.v1.*",
		},
		{
			name: "movement",
			arg:  mo.Some[models.EventType](models.EventTypeMovement),
			want: "event.v1.movement",
		},
		{
			name: "invaded",
			arg:  mo.Some[models.EventType](models.EventTypeInvaded),
			want: "event.v1.invaded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRoutingKey(tt.arg); got != tt.want {
				t.Errorf("GetRoutingKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
