package graph

import (
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
)

// GqlParameterToPagination converts the GraphQL parameters to a pagination object.
//
// If there is after cursor, it will return a forward pagination object
// Then, if there is before cursor, it will return a backward pagination object.
// If both cursors are nil, it will return a default forward pagination object with the specified first.
func GqlParameterToPagination(first *int, before *string, last *int, after *string) models.Pagination {
	if after != nil {
		return models.ForwardPagination{
			First: lo.FromPtrOr(first, 0),
			After: *after,
		}
	}

	if before != nil {
		return models.BackwardPagination{
			Last:   lo.FromPtrOr(last, 0),
			Before: *before,
		}
	}

	return models.ForwardPagination{
		First: lo.FromPtrOr(first, 0),
	}
}

// PaginationToTraceAttributes converts the pagination object to a map of trace attributes.
func PaginationToTraceAttributes(pagination models.Pagination) []attribute.KeyValue {
	switch pagination := pagination.(type) {
	case models.ForwardPagination:
		return []attribute.KeyValue{
			attribute.String("strategy", "forward"),
			attribute.Int("first", pagination.First),
			attribute.String("after", pagination.After),
		}
	case models.BackwardPagination:
		return []attribute.KeyValue{
			attribute.String("strategy", "backward"),
			attribute.Int("last", pagination.Last),
			attribute.String("before", pagination.Before),
		}
	default:
		return nil
	}
}
