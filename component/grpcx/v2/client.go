package grpcx

import (
	"errors"
	"math"
	"time"

	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/container"
	"github.com/goslacker/slacker/core/grpcx"
	"github.com/goslacker/slacker/core/grpcx/interceptor"
	"github.com/goslacker/slacker/core/registry"
	"github.com/goslacker/slacker/core/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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
	opts = append(opts, grpc.WithUnaryInterceptor(trace.UnaryTraceClientInterceptor), grpc.WithStreamInterceptor(trace.StreamTraceClientInterceptor))
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor.UnaryThroughClientInterceptor), grpc.WithStreamInterceptor(interceptor.StreamThroughClientInterceptor))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)))
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    10 * time.Second,
		Timeout: 5 * time.Second,
	}))
	return grpcx.NewClient(target, provider, opts...)
}
