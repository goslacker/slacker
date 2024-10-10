package serviceregistry

import (
	"fmt"
	"github.com/goslacker/slacker/core/serviceregistry/registry"

	_ "github.com/goslacker/slacker/core/serviceregistry/etcd"
)

func New(conf *registry.RegistryConfig) (sr registry.ServiceRegistry, err error) {
	newFunc, ok := registry.Registers[conf.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported registry type %s", conf.Type)
	}
	return newFunc(conf)
}
