package interceptor

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var Providers = map[string]*traceSdk.TracerProvider{}

func UnaryTraceServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	names := strings.Split(strings.Trim(info.FullMethod, "/"), "/")

	tp, ok := Providers[names[0]]
	if ok {
		otel.SetTracerProvider(tp)
	} else {
		otel.SetTracerProvider(nil)
	}

	newCtx, span := startServerSpan(ctx, names[1])
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
	var srcNames []string

	if srcMethod, ok := grpc.Method(ctx); ok {
		srcNames = strings.Split(strings.Trim(srcMethod, "/"), "/")
	}

	// targetNames := strings.Split(strings.Trim(method, "/"), "/")
	tp, ok := Providers[srcNames[0]]
	if ok {
		otel.SetTracerProvider(tp)
	} else {
		otel.SetTracerProvider(nil)
	}

	newCtx, span := startClientSpan(ctx, method, srcNames[1])
	defer span.End()
	return invoker(newCtx, method, req, reply, cc, opts...)
}

func startClientSpan(ctx context.Context, method, target string) (context.Context, trace.Span) {
	tr := otel.Tracer("slacker")
	ctx, span := tr.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))

	return ctx, span
}
