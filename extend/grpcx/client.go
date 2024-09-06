package grpcx

import (
	"fmt"

	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/extend/grpcx/interceptor"
	"github.com/goslacker/slacker/serviceregistry/registry"
	"google.golang.org/grpc"
)

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor.UnaryTraceClientInterceptor))

	registry, err := app.Resolve[registry.ServiceRegistry]()
	if err == nil {
		target, err = registry.Resolve(target)
		if err != nil {
			err = fmt.Errorf("resolve service registry failed: %w", err)
			return
		}
	}

	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
