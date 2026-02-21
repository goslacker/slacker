package grpcgatewayx

import (
	"fmt"
	"log/slog"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/grpcgatewayx"
	"github.com/spf13/viper"
)

type config struct {
	Endpoint string `mapstructure:"endpoint"`
	Addr     string `mapstructure:"addr"`
}

func NewComponent() *Component {
	return &Component{}
}

type Component struct {
	app.Component
	server *grpcgatewayx.Server
}

func (c *Component) Init() (err error) {
	err = app.Bind[*grpcgatewayx.GrpcGatewayBuilder](func(conf *viper.Viper) (builder *grpcgatewayx.GrpcGatewayBuilder, err error) {
		var cfg config
		if err = conf.UnmarshalKey("grpcgatewayx", &cfg); err != nil {
			err = fmt.Errorf("Failed to unmarshal config: %w", err)
			return
		}
		b := &grpcgatewayx.GrpcGatewayBuilder{
			Endpoint: cfg.Endpoint,
			Addr:     cfg.Addr,
		}

		return b, nil
	})
	if err != nil {
		return
	}
	return
}

func (c *Component) Start() {
	builder := app.MustResolve[*grpcgatewayx.GrpcGatewayBuilder]()
	server, err := builder.Build()
	if err != nil {
		slog.Error("Failed to build grpc gateway server", "err", err)
		return
	}
	if err = server.Start(); err != nil {
		slog.Error("Failed to start grpc gateway server", "err", err)
		return
	}
}

func (c *Component) Stop() {
	if c.server == nil {
		return
	}
	if err := c.server.Stop(); err != nil {
		slog.Error("Failed to stop grpc gateway server", "err", err)
	}
}
