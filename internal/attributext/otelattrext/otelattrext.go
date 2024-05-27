// Package otelattrext provides additional attribute.KeyValue constructors.
package otelattrext

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// UUID returns a new attribute.KeyValue with the given key and uuid (represented as String).
func UUID(key string, uuid uuid.UUID) attribute.KeyValue {
	return attribute.Key(key).String(uuid.String())
}
