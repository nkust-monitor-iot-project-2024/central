package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opentelemetry.io/otel/trace"
)

type GraphqlErrorCode struct {
	Code string
	Desc string
}

func (g GraphqlErrorCode) String() string {
	return fmt.Sprintf("%s: %s", g.Code, g.Desc)
}

func NewGraphqlErrorCode(code, desc string) GraphqlErrorCode {
	return GraphqlErrorCode{Code: code, Desc: desc}
}

var (
	GraphqlErrorCodeInternalError = NewGraphqlErrorCode("INTERNAL_ERROR", "Internal error")
	GraphqlErrorCodeInvalidCursor = NewGraphqlErrorCode("INVALID_CURSOR", "Invalid cursor")
)

func NewTraceableErrorWithCode(ctx context.Context, code GraphqlErrorCode, err error) error {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	return &gqlerror.Error{
		Err:     err,
		Message: code.String(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code":      code.Code,
			"requestID": fmt.Sprintf("%s::%s", traceID, spanID),
		},
	}
}
