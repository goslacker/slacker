package grpcx

import (
	"github.com/goslacker/slacker/core/app"
	"google.golang.org/grpc"
)

func RegisterGrpcService(registers ...func(grpc.ServiceRegistrar)) {
	app.RegisterListener(func(event app.AfterInit) {
		app.MustResolve[*GrpcServerBuilder]().Register(registers...)
	})
}

func Config(f func(*GrpcServerBuilder)) {
	app.RegisterListener(func(event app.AfterInit) {
		server := app.MustResolve[*GrpcServerBuilder]()
		f(server)
	})
}
