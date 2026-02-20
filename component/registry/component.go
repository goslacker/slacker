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

func (c *Component) Init() (err error) {
	err = app.Bind[*registry.Driver](func(conf *viper.Viper) (driver registry.Driver, err error) {
		registryConf := &RegistryConfig{}
		if err = conf.UnmarshalKey("grpcx.registry", registryConf); err != nil {
			return
		}
		return registry.BuildDriver(registryConf.Type, registryConf.Endpoints)
	})

	return
}

func NewComponent() *Component {
	return &Component{}
}
