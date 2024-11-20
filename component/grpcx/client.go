package grpcx

import (
	"github.com/goslacker/slacker/component/grpcx/interceptor"
	"github.com/goslacker/slacker/component/grpcx/resolver"
	"github.com/goslacker/slacker/core/serviceregistry/registry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// registryCache client用来查询服务
var registryCache registry.ServiceRegistry

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	var conf Config
	err = viper.Sub("grpcx").Unmarshal(&conf)
	if err != nil {
		return
	}

	if conf.Trace != nil {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(interceptor.UnaryTraceClientInterceptor, interceptor.UnaryThroughClientInterceptor),
			grpc.WithChainStreamInterceptor(interceptor.StreamTraceClientInterceptor, interceptor.StreamThroughClientInterceptor),
		)
	}

	opts = append(opts, grpc.WithResolvers(&resolver.EtcdResolverBuilder{RegistryCache: &registryCache}))
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
