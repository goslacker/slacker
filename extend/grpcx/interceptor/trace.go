package interceptor

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryTraceInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	newCtx, span := startSpan(ctx, info.FullMethod)
	defer span.End()
	return handler(newCtx, req)
}

func startSpan(ctx context.Context, name string) (newCtx context.Context, span trace.Span) {
	tr := otel.Tracer("slacker")
	newCtx, span = tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		name,
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return
}
