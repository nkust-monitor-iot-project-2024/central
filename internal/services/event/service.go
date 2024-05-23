package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"go.uber.org/fx"
)

var FxModule = fx.Module(
	"services/event",
	fx.Provide(NewStorer),
	fx.Provide(New),
	fx.Invoke(func(s *Service) error {
		return s.Run()
	}),
)

type Service struct {
	client *ent.Client
}

func New(client *ent.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Run() error {
	wg := sync.WaitGroup{}
	storer, err := NewStorer(s)
	if err != nil {
		return fmt.Errorf("initialize storer: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		storer.Run(context.Background(), nil /* wip */)
	}()

	wg.Wait()
	return nil
}
