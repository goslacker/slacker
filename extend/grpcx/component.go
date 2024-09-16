package grpcx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/extend/grpcx/interceptor"
	"github.com/goslacker/slacker/serviceregistry"
	"github.com/goslacker/slacker/serviceregistry/registry"
	"github.com/goslacker/slacker/tool"
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
		c.unaryServerInterceptors = middlewares
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
	grpcServer               *grpc.Server
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
	registers                []func(grpc.ServiceRegistrar)
}

func (c *Component) Init() (err error) {
	auth := interceptor.NewJWTAuth()
	c.unaryServerInterceptors = []grpc.UnaryServerInterceptor{
		auth.UnaryAuthInterceptor,
		interceptor.UnaryValidateInterceptor,
	}

	c.streamServerInterceptors = []grpc.StreamServerInterceptor{
		auth.StreamAuthInterceptor,
		interceptor.StreamValidateInterceptor,
	}

	err = app.Bind[*interceptor.JWTAuth](auth)
	if err != nil {
		return
	}
	return app.Bind[*Component](c)
}

func (c *Component) RegisterUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	c.unaryServerInterceptors = append(c.unaryServerInterceptors, interceptors...)
}

func (c *Component) RegisterStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	c.streamServerInterceptors = append(c.streamServerInterceptors, interceptors...)
}

func (c *Component) Register(registers ...func(grpc.ServiceRegistrar)) {
	c.registers = append(c.registers, registers...)
}

func (c *Component) Start() {
	if len(c.registers) <= 0 {
		panic(errors.New("no grpc service registered"))
	}

	var conf Config
	err := viper.Sub("grpc").Unmarshal(&conf)
	if err != nil {
		panic(fmt.Errorf("read config failed: %w", err))
	}

	addr, err := c.detectAddr(conf.Addr)
	if err != nil {
		panic(fmt.Errorf("get local ip failed: %w", err))
	}

	if conf.Trace != nil {
		c.unaryServerInterceptors = append([]grpc.UnaryServerInterceptor{interceptor.UnaryTraceServerInterceptor}, c.unaryServerInterceptors...)
		c.streamServerInterceptors = append([]grpc.StreamServerInterceptor{interceptor.StreamTraceServerInterceptor}, c.streamServerInterceptors...)
	}

	c.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(c.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(c.streamServerInterceptors...),
	)

	for _, register := range c.registers {
		register(c.grpcServer)
	}

	if conf.HealthCheck {
		healthCheck := health.NewServer()
		healthgrpc.RegisterHealthServer(c.grpcServer, healthCheck)
		err := app.Bind[*health.Server](healthCheck)
		if err != nil {
			slog.Error("bind health check failed", "error", err)
			return
		}
	}

	if conf.Reflection {
		reflection.Register(c.grpcServer)
	}

	if conf.Trace != nil {
		var deferFunc func()
		interceptor.Providers, deferFunc = traceAgent(conf.Trace, c.grpcServer, addr)
		defer deferFunc()
	}

	if conf.Registry != nil {
		defer registerService(conf.Registry, addr, c.grpcServer)()
	}

	var lis net.Listener
	lis, err = net.Listen("tcp", conf.Addr)
	if err != nil {
		panic(fmt.Errorf("tcp listen port failed: %w", err))
	}

	slog.Info("Serving gRPC on " + conf.Addr)
	err = c.grpcServer.Serve(lis)
	if err != nil {
		slog.Error("grpc server shutdown", "error", err)
	} else {
		slog.Info("grpc server shutdown")
	}
}

func (c *Component) detectAddr(oriAddr string) (realAddr string, err error) {
	if oriAddr == "" {
		return "", nil
	}
	addr := strings.Split(oriAddr, ":")
	if len(addr) != 2 {
		err = fmt.Errorf("invalid addr: %s", oriAddr)
		return
	}

	if addr[0] != "0.0.0.0" {
		return oriAddr, nil
	}

	selfIP, err := tool.IP()
	if err != nil {
		err = fmt.Errorf("get self ip failed: %w", err)
		return
	}
	realAddr = selfIP + ":" + addr[1]
	return
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

	registryCache = r

	deRegister = func() {
		err = r.Deregister()
		if err != nil {
			slog.Warn("deregister service failed", "error", err)
		}
	}

	return
}

func traceAgent(conf *trace.TraceConfig, svr *grpc.Server, addr string) (providers map[string]*traceSdk.TracerProvider, deferFunc func()) {
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
		conf.Addr = addr
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
