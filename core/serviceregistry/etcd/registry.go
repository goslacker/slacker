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
	leaseID clientv3.LeaseID
	c       *clientv3.Client
	addr    string
}

func (r *Registry) Register(serviceName string) (err error) {
	leaseID, err := r.getLeaseID(false)
	if err != nil {
		err = fmt.Errorf("get lease id failed: %w", err)
		return
	}

	key := serviceName + "/" + strconv.FormatInt(int64(r.leaseID), 10)

	_, err = r.c.Put(context.Background(), key, r.addr, clientv3.WithLease(leaseID))
	if err != nil {
		return
	}
	slog.Info("register service success", "service", serviceName)

	go r.watch(key, func() {
		sec := rand.IntN(5) + 1
		time.Sleep(time.Duration(sec) * time.Second)
		err := r.Register(serviceName)
		if err != nil {
			slog.Error("register service failed", "service", serviceName, "err", err)
		}
	})

	return
}

func (r *Registry) watch(key string, register func()) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watcher := r.c.Watch(ctx, key, clientv3.WithPrefix())
	for resp := range watcher {
		if resp.CompactRevision > 0 {
			slog.Info("rewatch key", "key", key)
			go r.watch(key, register)
			return
		}
		for _, event := range resp.Events {
			switch event.Type {
			case clientv3.EventTypeDelete:
				register()
				slog.Info("Service reregistered", "service", string(event.Kv.Key))
				return
			}
		}
	}
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

func (r *Registry) Deregister() (err error) {
	if r.leaseID != 0 {
		_, err = r.c.Revoke(context.Background(), r.leaseID)
	}
	if r.c != nil {
		r.c.Close()
		r.c = nil
	}
	return
}

func (r *Registry) getLeaseID(force bool) (leaseID clientv3.LeaseID, err error) {
	if r.leaseID == 0 || force {
		var resp *clientv3.LeaseGrantResponse
		resp, err = r.c.Grant(context.Background(), 10)
		if err != nil {
			err = fmt.Errorf("grant lease failed: %w", err)
			return
		}
		r.leaseID = resp.ID
		ch, err := r.c.KeepAlive(context.Background(), r.leaseID)
		if err != nil {
			r.c.Revoke(context.Background(), r.leaseID)
			return 0, fmt.Errorf("keep alive failed: %w", err)
		}
		go func() {
			for range ch {
				// slog.Debug("keep alive success", "response", resp)
			}
			sec := rand.IntN(5) + 1
			time.Sleep(time.Duration(sec) * time.Second)
			_, err = r.getLeaseID(true)
			if err != nil {
				slog.Error("get lease id failed", "err", err)
			}
		}()
	}
	return r.leaseID, nil
}
