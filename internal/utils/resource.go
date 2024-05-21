package utils

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func MustNewResource(serviceName string, serviceVersion string) *resource.Resource {
	res, err := NewResource(serviceName, serviceVersion)
	if err != nil {
		panic(err)
	}

	return res
}

func NewResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
