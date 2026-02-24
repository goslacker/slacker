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
	wg       sync.WaitGroup
}

func (r *Resolver) ResolveNow(p0 resolver.ResolveNowOptions) {
	if !r.watched.CompareAndSwap(false, true) {
		return
	}
	r.wg.Add(1)
	go r.watch()
}

func (r *Resolver) watch() {
	defer func() {
		r.watched.Store(false)
		r.wg.Done()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := r.resolver.Watch(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	r.cancel = cancel
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
	r.wg.Wait()
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
