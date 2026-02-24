package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/maps"
)

type DefaultRegistrar struct {
	addr   string
	driver Driver
}

func NewDefaultRegistrar(addr string, driver Driver) *DefaultRegistrar {
	return &DefaultRegistrar{addr: addr, driver: driver}
}

func (r *DefaultRegistrar) Register(ctx context.Context, service string) (err error) {
	return r.driver.Register(ctx, service, r.addr)
}

type DefaultResolver struct {
	driver Driver
}

func NewDefaultResolver(driver Driver) *DefaultResolver {
	return &DefaultResolver{driver: driver}
}

func (r *DefaultResolver) Resolve(ctx context.Context, service string) (addrs []string, err error) {
	return r.driver.Resolve(ctx, service)
}

func (r *DefaultResolver) Watch(ctx context.Context, service string) (addrsChan chan []string, err error) {
	return r.driver.Watch(ctx, service)
}

type EtcdDriver struct {
	c *clientv3.Client
}

func NewEtcdDriver(c *clientv3.Client) *EtcdDriver {
	return &EtcdDriver{c: c}
}

func (e *EtcdDriver) Register(ctx context.Context, service string, addr string) (err error) {
	var resp *clientv3.LeaseGrantResponse
	{
		resp, err = e.c.Grant(ctx, 20)
		if err != nil {
			return fmt.Errorf("grant lease failed: %w", err)
		}
	}

	key := service + "/" + strconv.FormatInt(int64(resp.ID), 10)

	_, err = e.c.Put(ctx, key, addr, clientv3.WithLease(resp.ID))
	if err != nil {
		err = fmt.Errorf("put service '%s' info to etcd failed: %w", service, err)
		return
	}
	slog.Info("register service success", "service", service, "addr", addr)
	var (
		ch     <-chan *clientv3.LeaseKeepAliveResponse
		cancel context.CancelFunc
	)
	{
		c, cancelFunc := context.WithCancel(ctx)
		cancel = func() {
			e.c.Revoke(c, resp.ID)
			cancelFunc()
		}
		ch, err = e.c.KeepAlive(c, resp.ID)
		if err != nil {
			cancel()
			return fmt.Errorf("keep service '%s' alive failed: %w", service, err)
		}
	}

	go func() {
		for range ch {
		}
		cancel()
		for {
			sec := rand.IntN(5) + 1
			time.Sleep(time.Duration(sec) * time.Second)
			err = e.Register(ctx, service, addr)
			if err != nil {
				slog.Error("register service failed:", "serviceName", service, "error", err)
			} else {
				break
			}
		}
	}()

	return
}

func (e *EtcdDriver) Resolve(ctx context.Context, service string) (addrs []string, err error) {
	resp, err := e.c.Get(ctx, service, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, value := range resp.Kvs {
		addrs = append(addrs, string(value.Value))
	}

	return
}

func (e *EtcdDriver) Watch(ctx context.Context, service string) (addrsChan chan []string, err error) {
	addrsChan = make(chan []string, 1)

	resp, err := e.c.Get(ctx, service, clientv3.WithPrefix())
	if err != nil {
		slog.Error("service resolve failed", "service", service, "error", err)
		return
	}
	rev := resp.Header.Revision
	addrs := make(map[string]string, len(resp.Kvs))
	for _, value := range resp.Kvs {
		addrs[string(value.Key)] = string(value.Value)
	}
	if len(addrs) > 0 {
		addrsChan <- maps.Values(addrs)
	}

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}
	if rev != 0 {
		opts = append(opts, clientv3.WithRev(rev+1))
	}
	rch := e.c.Watch(ctx, service, opts...)
	go func() {
		defer close(addrsChan)
	LOOP:
		for res := range rch {
			select {
			case <-ctx.Done():
				break LOOP
			default:
			}
			if res.Err() != nil {
				if errors.Is(res.Err(), rpctypes.ErrCompacted) {
					slog.Debug("watch compacted", "service", service)
				} else {
					slog.Error("watch service failed", "service", service, "error", res.Err())
				}
				break
			}

			update := false
			for _, ev := range res.Events {
				switch ev.Type {
				case mvccpb.PUT:
					slog.Debug("receive put", "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
					if _, ok := addrs[string(ev.Kv.Key)]; !ok {
						addrs[string(ev.Kv.Key)] = string(ev.Kv.Value)
						update = true
					}
				case mvccpb.DELETE:
					slog.Debug("receive delete", "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
					if _, ok := addrs[string(ev.Kv.Key)]; ok {
						delete(addrs, string(ev.Kv.Key))
						update = true
					}
				}
			}

			if update == true {
				addrsChan <- maps.Values(addrs)
			}
		}
	}()
	return
}

func BuildDriver(typ string, endpoints []string) (driver Driver, err error) {
	switch typ {
	case "etcd":
		var c *clientv3.Client
		c, err = clientv3.New(clientv3.Config{
			Endpoints:            endpoints,
			DialKeepAliveTime:    30 * time.Second,
			DialKeepAliveTimeout: 60 * time.Second,
			DialTimeout:          10 * time.Second,
		})
		if err != nil {
			err = fmt.Errorf("new etcd client failed: %w", err)
			return
		}
		driver = NewEtcdDriver(c)
	default:
		err = fmt.Errorf("unknown registry type: %s", typ)
		return
	}
	return
}
