package grpcgatewayx

import (
	"context"

	"github.com/goslacker/slacker/core/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func RegisterGateway(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	app.RegisterListener(func(event app.AfterInit) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		gateway.Register(registers...)
	})
}

func RegisterCustomerHandler(handlers ...CustomerHandler) {
	app.RegisterListener(func(event app.AfterInit) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		gateway.RegisterCustomHandler(handlers...)
	})
}

func Config(f func(*GrpcGatewayBuilder)) {
	app.RegisterListener(func(event app.AfterInit) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		f(gateway)
	})
}
