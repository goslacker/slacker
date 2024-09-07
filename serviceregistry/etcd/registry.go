package etcd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/goslacker/slacker/serviceregistry/registry"
	"github.com/goslacker/slacker/tool"
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
	leaseID, err := r.getLeaseID()
	if err != nil {
		err = fmt.Errorf("get lease id failed: %w", err)
		return
	}

	_, err = r.c.Put(context.Background(), serviceName+"/"+strconv.FormatInt(int64(r.leaseID), 10), r.GetAddr(), clientv3.WithLease(leaseID))
	if err != nil {
		return
	}
	slog.Info("register service success", "service", serviceName)
	return
}

func (r *Registry) Resolve(serviceName string) (addr string, err error) {
	resp, err := r.c.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, value := range resp.Kvs {
		addr = string(value.Value)
		break
	}

	//TODO: watch

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

func (r *Registry) GetAddr() string {
	if r.addr == "" {
		return ""
	}
	addr := strings.Split(r.addr, ":")
	if len(addr) != 2 {
		panic(fmt.Errorf("invalid addr: %s", r.addr))
	}

	if addr[0] != "0.0.0.0" {
		return r.addr
	}

	selfIP, err := tool.SelfIP(r.c.Endpoints()[0])
	if err != nil {
		panic(fmt.Errorf("get self ip failed: %w", err))
	}

	return selfIP + ":" + addr[1]
}

func (r *Registry) getLeaseID() (leaseID clientv3.LeaseID, err error) {
	if r.leaseID == 0 {
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
		}()
	}
	return r.leaseID, nil
}