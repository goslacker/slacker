package interceptor

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

	span.SetAttributes(semconv.RPCServiceKey.String(names[0]))
	span.SetAttributes(semconv.RPCMethodKey.String(names[1]))
	result, err = handler(newCtx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))
			span.SetAttributes(semconv.RPCConnectRPCErrorCodeKey.String(s.Code().String()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.Ok)))
	span.SetAttributes(semconv.RPCConnectRPCErrorCodeKey.String(codes.Ok.String()))
	return
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

func UnaryTraceClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
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

	newCtx, span := startClientSpan(ctx, "call "+method)
	defer span.End()

	targetNames := strings.Split(strings.Trim(method, "/"), "/")
	span.SetAttributes(semconv.RPCServiceKey.String(targetNames[0]))
	span.SetAttributes(semconv.RPCMethodKey.String(targetNames[1]))

	err = invoker(newCtx, method, req, reply, cc, opts...)

	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))
			span.SetAttributes(semconv.RPCConnectRPCErrorCodeKey.String(s.Code().String()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(codes.Ok)))
	span.SetAttributes(semconv.RPCConnectRPCErrorCodeKey.String(codes.Ok.String()))
	return
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
