package grpcx

import (
	"github.com/goslacker/slacker/core/app"
	"google.golang.org/grpc"
)

func RegisterGrpcService(registers ...func(grpc.ServiceRegistrar)) {
	app.RegisterListener(func(event app.AfterInit) (err error) {
		app.MustResolve[*GrpcServerBuilder]().Register(registers...)
		return
	})
}

func Config(f func(*GrpcServerBuilder)) {
	app.RegisterListener(func(event app.AfterInit) (err error) {
		server := app.MustResolve[*GrpcServerBuilder]()
		f(server)
		return
	})
}
