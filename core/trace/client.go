package trace

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"
)

func ClientEndTrace(ctx context.Context, dstMethodFullName string, carrier propagation.TextMapCarrier, provider *traceSdk.TracerProvider, f func(ctx context.Context) error) (err error) {
	otel.SetTracerProvider(provider)

	newCtx, span := startClientSpan(ctx, carrier, dstMethodFullName)
	defer span.End()

	targetNames := strings.Split(strings.Trim(dstMethodFullName, "/"), "/")
	span.SetAttributes(semconv.RPCServiceKey.String(targetNames[0]))
	span.SetAttributes(semconv.RPCMethodKey.String(targetNames[1]))

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

func startClientSpan(ctx context.Context, carrier propagation.TextMapCarrier, methodName string) (context.Context, trace.Span) {
	tr := otel.Tracer("slacker")
	ctx, span := tr.Start(ctx, "call "+methodName, trace.WithSpanKind(trace.SpanKindClient))

	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return ctx, span
}
