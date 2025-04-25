package etcd

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/goslacker/slacker/core/serviceregistry/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func init() {
	registry.Register(registry.Etcd, NewRegistry)
}

func NewRegistry(conf *registry.RegistryConfig) (regsitry registry.ServiceRegistry, err error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: conf.Endpoints,
	})
	if err != nil {
		return
	}

	return &Registry{
		c:    c,
		addr: conf.Addr,
	}, nil
}

type Registry struct {
	c    *clientv3.Client
	addr string
}

func (r *Registry) Register(serviceName string) (err error) {
	var resp *clientv3.LeaseGrantResponse
	resp, err = r.c.Grant(context.Background(), 10)
	if err != nil {
		return fmt.Errorf("grant lease failed: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := r.c.KeepAlive(ctx, resp.ID)
	if err != nil {
		r.c.Revoke(context.Background(), resp.ID)
		cancel()
		return fmt.Errorf("keep alive failed: %w", err)
	}

	go func() {
		defer cancel()
		for range ch {
		}

		for {
			sec := rand.IntN(5) + 1
			time.Sleep(time.Duration(sec) * time.Second)
			err = r.Register(serviceName)
			if err != nil {
				slog.Error("register service failed:", err)
			} else {
				break
			}
		}
	}()

	key := serviceName + "/" + strconv.FormatInt(int64(resp.ID), 10)

	_, err = r.c.Put(context.Background(), key, r.addr, clientv3.WithLease(resp.ID))
	if err != nil {
		return
	}
	slog.Info("register service success", "service", serviceName)

	return
}

func (r *Registry) Resolve(serviceName string) (addrs []string, err error) {
	resp, err := r.c.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, value := range resp.Kvs {
		addrs = append(addrs, string(value.Value))
	}

	return
}

func (r *Registry) Close() error {
	return r.c.Close()
}
