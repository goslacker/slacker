package grpcx

import (
	"fmt"
	"github.com/goslacker/slacker/component/grpcx/interceptor"
	"github.com/goslacker/slacker/core/serviceregistry/registry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"regexp"
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

	ipReg := regexp.MustCompile(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	if !ipReg.MatchString(target) {
		if conf.Registry != nil && registryCache != nil {
			target, err = registryCache.Resolve(target)
			if err != nil {
				err = fmt.Errorf("resolve service registry failed: %w", err)
				return
			}
		}
	}

	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
