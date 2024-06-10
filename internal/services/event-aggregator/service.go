// Package event_aggregator is the core logic of the service "event-aggregator".
package event_aggregator

import (
	"context"
	"fmt"
	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"

	"github.com/nkust-monitor-iot-project-2024/central/internal/services"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

// FxModule is the fx module for the Service that handles the cleanup.
var FxModule = fx.Module(
	"event-aggregator",
	models.EventRepositoryEntFx,
	mqv2.FxModule,
	fx.Provide(fx.Annotate(New, fx.As(new(services.Service)))),
	fx.Invoke(services.BootstrapFxService),
)

// Service is the service that aggregates the events.
type Service struct {
	repo models.EntEventRepository
	amqp *mqv2.AmqpWrapper
}

// New creates a new Service.
func New(repo models.EntEventRepository, amqp *mqv2.AmqpWrapper) *Service {
	return &Service{
		repo: repo,
		amqp: amqp,
	}
}

// Run creates the sub-service (Storer) and blocks until something wrong or the context is canceled.
func (s *Service) Run(ctx context.Context) error {
	storer, err := NewStorer(s)
	if err != nil {
		return fmt.Errorf("initialize storer: %w", err)
	}

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		storer.Run(ctx)
		return nil
	})

	return group.Wait()
}
