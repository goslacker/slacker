package trace

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TraceConfig struct {
	Type     TraceType
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

func BuildTraceProviders(typ TraceType, endpoint string, serviceNames []string, addr string) (providers map[string]*traceSdk.TracerProvider, deferFunc func()) {
	deferFunc = func() {}
	if typ == "" || endpoint == "" || len(serviceNames) == 0 || addr == "" {
		return
	}

	providers = make(map[string]*traceSdk.TracerProvider, len(serviceNames))
	for _, name := range serviceNames {
		if strings.Contains(name, "grpc") {
			continue
		}
		var err error
		conf := &TraceConfig{
			Type:     typ,
			Endpoint: endpoint,
			Name:     name,
			Addr:     addr,
		}
		providers[name], err = NewTraceProvider(conf)
		if err != nil {
			panic(fmt.Errorf("create trace provider failed: %w", err))
		}
	}

	return providers, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, tp := range providers {
			tp.Shutdown(ctx)
		}
	}
}

var providers map[string]*traceSdk.TracerProvider

func InitTraceProviders(typ TraceType, endpoint string, serviceNames []string, addr string) (deferFunc func(), err error) {
	if len(providers) > 0 {
		err = fmt.Errorf("trace providers already initialized")
		return
	}
	providers, deferFunc = BuildTraceProviders(typ, endpoint, serviceNames, addr)
	return
}
