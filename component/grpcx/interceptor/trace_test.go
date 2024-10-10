package interceptor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
)

func TestTrace(t *testing.T) {
	t.Skip()
	stdoutExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	require.NoError(t, err)

	trExporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpointURL("http://127.0.0.1:4317"))
	require.NoError(t, err)

	r, _ := resource.Merge(resource.Default(), resource.NewSchemaless(semconv.ServiceNameKey.String("testService")))
	tp := trace.NewTracerProvider(trace.WithBatcher(trExporter), trace.WithBatcher(stdoutExporter), trace.WithResource(r))
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	UnaryTraceServerInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "1/1"}, func(c context.Context, req interface{}) (any, error) {
		ctx = c
		time.Sleep(2 * time.Second)
		return nil, nil
	})
	UnaryTraceServerInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "1/2"}, func(ctx context.Context, req interface{}) (any, error) {
		time.Sleep(1 * time.Second)
		return nil, nil
	})
}

func TestDefault(t *testing.T) {
	res := resource.Default()
	require.False(t, res.Equal(resource.Empty()))
	require.True(t, res.Set().HasValue(semconv.ServiceNameKey))

	serviceName, _ := res.Set().Value(semconv.ServiceNameKey)
	require.True(t, strings.HasPrefix(serviceName.AsString(), "unknown_service:"))
	require.Greaterf(t, len(serviceName.AsString()), len("unknown_service:"), "default service.name should include executable name")

	require.Contains(t, res.Attributes(), semconv.TelemetrySDKLanguageGo)
	require.Contains(t, res.Attributes(), semconv.TelemetrySDKVersion(sdk.Version()))
	require.Contains(t, res.Attributes(), semconv.TelemetrySDKName("opentelemetry"))
}
