package grpcx

import (
	"context"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/grpcx"
	"github.com/goslacker/slacker/core/registry"
	"github.com/goslacker/slacker/core/trace"
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

func (c *Component) getConfig(conf *viper.Viper) (cfg config) {
	cfg = config{
		Addr:        conf.GetString("grpcx.addr"),
		Network:     conf.GetString("grpcx.network"),
		HealthCheck: conf.GetBool("grpcx.health_check"),
		Reflection:  conf.GetBool("grpcx.reflection"),
		PprofPort:   conf.GetInt("grpcx.pprof_port"),
		Registry: grpcx.RegistryConfig{
			Type:      conf.GetString("grpcx.registry.type"),
			Endpoints: conf.GetStringSlice("grpcx.registry.endpoints"),
		},
		Trace: grpcx.TraceConfig{
			Type:     trace.TraceType(conf.GetString("grpcx.trace.type")),
			Endpoint: conf.GetString("grpcx.trace.endpoint"),
		},
	}
	return
}

func (c *Component) Init() (err error) {
	err = app.Bind[registry.Driver](func(conf *viper.Viper) (driver registry.Driver, err error) {
		cfg := c.getConfig(conf)
		return registry.BuildDriver(cfg.Registry.Type, cfg.Registry.Endpoints)
	})
	if err != nil {
		return
	}

	err = app.Bind[*grpcx.GrpcServerBuilder](func(conf *viper.Viper, driver registry.Driver) (server *grpcx.GrpcServerBuilder, err error) {
		// conf.UnmarshalKey("grpcx", &config)有问题, 不能取到环境变量中的Network字段
		cfg := c.getConfig(conf)
		b := &grpcx.GrpcServerBuilder{
			Addr:           cfg.Addr,
			Network:        cfg.Network,
			HealthCheck:    cfg.HealthCheck,
			Reflection:     cfg.Reflection,
			RegistryConfig: &cfg.Registry,
			TraceConfig:    &cfg.Trace,
			PprofPort:      cfg.PprofPort,
			RegistryDriver: driver,
		}
		return b, nil
	})
	if err != nil {
		return
	}

	return
}

func (c *Component) Start() {
	builder := app.MustResolve[*grpcx.GrpcServerBuilder]()
	server, err := builder.Build()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.server = server
	c.cancel = cancel
	server.Start(ctx)
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
