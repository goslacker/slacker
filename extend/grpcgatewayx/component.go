package grpcgatewayx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/goslacker/slacker/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

func WithRegisters(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) func(*Component) {
	return func(c *Component) {
		c.registers = registers
	}
}

func NewComponent(opts ...func(*Component)) *Component {
	c := &Component{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Component struct {
	app.Component
	cancel    context.CancelFunc
	registers []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

// Start 启动服务并阻塞, 框架一般会将这个方法作为协程调用, 报错应打日志记录
func (c *Component) Start() {
	if len(c.registers) > 0 {
		slog.Warn("no gateway register")
		return
	}

	var ctx context.Context
	ctx, c.cancel = context.WithCancel(context.Background())

	endpoint := viper.GetString("grpc.addr")
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to dial server", "err", err)
		return
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	mux := runtime.NewServeMux()
	for _, register := range c.registers {
		err := register(ctx, mux, conn)
		if err != nil {
			slog.Error("Failed to register gateway", "err", err)
		}
	}

	gwServer := &http.Server{
		Addr:    viper.Sub("grpcGateway").GetString("addr"),
		Handler: mux,
	}

	slog.Info("Serving gRPC-Gateway on " + viper.Sub("grpcGateway").GetString("addr"))
	slog.Error("grpc server shutdown", "err", gwServer.ListenAndServe())
}

// Stop 停止服务并阻塞, 报错应打日志记录
func (c *Component) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}
