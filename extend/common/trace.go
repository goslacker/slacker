package common

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, name string) (newCtx context.Context, span trace.Span) {
	tr := otel.Tracer("slacker")
	newCtx, span = tr.Start(ctx, name)
	return
}
