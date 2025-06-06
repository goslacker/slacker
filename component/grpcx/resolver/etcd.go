package resolver

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/maps"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/goslacker/slacker/core/serviceregistry/registry"
	"github.com/goslacker/slacker/core/trace"
	"github.com/spf13/viper"
	"google.golang.org/grpc/resolver"
)

type Config struct {
	HealthCheck bool                     //是否开启健康检查
	Reflection  bool                     //是否开启反射服务
	Addr        string                   //服务地址
	Trace       *trace.TraceConfig       //启链路追踪配置
	Registry    *registry.RegistryConfig //服务注册中心配置
}

func NewEtcdResolver(target resolver.Target, cc resolver.ClientConn) *etcdResolver {
	r := &etcdResolver{
		target: target,
		cc:     cc,
	}

	var conf Config
	err := viper.Sub("grpcx").Unmarshal(&conf)
	if err != nil {
		slog.Error("new registry failed", "error", err)
		return nil
	}
	r.c, err = clientv3.New(clientv3.Config{
		Endpoints: conf.Registry.Endpoints,
	})
	if err != nil {
		slog.Error("new etcd client failed", "error", err)
		return nil
	}

	shouldWatch, addrs, err := r.resolve(target.Endpoint())
	if err != nil {
		slog.Error("service resolve failed", "service", target.Endpoint(), "error", err)
	}

	if shouldWatch {
		go r.watch(target.Endpoint(), addrs)
	}

	return r
}

type etcdResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	c      *clientv3.Client
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *etcdResolver) Close() {
	r.c.Close()
}

func (r *etcdResolver) resolve(target string) (shouldWatch bool, addrs map[string]resolver.Address, err error) {
	ipReg := regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if !ipReg.MatchString(target) {
		var resp *clientv3.GetResponse
		resp, err = r.c.Get(ctx, target, clientv3.WithPrefix())
		if err != nil {
			err = fmt.Errorf("service resolve failed: %w(service=%s)", err, target)
			return
		}
		addrs = make(map[string]resolver.Address, len(resp.Kvs))
		for _, value := range resp.Kvs {
			addrs[string(value.Key)] = resolver.Address{Addr: string(value.Value)}
		}
		err = r.cc.UpdateState(resolver.State{Addresses: maps.Values(addrs)})
		if err != nil {
			err = fmt.Errorf("service update state failed: %w(service=%s)", err, target)
			return
		}
		shouldWatch = true
	} else {
		err = r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
			{Addr: target},
		}})
		if err != nil {
			err = fmt.Errorf("service update state failed: %w(service=%s)", err, target)
			return
		}
	}
	return
}

func (r *etcdResolver) watch(prefix string, addrs map[string]resolver.Address) {
	for {
		slog.Debug("watch service", "prefix", prefix, "addrs", addrs)
		ctx, cancel := context.WithCancel(context.Background())
		opts := []clientv3.OpOption{
			clientv3.WithPrefix(),
		}
		rch := r.c.Watch(ctx, prefix, opts...)
		for n := range rch {
			if n.Err() != nil {
				cancel()
				if errors.Is(n.Err(), rpctypes.ErrCompacted) {
					slog.Debug("watch compacted, reload", "prefix", prefix)
					var err error
					_, addrs, err = r.resolve(prefix)
					if err != nil {
						slog.Error("service resolve failed", "service", prefix, "error", err)
					}
				} else {
					slog.Error("watch service failed", "prefix", prefix, "error", n.Err())
				}
				break
			}

			update := false
			for _, ev := range n.Events {
				switch ev.Type {
				case mvccpb.PUT:
					slog.Debug("receive put", "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
					if _, ok := addrs[string(ev.Kv.Key)]; !ok {
						addrs[string(ev.Kv.Key)] = resolver.Address{Addr: string(ev.Kv.Value)}
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
				slog.Debug("update service state", "addrs", addrs)
				err := r.cc.UpdateState(resolver.State{Addresses: maps.Values(addrs)})
				if err != nil {
					slog.Error("update service state failed", "service", prefix, "error", err)
				}
			}
		}
		cancel()
		s := rand.IntN(10) + 1
		time.Sleep(time.Second * time.Duration(s))
	}
}

// EtcdResolverBuilder 需实现 Builder 接口
type EtcdResolverBuilder struct{}

func (e *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewEtcdResolver(target, cc)
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*EtcdResolverBuilder) Scheme() string { return "" }
