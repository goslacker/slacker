package grpcx

import (
	"errors"
	"math"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/container"
	"github.com/goslacker/slacker/core/grpcx"
	"github.com/goslacker/slacker/core/registry"
	"github.com/goslacker/slacker/core/trace"
	"google.golang.org/grpc"
)

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	driver, err := app.Resolve[registry.Driver]()
	if err != nil && !errors.Is(err, container.ErrNotFound) {
		return
	}
	err = nil
	if driver != nil {
		registryResolver := registry.NewDefaultResolver(driver)
		opts = append(opts, grpc.WithResolvers(&grpcx.ResolverBuilder{Resolver: registryResolver}))
	}
	opts = append(opts, grpc.WithUnaryInterceptor(trace.UnaryTraceClientInterceptor))
	opts = append(opts, grpc.WithStreamInterceptor(trace.StreamTraceClientInterceptor))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)))
	return grpcx.NewClient(target, provider, opts...)
}
