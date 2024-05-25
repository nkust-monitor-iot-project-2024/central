package utils

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func NewResourceBuilder(serviceName string, serviceVersion string) func() (*resource.Resource, error) {
	return func() (*resource.Resource, error) {
		return NewResource(serviceName, serviceVersion)
	}
}

func NewResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.ServiceInstanceID(uuid.New().String()),
		),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
