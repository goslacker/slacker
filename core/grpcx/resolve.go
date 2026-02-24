package grpcx

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"

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
	watched  atomic.Bool
	once     sync.Once
}

func (r *Resolver) ResolveNow(p0 resolver.ResolveNowOptions) {
	if r.watched.Load() {
		return
	}
	r.watched.Store(true)
	go r.watch()
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
			slog.Error("update grpc state failed", "error", err)
			r.cc.ReportError(err)
		}
	}
}

func (r *Resolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	r.watched.Store(false)
}

// ResolverBuilder 需实现 Builder 接口
type ResolverBuilder struct {
	Resolver registry.Resolver
}

func (e *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewResolver(target, cc, e.Resolver)
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*ResolverBuilder) Scheme() string { return "" }
