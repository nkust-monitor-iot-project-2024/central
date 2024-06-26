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
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nkust-monitor-iot-project-2024/central/graph"
	"github.com/nkust-monitor-iot-project-2024/central/internal/services"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	gqlgenopentelemetry "github.com/zhevron/gqlgen-opentelemetry/v2"
	"go.uber.org/fx"
)

var FxModule = fx.Module("public-event-facade",
	graph.ResolverFxModule,
	fx.Provide(New),
	fx.Provide(fx.Annotate(New, fx.As(new(services.Service)))),
	fx.Invoke(services.BootstrapFxService),
)

// Service is the core of the service, "public-event-facade".
type Service struct {
	config   utils.Config
	resolver *graph.Resolver
}

// New creates a new Service.
func New(config utils.Config, resolver *graph.Resolver) *Service {
	return &Service{
		config:   config,
		resolver: resolver,
	}
}

// Run runs the service.
func (s *Service) Run(_ context.Context) error {
	port := s.config.Int("service.publiceventfacade.port")
	if port == 0 {
		port = 8080
	}

	certFile := s.config.String("service.publiceventfacade.tls.cert_file")
	keyFile := s.config.String("service.publiceventfacade.tls.key_file")

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: s.resolver}))
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

	if certFile != "" && keyFile != "" {
		slog.Info("start listening with TLS", slog.String("certFile", certFile), slog.String("keyFile", keyFile))
		return http.ListenAndServeTLS(listenOn, certFile, keyFile, nil)
	}

	slog.Info("start listening without TLS", slog.String("listenOn", listenOn))
	return http.ListenAndServe(listenOn, nil)
}
