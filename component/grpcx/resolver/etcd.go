package resolver

import (
	"fmt"
	"github.com/goslacker/slacker/core/serviceregistry"
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

func NewEtcdResolverV2(target resolver.Target, cc resolver.ClientConn) *etcdResolver {
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
	r.registry, err = serviceregistry.New(conf.Registry)
	if err != nil {
		slog.Error("new registry failed", "error", err)
		return nil
	}

	return r
}

type etcdResolver struct {
	target   resolver.Target
	cc       resolver.ClientConn
	registry registry.ServiceRegistry
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
	target := r.target.Endpoint()

	ipReg := regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	if !ipReg.MatchString(target) {
		for {
			addrs, err := r.registry.Resolve(target)
			if err != nil {
				r.cc.ReportError(fmt.Errorf("resolve failed: %w", err))
				slog.Error("resolve failed", "target", target, "error", err)
			} else {
				addresses := make([]resolver.Address, 0, len(addrs))
				for _, addr := range addrs {
					addresses = append(addresses, resolver.Address{Addr: addr, ServerName: target})
				}
				err = r.cc.UpdateState(resolver.State{Addresses: addresses})
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
	r.registry.Close()
}

// EtcdResolverBuilder 需实现 Builder 接口
type EtcdResolverBuilder struct{}

func (e *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewEtcdResolverV2(target, cc)
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*EtcdResolverBuilder) Scheme() string { return "" }
