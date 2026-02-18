package registry

import "context"

type Registrar interface {
	Register(ctx context.Context, service string) (err error)
}

type Resolver interface {
	Resolve(ctx context.Context, service string) (addrs []string, err error)
	Watch(ctx context.Context, service string) (addrsChan chan []string, err error)
}

type Driver interface {
	Register(ctx context.Context, service string, addr string) (err error)
	Resolve(ctx context.Context, service string) (addrs []string, err error)
	Watch(ctx context.Context, service string) (addrsChan chan []string, err error)
}
