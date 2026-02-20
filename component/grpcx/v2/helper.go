package grpcx

import (
	"github.com/goslacker/slacker/core/app"
	"github.com/goslacker/slacker/core/grpcx"
	"google.golang.org/grpc"
)

func RegisterGrpcService(registers ...func(grpc.ServiceRegistrar)) {
	app.RegisterListener(func(event app.AfterInit) {
		app.MustResolve[*grpcx.GrpcServerBuilder]().Register(registers...)
	})
}
