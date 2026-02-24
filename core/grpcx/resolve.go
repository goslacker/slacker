package grpcx

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/goslacker/slacker/core/registry"
	"google.golang.org/grpc/resolver"
)

func NewResolver(target resolver.Target, cc resolver.ClientConn, registry registry.Resolver) *Resolver {
	r := &Resolver{
		target:   target,
		cc:       cc,
		resolver: registry,
	}

	return r
}

type Resolver struct {
	target   resolver.Target
	cc       resolver.ClientConn
	resolver registry.Resolver
	cancel   context.CancelFunc
	once     sync.Once
}

func (r *Resolver) ResolveNow(p0 resolver.ResolveNowOptions) {
	fmt.Printf("resolveNow: %+v", r.target)
	r.once.Do(func() {
		go r.watch()
	})
}

func (r *Resolver) watch() {
	var ctx context.Context
	ctx, r.cancel = context.WithCancel(context.Background())
	c, err := r.resolver.Watch(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	var addresses []resolver.Address
	for addrs := range c {
		select {
		case <-ctx.Done():
			return
		default:
		}
		addresses = addresses[:0]
		addresses = slices.Grow(addresses, len(addrs))
		for _, addr := range addrs {
			addresses = append(addresses, resolver.Address{Addr: addr})
		}
		err = r.cc.UpdateState(resolver.State{Addresses: addresses})
		if err != nil {
			r.cc.ReportError(err)
		}
	}
}

func (r *Resolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}

// ResolverBuilder 需实现 Builder 接口
type ResolverBuilder struct {
	Resolver registry.Resolver
}

func (e *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewResolver(target, cc, e.Resolver)
	return r, nil
}
func (*ResolverBuilder) Scheme() string { return "" }
