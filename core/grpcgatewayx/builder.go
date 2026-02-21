package grpcgatewayx

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GrpcGatewayBuilder struct {
	Endpoint       string //grpc连接地址
	Addr           string //http服务地址
	Registers      []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	Options        []runtime.ServeMuxOption
	ClientOpts     []grpc.DialOption
	MetadataFuncs  []func(context.Context, *http.Request) metadata.MD
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

func (c *GrpcGatewayBuilder) SetClientOptions(opts ...grpc.DialOption) {
	c.ClientOpts = append(c.ClientOpts, opts...)
}

func (c *GrpcGatewayBuilder) SetMetadataFuncs(fns ...func(context.Context, *http.Request) metadata.MD) {
	c.MetadataFuncs = append(c.MetadataFuncs, fns...)
}

func (c *GrpcGatewayBuilder) Build() (server *Server, err error) {
	server = &Server{}
	if len(c.Registers) <= 0 {
		err = fmt.Errorf("no gateway register")
		return
	}

	conn, err := grpc.NewClient(
		c.Endpoint,
		c.ClientOpts...,
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

	defaultOpts := []runtime.ServeMuxOption{
		runtime.WithForwardResponseRewriter(DefaultResponseRewriter),
		runtime.WithErrorHandler(DefaultErrorHandler),
	}
	c.Options = append(defaultOpts, c.Options...)

	if len(c.MetadataFuncs) > 0 {
		c.Options = append(c.Options, runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			result := make(metadata.MD)
			for _, v := range c.MetadataFuncs {
				md := v(ctx, request)
				if len(md) == 0 {
					continue
				}
				for k, v := range md {
					result[k] = v
				}
			}
			return result
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	server.defers = append(server.defers, cancel)

	mux := runtime.NewServeMux(c.Options...)
	for _, register := range c.Registers {
		err = register(ctx, mux, conn)
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
