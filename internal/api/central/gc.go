package central

import (
	"context"
	"log/slog"
	"time"
)

// gcEventLoop clean up the loop of obsolete events
func (s *service) gcEventLoop(ctx context.Context) {
	gcInterval := s.config.Duration("gc.event.interval")
	if gcInterval == 0 {
		slog.WarnContext(
			ctx,
			"GC event interval is not set. You should set it to prevent garbage.",
		)
	}

	tick := time.NewTicker(gcInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			// list all
		}
	}
}

func (s *service) gcEvent(ctx context.Context, interval time.Duration) {
	s.logger.DebugContext(ctx, "garbage collecting events")

	cleanupCreatedAt := time.Now().Add(-interval)

	// Mark all events that are older than the cleanupCreatedAt

}
