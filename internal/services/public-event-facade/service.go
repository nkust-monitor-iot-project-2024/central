// Package public_event_facade provides the core of public event facade service.
//
// It provides the read-only interfaces to external clients to get the event information
// by the GraphQL interface.
// The definition of GraphQL can be got in `graph/schema.graphqls`.
//
// The "Authentication" has not been implemented in this service yet.
// You can implement your authentication in your gateway service.
package public_event_facade

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nkust-monitor-iot-project-2024/central/graph"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	gqlgenopentelemetry "github.com/zhevron/gqlgen-opentelemetry"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

var FxModule = fx.Module("public-event-facade", fx.Provide(New))

// Service is the core of the service, "public-event-facade".
type Service struct {
	config   utils.Config
	resolver *graph.Resolver

	tracer trace.Tracer
}

// New creates a new Service.
func New(config utils.Config, resolver *graph.Resolver) *Service {
	return &Service{
		config:   config,
		resolver: resolver,
	}
}

// Run runs the service.
func (s *Service) Run() error {
	port := s.config.Int("service.publiceventfacade.port")
	if port == 0 {
		port = 8080
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
	srv.Use(gqlgenopentelemetry.Tracer{
		IncludeVariables: true,
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", srv)

	listenOn := fmt.Sprintf(":%d", port)
	slog.Info(
		"graphql playground address",
		slog.String("address", "http://localhost"+listenOn),
		slog.String("listenOn", listenOn))

	log.Fatal(http.ListenAndServe(listenOn, nil))

	return nil
}
