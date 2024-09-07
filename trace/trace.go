package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type traceType string

const (
	TraceTypeJaeger traceType = "jaeger"
)

type TraceConfig struct {
	Type     traceType
	Endpoint string
	Name     string
}

func NewTraceProvider(conf *TraceConfig) (tp *traceSdk.TracerProvider, err error) {
	var exporter *otlptrace.Exporter

	switch conf.Type {
	case TraceTypeJaeger:
		exporter, err = otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpointURL(conf.Endpoint))
	default:
		err = fmt.Errorf("unsupported trace type %s", conf.Type)
	}
	if err != nil {
		return
	}

	r, err := resource.Merge(resource.Default(), resource.NewSchemaless(semconv.ServiceNameKey.String(conf.Name)))
	if err != nil {
		return
	}

	tp = traceSdk.NewTracerProvider(traceSdk.WithBatcher(exporter), traceSdk.WithResource(r))
	return
}
