package resolver

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/maps"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"sync"
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

	return r
}

type etcdResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	c      *clientv3.Client
}

var once sync.Once

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
	target := r.target.Endpoint()

	ipReg := regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	if !ipReg.MatchString(target) {
		var (
			addrs map[string]resolver.Address
			err   error
		)
		for {
			addrs, err = r.resolve(target)
			if err != nil {
				r.cc.ReportError(fmt.Errorf("resolve failed: %w", err))
				slog.Error("resolve failed", "target", target, "error", err)
			} else {
				err = r.cc.UpdateState(resolver.State{Addresses: maps.Values(addrs)})
				if err != nil {
					r.cc.ReportError(err)
					slog.Error("update state failed", "target", target, "error", err)
				} else {
					break
				}
			}

			s := rand.IntN(10) + 1
			time.Sleep(time.Second * time.Duration(s))
		}
		once.Do(func() {
			go r.watch(target, addrs)
		})
	} else {
		err := r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
			{Addr: target},
		}})
		if err != nil {
			r.cc.ReportError(err)
		}
	}
}

func (r *etcdResolver) Close() {
	r.c.Close()
}

func (r *etcdResolver) resolve(serviceName string) (addrs map[string]resolver.Address, err error) {
	resp, err := r.c.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	addrs = make(map[string]resolver.Address, len(resp.Kvs))
	for _, value := range resp.Kvs {
		addrs[string(value.Key)] = resolver.Address{Addr: string(value.Value)}
	}

	return
}

func (r *etcdResolver) watch(prefix string, addrList map[string]resolver.Address) {
	slog.Debug("watch service", "prefix", prefix, "addrList", addrList)
	rch := r.c.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for n := range rch {
		update := false
		for _, ev := range n.Events {
			switch ev.Type {
			case mvccpb.PUT:
				slog.Debug("receive put", "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
				if _, ok := addrList[string(ev.Kv.Key)]; !ok {
					addrList[string(ev.Kv.Key)] = resolver.Address{Addr: string(ev.Kv.Value)}
					update = true
				}
			case mvccpb.DELETE:
				slog.Debug("receive delete", "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
				if _, ok := addrList[string(ev.Kv.Key)]; ok {
					delete(addrList, string(ev.Kv.Key))
					update = true
				}
			}
		}

		if update == true {
			slog.Debug("update addrList", "addrList", addrList)
			r.cc.UpdateState(resolver.State{Addresses: maps.Values(addrList)})
		}
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
