// Package event_aggregator is the core logic of the service "event-aggregator".
package event_aggregator

import (
	"context"
	"fmt"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.uber.org/fx"
)

// FxModule is the fx module for the Service that handles the cleanup.
var FxModule = fx.Module(
	"event-aggregator",
	models.EventRepositoryEntFx,
	mq.FxModule,
	fx.Provide(fx.Annotate(New, fx.As(new(services.Service)))),
	fx.Invoke(services.BootstrapFxService),
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		storer.Run(ctx)
	}()

	wg.Wait()
	return ctx.Err()
}
