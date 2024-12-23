package resolver

import (
	"fmt"
	"github.com/goslacker/slacker/core/serviceregistry/registry"
	"github.com/goslacker/slacker/core/trace"
	"github.com/spf13/viper"
	"google.golang.org/grpc/resolver"
	"regexp"
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
			target, err = (*r.registryCache).Resolve(target)
			if err != nil {
				r.cc.ReportError(fmt.Errorf("resolve service registry failed: %w", err))
				return
			}
		}
	}

	r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
		{Addr: target},
	}})
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
