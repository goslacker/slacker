package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/maps"
)

type DefaultRegistrar struct {
	addr    string
	network string
	driver  Driver
}

func NewDefaultRegistrar(addr string, network string, driver Driver) *DefaultRegistrar {
	return &DefaultRegistrar{addr: addr, network: network, driver: driver}
}

func (r *DefaultRegistrar) Register(ctx context.Context, service string) (err error) {
	addr, err := DetectAddr(r.addr, r.network)
	if err != nil {
		err = fmt.Errorf("detect addr failed: %w", err)
		return
	}
	return r.driver.Register(ctx, service, addr)
}

func DetectAddr(oriAddr string, network string) (realAddr string, err error) {
	arr := strings.Split(oriAddr, ":")
	if len(arr) != 2 {
		err = fmt.Errorf("invalid oriAddr: %s, forget port?", oriAddr)
		return
	}
	hostname := arr[0]
	if hostname != "" && hostname != "localhost" && hostname != "127.0.0.1" && hostname != "0.0.0.0" {
		realAddr = oriAddr
		return
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		err = fmt.Errorf("get interfaces failed: %w", err)
		return
	}

	var ip string
	for _, iface := range interfaces {
		if network != "" && iface.Name != network {
			continue
		}
		var addrs []net.Addr
		addrs, err = iface.Addrs()
		if err != nil {
			err = fmt.Errorf("get iface %s's addrs failed: %w", iface.Name, err)
			return
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil { //目前只持支ipv4
				ip = ipnet.IP.String()
				break
			}
		}
	}

	if ip == "" {
		err = fmt.Errorf("no valid address found")
		return
	}

	realAddr = ip + ":" + arr[1]

	return
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
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		resp, err = e.c.Grant(ctx, 10)
		cancel()
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
		ctx := ctx
		ctx, cancel = context.WithCancel(ctx)
		ch, err = e.c.KeepAlive(ctx, resp.ID)
		if err != nil {
			e.c.Revoke(ctx, resp.ID)
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
		for res := range rch {
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
			Endpoints: endpoints,
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
