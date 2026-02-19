package grpcx

import (
	"google.golang.org/grpc"
)

func NewClient[T any](target string, provider func(cc grpc.ClientConnInterface) T, opts ...grpc.DialOption) (result T, err error) {
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return
	}

	return provider(cc), nil
}
