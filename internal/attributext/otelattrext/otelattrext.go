package otelattrext

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

func UUID(key string, uuid uuid.UUID) attribute.KeyValue {
	return attribute.Key(key).String(uuid.String())
}
