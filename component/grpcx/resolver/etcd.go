package resolver

import (
	"fmt"
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

// etcdResolver 自定义name resolver，实现Resolver接口
type etcdResolver struct {
	target        resolver.Target
	cc            resolver.ClientConn
	registryCache *registry.ServiceRegistry
}

func (r *etcdResolver) doResolveNow(target string) (err error) {
	slog.Debug("resolve", "target", target)
	addrs, err := (*r.registryCache).Resolve(target)
	if err != nil {
		err = fmt.Errorf("resolve service registry failed: %w", err)
		r.cc.ReportError(err)
		slog.Error("resolve failed", "error", err)
		return
	}
	addresses := make([]resolver.Address, 0, len(addrs))
	for _, addr := range addrs {
		addresses = append(addresses, resolver.Address{Addr: addr, ServerName: target})
	}
	err = r.cc.UpdateState(resolver.State{Addresses: addresses})
	if err != nil {
		err = fmt.Errorf("resolve service registry failed: %w", err)
		r.cc.ReportError(err)
		slog.Error("resolve failed", "error", err, "addresses", addresses)
	}
	return
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
	target := r.target.Endpoint()
	var conf Config
	err := viper.Sub("grpcx").Unmarshal(&conf)
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	ipReg := regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	if !ipReg.MatchString(target) {
		if conf.Registry != nil && *r.registryCache != nil {
			err = r.doResolveNow(target)
			if err != nil {
				go func() {
					for err != nil {
						s := rand.IntN(10) + 1
						time.Sleep(time.Second * time.Duration(s))
						err = r.doResolveNow(target)
					}
				}()
			}
		}
	} else {
		err = r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
			{Addr: target},
		}})
	}
	if err != nil {
		r.cc.ReportError(err)
	}
}

func (*etcdResolver) Close() {}

// EtcdResolverBuilder 需实现 Builder 接口
type EtcdResolverBuilder struct {
	RegistryCache *registry.ServiceRegistry
}

func (e *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &etcdResolver{
		target:        target,
		cc:            cc,
		registryCache: e.RegistryCache,
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*EtcdResolverBuilder) Scheme() string { return "" }
