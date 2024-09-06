package serviceregistry

import (
	"fmt"

	_ "github.com/goslacker/slacker/serviceregistry/etcd"
	"github.com/goslacker/slacker/serviceregistry/registry"
)

func New(conf *registry.RegistryConfig) (sr registry.ServiceRegistry, err error) {
	newFunc, ok := registry.Registers[conf.RegistryType]
	if !ok {
		return nil, fmt.Errorf("unsupported registry type %s", conf.RegistryType)
	}
	return newFunc(conf)
}
