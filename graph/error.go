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
)

func NewTraceableErrorWithCode(ctx context.Context, code GraphqlErrorCode, err error) error {
	spanID := trace.SpanFromContext(ctx).SpanContext().SpanID().String()

	return &gqlerror.Error{
		Err:     err,
		Message: code.String(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]any{
			"code":      code.Code,
			"requestID": spanID,
		},
	}
}
