package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/fx"
)

type Service interface {
	Run(ctx context.Context) error
}

type BootstrapFxServiceParam struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
	Resource   *resource.Resource
	Service    Service
}

// BootstrapFxService is the bootstrap function that bootstraps a service
// that implements the Service interface.
//
// It will run the service, then cancel the context and shutdown
// the other Fx services when the service is stopped.
//
// This bootstrap should be only used in the Fx context.
// You should provide the resource of this service (for identifying),
// and the service that you want to run.
func BootstrapFxService(param BootstrapFxServiceParam) error {
	ctx, cancel := context.WithCancel(context.Background())

	param.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				defer cancel()

				if err := param.Service.Run(ctx); err != nil {
					if !errors.Is(err, context.Canceled) {
						slog.Error("service stopped with errors",
							slog.String("resource", param.Resource.String()),
							slogext.Error(err))
						_ = param.Shutdowner.Shutdown(fx.ExitCode(1))
					}
				}

				slog.InfoContext(ctx, "service stopped",
					slog.String("resource", param.Resource.String()))
				_ = param.Shutdowner.Shutdown(fx.ExitCode(0))
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	return nil
}
