package event

import (
	"context"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"go.uber.org/fx"
)

var FxModule = fx.Module(
	"services/event",
	fx.Provide(NewStorer),
	fx.Provide(New),
)

type Service struct {
	client *ent.Client
	storer *Storer
}

func New(client *ent.Client, storer *Storer) *Service {
	return &Service{
		client: client,
		storer: storer,
	}
}

func (s *Service) Run() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.storer.Run(context.Background(), nil /* wip */)
	}()

	wg.Wait()
}
