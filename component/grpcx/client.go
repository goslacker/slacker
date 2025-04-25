package grpcx

import (
	"github.com/goslacker/slacker/component/grpcx/interceptor"
	"github.com/goslacker/slacker/component/grpcx/resolver"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"math"
)

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

	opts = append(opts, grpc.WithResolvers(&resolver.EtcdResolverBuilder{}))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)))
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
