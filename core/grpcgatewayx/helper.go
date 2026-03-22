package grpcgatewayx

import (
	"context"

	"github.com/goslacker/slacker/core/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func RegisterGateway(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	app.RegisterListener(func(event app.AfterInit) (err error) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		gateway.Register(registers...)
		return
	})
}

func RegisterCustomerHandler(handlers ...CustomerHandler) {
	app.RegisterListener(func(event app.AfterInit) (err error) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		gateway.RegisterCustomHandler(handlers...)
		return
	})
}

func Config(f func(*GrpcGatewayBuilder)) {
	app.RegisterListener(func(event app.AfterInit) (err error) {
		gateway := app.MustResolve[*GrpcGatewayBuilder]()
		f(gateway)
		return
	})
}
