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

var unaryClientInterceptors = []grpc.UnaryClientInterceptor{
	interceptor.UnaryTraceClientInterceptor,
}
var streamClientInterceptors = []grpc.StreamClientInterceptor{
	interceptor.StreamTraceClientInterceptor,
}

func RegisterUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) {
	unaryClientInterceptors = append(unaryClientInterceptors, interceptors...)
}

func RegisterStreamClientInterceptors(interceptors ...grpc.StreamClientInterceptor) {
	streamClientInterceptors = append(streamClientInterceptors, interceptors...)
}

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	var conf Config
	err = viper.Sub("grpcx").Unmarshal(&conf)
	if err != nil {
		return
	}

	if conf.Trace != nil {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(unaryClientInterceptors...),
			grpc.WithChainStreamInterceptor(streamClientInterceptors...),
		)
	}

	opts = append(opts, grpc.WithResolvers(&resolver.EtcdResolverBuilder{RegistryCache: &registryCache}))
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
