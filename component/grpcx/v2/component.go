package grpcx

import (
	"context"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/grpcx"
	"github.com/goslacker/slacker/core/registry"
	"github.com/spf13/viper"
)

func NewComponent() *Component {
	return &Component{}
}

type Component struct {
	app.Component
	server *grpcx.Server
	cancel context.CancelFunc
}

func (c *Component) Init() (err error) {
	err = app.Bind[registry.Driver](func(conf *viper.Viper) (driver registry.Driver, err error) {
		var config config
		if err = conf.UnmarshalKey("grpcx", &config); err != nil {
			return
		}
		return registry.BuildDriver(config.Registry.Type, config.Registry.Endpoints)
	})
	if err != nil {
		return
	}

	err = app.Bind[*grpcx.GrpcServerBuilder](func(conf *viper.Viper, driver registry.Driver) (server *grpcx.GrpcServerBuilder, err error) {
		var config config
		if err = conf.UnmarshalKey("grpcx", &config); err != nil {
			return
		}
		b := &grpcx.GrpcServerBuilder{
			Addr:           config.Addr,
			Network:        config.Network,
			HealthCheck:    config.HealthCheck,
			Reflection:     config.Reflection,
			RegistryConfig: &config.Registry,
			TraceConfig:    &config.Trace,
			PprofPort:      config.PprofPort,
			RegistryDriver: driver,
		}
		return b, nil
	})
	if err != nil {
		return
	}

	return
}

func (c *Component) Start() (err error) {
	builder := app.MustResolve[*grpcx.GrpcServerBuilder]()
	server, err := builder.Build()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.server = server
	c.cancel = cancel
	server.Start(ctx)
	return
}

func (c *Component) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

type config struct {
	Addr        string               `mapstructure:"addr"`
	Network     string               `mapstructure:"network"`
	HealthCheck bool                 `mapstructure:"health_check"`
	Reflection  bool                 `mapstructure:"reflection"`
	PprofPort   int                  `mapstructure:"pprof_port"`
	Registry    grpcx.RegistryConfig `mapstructure:"registry"`
	Trace       grpcx.TraceConfig    `mapstructure:"trace"`
}
