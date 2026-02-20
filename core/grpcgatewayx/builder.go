package grpcgatewayx

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcGatewayBuilder struct {
	Endpoint       string //grpc连接地址
	Addr           string //http服务地址
	Ctx            context.Context
	Registers      []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	Options        []runtime.ServeMuxOption
	CustomHandlers map[string]runtime.HandlerFunc
}

func (c *GrpcGatewayBuilder) RegisterCustomHandler(method string, path string, handler runtime.HandlerFunc) {
	if c.CustomHandlers == nil {
		c.CustomHandlers = make(map[string]runtime.HandlerFunc)
	}
	c.CustomHandlers[fmt.Sprintf("%s|%s", method, path)] = handler
}

func (c *GrpcGatewayBuilder) Register(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	c.Registers = append(c.Registers, registers...)
}

func (c *GrpcGatewayBuilder) SetOptions(options ...runtime.ServeMuxOption) {
	c.Options = append(c.Options, options...)
}

func (c *GrpcGatewayBuilder) Build() (server *Server, err error) {
	server = &Server{}
	if len(c.Registers) <= 0 {
		err = fmt.Errorf("no gateway register")
		return
	}

	conn, err := grpc.NewClient(
		c.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)),
	)
	if err != nil {
		err = fmt.Errorf("Failed to dial server: %w", err)
		return
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				slog.Error("Failed to close conn", "err", cerr)
			}
		}
	}()
	server.defers = append(server.defers, func() {
		if cerr := conn.Close(); cerr != nil {
			slog.Error("Failed to close conn", "err", cerr)
		}
	})

	mux := runtime.NewServeMux(c.Options...)
	for _, register := range c.Registers {
		err = register(c.Ctx, mux, conn)
		if err != nil {
			err = fmt.Errorf("Failed to register gateway: %w", err)
			return
		}
	}

	for key, handler := range c.CustomHandlers {
		info := strings.Split(key, "|")
		err = mux.HandlePath(info[0], info[1], handler)
		if err != nil {
			err = fmt.Errorf("Failed to set custom handler(info=%+v): %w", info, err)
			return
		}
	}

	muxWithCORS := cors.AllowAll().Handler(mux)
	server.Server = &http.Server{
		Addr:    c.Addr,
		Handler: muxWithCORS,
	}

	return
}
