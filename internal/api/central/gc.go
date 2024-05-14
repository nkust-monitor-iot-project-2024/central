package central

import (
	"context"
	"log/slog"
	"time"

	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"go.mongodb.org/mongo-driver/bson"
)

// gcEventLoop clean up the loop of obsolete events
func (s *service) gcEventLoop(ctx context.Context) {
	gcInterval := s.config.Duration("gc.event.interval")
	if gcInterval == 0 {
		slog.WarnContext(
			ctx,
			"GC event interval is not set. You should set it to clean up garbage.",
		)
		return
	}

	gcPreserveDuration := s.config.Duration("gc.event.preserve_duration")
	if gcPreserveDuration == 0 {
		slog.WarnContext(ctx, "GC event preserve duration is not set. You should set it to clean up garbage.")
		return // no gc enabled
	}

	s.logger.InfoContext(ctx, "first garbage collecting events")
	s.gcEvent(ctx, gcPreserveDuration)

	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-ticker.C:
			s.logger.InfoContext(ctx, "garbage collecting events", slog.Time("tick", tick))

			s.gcEvent(ctx, gcPreserveDuration)
		}
	}
}

func (s *service) gcEvent(ctx context.Context, dur time.Duration) {
	cleanupCreatedAt := time.Now().Add(-dur)

	// Find all events that are older than the cleanupCreatedAt
	events, err := s.db.DeleteEvents(ctx, &database.DeleteEventsRequest{
		Filter: bson.M{
			"createdAt": bson.M{
				"$lt": cleanupCreatedAt,
			},
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to delete events", slogext.Error(err))
		return
	}
	if len(events.DeletedEventID) > 0 {
		s.logger.InfoContext(ctx, "garbage collected events", slog.Int("count", len(events.DeletedEventID)))
		return
	}
}
