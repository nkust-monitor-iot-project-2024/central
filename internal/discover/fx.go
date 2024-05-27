package discover

import (
	"context"

	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"go.uber.org/fx"
)

// DiscovererFxModule is the fx module for the discoverer.
//
// You should not use this module directly; use the service-specific module
// (such as EntityRecognitionServiceFxModule) instead.
var DiscovererFxModule = fx.Module("discoverer", fx.Provide(NewDiscoverer))

// EntityRecognitionServiceFxModule is the fx module for the entity recognition service.
var EntityRecognitionServiceFxModule = fx.Module("entity-recognition-service",
	DiscovererFxModule,
	fx.Provide(func(d *Discoverer) (entityrecognitionpb.EntityRecognitionClient, error) {
		return d.DiscoverEntityRecognitionService(context.Background())
	}))
