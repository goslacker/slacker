package trace

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func init() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

type traceType string

const (
	TraceTypeJaeger traceType = "jaeger"
)

type TraceConfig struct {
	Type     traceType
	Endpoint string
	Name     string
	Addr     string
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

	addr := strings.Split(conf.Addr, ":")
	ip := addr[0]
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		return
	}

	names := strings.Split(strings.Trim(conf.Name, "/"), ".")
	res := resource.NewSchemaless(
		semconv.ServiceNameKey.String(names[len(names)-1]),
		semconv.ServiceNamespaceKey.String(strings.Join(names[:len(names)-1], ".")),
		semconv.ServerAddress(ip),
		semconv.ServerPort(port),
	)
	r, err := resource.Merge(resource.Default(), res)
	if err != nil {
		return
	}

	tp = traceSdk.NewTracerProvider(traceSdk.WithBatcher(exporter), traceSdk.WithResource(r))
	return
}
