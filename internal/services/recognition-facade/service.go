// Package recognition_facade is the core logic of recognition-facade service.
//
// It retrieves the generic event (like MovementEvent, which reveals nothing about the entities in the image)
// from the message queue, and calls the corresponding EntityRecognition and SimilarityAnalysis (wip) service to
// get further information of an image or event.
package recognition_facade

import (
	"context"

	"github.com/nkust-monitor-iot-project-2024/central/internal/mq"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"go.uber.org/fx"
)

// FxModule is the fx module for the recognition-facade service.
var FxModule = fx.Module("recognition-facade", fx.Provide(New))

// Service is the core of the service, "recognition-facade".
type Service struct {
	messageQueue mq.MessageQueue

	entityRecognitionClient entityrecognitionpb.EntityRecognitionClient
}

// New creates a new Service.
func New(messageQueue mq.MessageQueue, entityRecognitionClient entityrecognitionpb.EntityRecognitionClient) *Service {
	return &Service{
		messageQueue:            messageQueue,
		entityRecognitionClient: entityRecognitionClient,
	}
}

// Run runs the service and blocks until the service is stopped by ctx or something wrong.
func (s *Service) Run(ctx context.Context) error {
	return nil
}
