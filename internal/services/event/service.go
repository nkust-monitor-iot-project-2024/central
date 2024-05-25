package event

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.uber.org/fx"
)

var FxModule = fx.Module(
	"event-aggregator",
	fx.Provide(NewStorer),
	fx.Provide(New),
	fx.Invoke(func(lifecycle fx.Lifecycle, s *Service) error {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := s.Run(ctx); err != nil {
						slog.ErrorContext(ctx, "event service stopped with errors", slogext.Error(err))
					}
				}()
				return nil
			},
		})
		return nil
	}),
)

type Service struct {
	repo         models.EntEventRepository
	messageQueue mq.MessageQueue
}

func New(repo models.EntEventRepository, mq mq.MessageQueue) *Service {
	return &Service{
		repo:         repo,
		messageQueue: mq,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	return nil
}
