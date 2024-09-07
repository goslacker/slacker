package grpcx

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/extend/grpcx/interceptor"
	"github.com/goslacker/slacker/serviceregistry"
	"github.com/goslacker/slacker/serviceregistry/registry"
	"github.com/goslacker/slacker/trace"
	"github.com/spf13/viper"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func WithMiddlewares(middlewares ...grpc.UnaryServerInterceptor) func(*Component) {
	return func(c *Component) {
		c.middlewares = middlewares
	}
}

func WithRegisters(registers ...func(grpc.ServiceRegistrar)) func(*Component) {
	return func(m *Component) {
		m.registers = registers
	}
}

func NewComponent(opts ...func(*Component)) *Component {
	m := &Component{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Component struct {
	app.Component
	grpcServer  *grpc.Server
	middlewares []grpc.UnaryServerInterceptor
	registers   []func(grpc.ServiceRegistrar)
}

func (c *Component) Init() error {
	c.middlewares = append(c.middlewares, interceptor.UnaryValidateInterceptor)
	return app.Bind[*Component](c)
}

func (m *Component) Register(registers ...func(grpc.ServiceRegistrar)) {
	m.registers = append(m.registers, registers...)
}

func (m *Component) Start() {
	var conf Config
	err := viper.Sub("grpc").Unmarshal(&conf)
	if err != nil {
		panic(fmt.Errorf("read config failed: %w", err))
	}

	if conf.Trace != nil {
		m.middlewares = append(m.middlewares, interceptor.UnaryTraceServerInterceptor)
	}

	m.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(m.middlewares...),
	)

	if conf.HealthCheck {
		healthCheck := health.NewServer()
		healthgrpc.RegisterHealthServer(m.grpcServer, healthCheck)
		err := app.Bind[*health.Server](healthCheck)
		if err != nil {
			slog.Error("bind health check failed", "error", err)
			return
		}
	}

	if conf.Reflection {
		reflection.Register(m.grpcServer)
	}

	if len(m.registers) <= 0 {
		slog.Warn("no grpc service registered")
		return
	}

	for _, register := range m.registers {
		register(m.grpcServer)
	}

	if conf.Trace != nil {
		var deferFunc func()
		interceptor.Providers, deferFunc = traceAgent(conf.Trace, m.grpcServer)
		defer deferFunc()
	}

	defer registerService(conf.Registry, conf.Addr, m.grpcServer)()

	var lis net.Listener
	lis, err = net.Listen("tcp", conf.Addr)
	if err != nil {
		panic(fmt.Errorf("tcp listen port failed: %w", err))
	}

	slog.Info("Serving gRPC on " + conf.Addr)
	err = m.grpcServer.Serve(lis)
	if err != nil {
		slog.Error("grpc server shutdown", "error", err)
	} else {
		slog.Info("grpc server shutdown")
	}
}

func registerService(config *registry.RegistryConfig, addr string, svr *grpc.Server) (deRegister func()) {
	deRegister = func() {}
	if config == nil {
		return
	}

	config.Addr = addr
	r, err := serviceregistry.New(config)
	if err != nil {
		panic(fmt.Errorf("create service registry failed: %w", err))
	}

	for name := range svr.GetServiceInfo() {
		//过滤掉反射服务
		if strings.Contains(name, "grpc") {
			continue
		}

		err := r.Register(name)
		if err != nil {
			panic(fmt.Errorf("register service<%s> to registry failed: %w", name, err))
		}
	}

	err = app.Bind[registry.ServiceRegistry](r)
	if err != nil {
		panic(fmt.Errorf("bind service registry failed: %w", err))
	}

	deRegister = func() {
		err = r.Deregister()
		if err != nil {
			slog.Warn("deregister service failed", "error", err)
		}
	}

	return
}

func traceAgent(conf *trace.TraceConfig, svr *grpc.Server) (providers map[string]*traceSdk.TracerProvider, deferFunc func()) {
	deferFunc = func() {}
	if conf == nil {
		return
	}

	svrMap := svr.GetServiceInfo()

	providers = make(map[string]*traceSdk.TracerProvider, len(svrMap))
	for name := range svrMap {
		if strings.Contains(name, "grpc") {
			continue
		}
		var err error
		conf.Name = name
		providers[name], err = trace.NewTraceProvider(conf)
		if err != nil {
			panic(fmt.Errorf("create trace provider failed: %w", err))
		}
	}

	return providers, func() {
		for _, tp := range providers {
			tp.Shutdown(context.Background())
		}
	}
}

func (c *Component) Stop() {
	c.grpcServer.Stop()
}
