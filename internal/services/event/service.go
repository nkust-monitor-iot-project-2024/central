package event

import (
	"context"
	"sync"

	"github.com/nkust-monitor-iot-project-2024/central/ent"
)

type Service struct {
	client *ent.Client
}

func New(client *ent.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Run() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.storeEventsTask(context.Background(), nil /* wip */)
	}()

	wg.Wait()
}
