package interceptor

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	propagator := otel.GetTextMapPropagator()
	rscc := propagator.Extract(ctx, metadataTextMapCarrier(md))

	rsc := trace.SpanContextFromContext(rscc)
	tr := otel.Tracer("slacker")
	newCtx, span = tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, rsc),
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

	tp, ok := Providers[srcNames[0]]
	if ok {
		otel.SetTracerProvider(tp)
	} else {
		otel.SetTracerProvider(nil)
	}

	targetNames := strings.Split(strings.Trim(method, "/"), "/")
	newCtx, span := startClientSpan(ctx, targetNames[0])
	defer span.End()

	return invoker(newCtx, method, req, reply, cc, opts...)
}

func startClientSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tr := otel.Tracer("slacker")
	ctx, span := tr.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	otel.GetTextMapPropagator().Inject(ctx, metadataTextMapCarrier(md))

	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx, span
}

type metadataTextMapCarrier metadata.MD

// Get returns the value associated with the passed key.
func (m metadataTextMapCarrier) Get(key string) string {
	g, ok := m[key]
	if !ok || len(g) == 0 {
		return ""
	}
	return g[0]
}

// Set stores the key-value pair.
func (m metadataTextMapCarrier) Set(key string, value string) {
	m[key] = []string{value}
}

// Keys lists the keys stored in this carrier.
func (m metadataTextMapCarrier) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
