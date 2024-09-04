package grpcx

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/goslacker/slacker/app"
	"github.com/spf13/viper"
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

func (m *Component) Start() {
	c := viper.Sub("grpc")

	m.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(m.middlewares...),
	)

	if c.GetBool("healthCheck") {
		healthCheck := health.NewServer()
		healthgrpc.RegisterHealthServer(m.grpcServer, healthCheck)
		err := app.Bind[*health.Server](healthCheck)
		if err != nil {
			slog.Error("bind health check failed", "error", err)
			return
		}
	}

	if c.GetBool("reflection") {
		reflection.Register(m.grpcServer)
	}

	if len(m.registers) <= 0 {
		slog.Warn("no grpc service registered")
		return
	}

	for _, register := range m.registers {
		register(m.grpcServer)
	}

	var lis net.Listener
	lis, err := net.Listen("tcp", c.GetString("addr"))
	if err != nil {
		panic(fmt.Errorf("tcp listen port failed: %w", err))
	}

	slog.Info("Serving gRPC on " + c.GetString("addr"))
	err = m.grpcServer.Serve(lis)
	if err != nil {
		slog.Error("grpc server shutdown", "error", err)
	} else {
		slog.Info("grpc server shutdown")
	}
}

func (c *Component) Stop() {
	c.grpcServer.Stop()
}
