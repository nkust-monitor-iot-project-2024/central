package utils

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

// NewResourceBuilder creates a new resource builder function to create the resource.Resource,
// useful in fx.Provide.
func NewResourceBuilder(serviceName string, serviceVersion string) func() (*resource.Resource, error) {
	return func() (*resource.Resource, error) {
		return NewResource(serviceName, serviceVersion)
	}
}

// NewResource creates a new resource.Resource with the given service name and version.
//
// If the schema is invalid, it will return an error.
func NewResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.ServiceNamespace("iot-monitor/central"),
			semconv.ServiceInstanceID(uuid.New().String()),
		),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
