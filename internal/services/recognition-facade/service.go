// Package recognition_facade is the core logic of recognition-facade service.
//
// It retrieves the generic event (like MovementEvent, which reveals nothing about the entities in the image)
// from the message queue, and calls the corresponding EntityRecognition and SimilarityAnalysis (wip) service to
// get further information of an image or event.
package recognition_facade

import (
	"context"
	"fmt"

	mqv2 "github.com/nkust-monitor-iot-project-2024/central/internal/mq/v2"

	"github.com/nkust-monitor-iot-project-2024/central/internal/discover"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

// FxModule is the fx module for the Service that handles the cleanup.
var FxModule = fx.Module(
	"recognition-facade",
	mqv2.FxModule,
	discover.EntityRecognitionServiceFxModule,
	fx.Provide(fx.Annotate(New, fx.As(new(services.Service)))),
	fx.Invoke(services.BootstrapFxService),
)

// Service is the core of the service, "recognition-facade".
type Service struct {
	amqp *mqv2.AmqpWrapper

	entityRecognitionClient entityrecognitionpb.EntityRecognitionClient
}

// New creates a new Service.
func New(amqp *mqv2.AmqpWrapper, entityRecognitionClient entityrecognitionpb.EntityRecognitionClient) *Service {
	return &Service{
		amqp:                    amqp,
		entityRecognitionClient: entityRecognitionClient,
	}
}

// Run runs the service and blocks until the service is stopped by ctx or something wrong.
func (s *Service) Run(ctx context.Context) error {
	recognizer, err := NewRecognizer(s)
	if err != nil {
		return fmt.Errorf("initialize recognizer: %w", err)
	}

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		recognizer.Run(ctx)
		return nil
	})

	return group.Wait()
}
