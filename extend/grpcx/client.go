package grpcx

import (
	"github.com/goslacker/slacker/extend/grpcx/interceptor"
	"google.golang.org/grpc"
)

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor.UnaryTraceClientInterceptor))
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
