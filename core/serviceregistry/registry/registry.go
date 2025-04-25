package registry

type RegistryType string

const (
	Consul RegistryType = "consul"
	Etcd   RegistryType = "etcd"
)

type RegistryConfig struct {
	Type      RegistryType
	Endpoints []string
	Addr      string
}

type ServiceRegistry interface {
	Register(serviceName string) (err error)
	Resolve(serviceName string) (addrs []string, err error)
	Close() error
}

var Registers = map[RegistryType]func(conf *RegistryConfig) (ServiceRegistry, error){}

func Register(registryType RegistryType, f func(conf *RegistryConfig) (ServiceRegistry, error)) {
	Registers[registryType] = f
}
