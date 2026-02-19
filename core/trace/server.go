package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"
)

func ServerEndTrace(ctx context.Context, serviceName string, methodName string, carrier propagation.TextMapCarrier, provider *traceSdk.TracerProvider, f func(ctx context.Context) error) (err error) {
	otel.SetTracerProvider(provider)

	newCtx, span := startServerSpan(ctx, carrier, serviceName)
	defer span.End()

	span.SetAttributes(semconv.RPCServiceKey.String(serviceName))
	span.SetAttributes(semconv.RPCMethodKey.String(methodName))

	err = f(newCtx)

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

func startServerSpan(ctx context.Context, carrier propagation.TextMapCarrier, name string) (newCtx context.Context, span trace.Span) {
	propagator := otel.GetTextMapPropagator()
	rscc := propagator.Extract(ctx, carrier)

	rsc := trace.SpanContextFromContext(rscc)
	tr := otel.Tracer("slacker")
	newCtx, span = tr.Start(
		trace.ContextWithRemoteSpanContext(ctx, rsc),
		name,
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return
}
