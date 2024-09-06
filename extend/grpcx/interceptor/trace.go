package interceptor

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryTraceServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	newCtx, span := startServerSpan(ctx, info.FullMethod)
	defer span.End()
	return handler(newCtx, req)
}

func startServerSpan(ctx context.Context, name string) (newCtx context.Context, span trace.Span) {
	tr := otel.Tracer("slacker")
	newCtx, span = tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		name,
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return
}

func UnaryTraceClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	newCtx, span := startClientSpan(ctx, method, cc.Target())
	defer span.End()
	return invoker(newCtx, method, req, reply, cc, opts...)
}

func startClientSpan(ctx context.Context, method, target string) (context.Context, trace.Span) {
	tr := otel.Tracer("slacker")
	ctx, span := tr.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))

	return ctx, span
}
