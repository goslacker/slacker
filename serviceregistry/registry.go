package serviceregistry

import (
	"fmt"

	_ "github.com/goslacker/slacker/serviceregistry/etcd"
	"github.com/goslacker/slacker/serviceregistry/registry"
)

func New(conf *registry.RegistryConfig) (sr registry.ServiceRegistry, err error) {
	newFunc, ok := registry.Registers[conf.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported registry type %s", conf.Type)
	}
	return newFunc(conf)
}
