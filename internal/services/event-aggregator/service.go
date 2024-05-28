// Package event_aggregator is the core logic of the service "event-aggregator".
package event_aggregator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.uber.org/fx"
)

// FxModule is the fx module for the Service that handles the cleanup.
var FxModule = fx.Module(
	"event-aggregator",
	models.EventRepositoryEntFx,
	mq.FxModule,
	fx.Provide(NewStorer),
	fx.Provide(New),
	fx.Invoke(func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, s *Service) error {
		ctx, cancel := context.WithCancel(context.Background())

		lifecycle.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					defer cancel()

					if err := s.Run(ctx); err != nil {
						if !errors.Is(err, context.Canceled) {
							slog.Error("event service stopped with errors", slogext.Error(err))
							_ = shutdowner.Shutdown(fx.ExitCode(1))
						}
					}

					slog.InfoContext(ctx, "event service stopped")
				}()
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})

		return nil
	}),
)

// Service is the service that aggregates the events.
type Service struct {
	repo         models.EntEventRepository
	messageQueue mq.MessageQueue
}

// New creates a new Service.
func New(repo models.EntEventRepository, mq mq.MessageQueue) *Service {
	return &Service{
		repo:         repo,
		messageQueue: mq,
	}
}

// Run creates the sub-service (Storer) and blocks until something wrong or the context is canceled.
func (s *Service) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	storer, err := NewStorer(s)
	if err != nil {
		return fmt.Errorf("initialize storer: %w", err)
	}

	eventChan, err := s.messageQueue.SubscribeEvent(ctx)
	if err != nil {
		return fmt.Errorf("subscribe to event: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		storer.Run(ctx, eventChan)
	}()

	wg.Wait()
	return ctx.Err()
}
