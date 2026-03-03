package registry

import (
	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/registry"
	"github.com/spf13/viper"
)

type RegistryConfig struct {
	Type      string
	Endpoints []string
	Addr      string
	Network   string
}

type Component struct {
	app.Component
}

func (c *Component) getConfig(conf *viper.Viper) (cfg RegistryConfig) {
	cfg = RegistryConfig{
		Type:      conf.GetString("grpcx.registry.type"),
		Endpoints: conf.GetStringSlice("grpcx.registry.endpoints"),
	}
	return
}

func (c *Component) Init() (err error) {
	err = app.Bind[*registry.Driver](func(conf *viper.Viper) (driver registry.Driver, err error) {
		cfg := c.getConfig(conf)
		return registry.BuildDriver(cfg.Type, cfg.Endpoints)
	})

	return
}

func NewComponent() *Component {
	return &Component{}
}
