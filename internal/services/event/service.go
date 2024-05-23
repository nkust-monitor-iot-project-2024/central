package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/fx"
)

var FxModule = fx.Module(
	"services/event",
	fx.Provide(NewStorer),
	fx.Provide(New),
	fx.Invoke(func(lifecycle fx.Lifecycle, s *Service) error {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				tracer := otel.GetTracerProvider().Tracer("services/event")

				go func() {
					ctx, span := tracer.Start(ctx, "event_service/on_start")
					defer span.End()

					if err := s.Run(ctx); err != nil {
						span.SetStatus(codes.Error, "event service stopped with errors")
						span.RecordError(err)
					}

					span.SetStatus(codes.Ok, "event service stopped")
				}()
				return nil
			},
		})
		return nil
	}),
)

type Service struct {
	client       *ent.Client
	messageQueue mq.MessageQueue
}

func New(client *ent.Client, mq mq.MessageQueue) *Service {
	return &Service{
		client:       client,
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
