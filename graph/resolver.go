package graph

import (
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	eventRepo models.EventRepository

	tracer trace.Tracer
}

// ResolverFxModule is the Fx module for the GraphQL resolver.
var ResolverFxModule = fx.Module("graphql-resolver-module",
	models.EventRepositoryEntFx,
	fx.Provide(NewResolver))

// NewResolver creates a new Resolver with the given dependencies.
func NewResolver(eventRepo models.EventRepository, tracer trace.Tracer) *Resolver {
	return &Resolver{
		eventRepo: eventRepo,
		tracer:    tracer,
	}
}
