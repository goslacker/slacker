package grpcgatewayx

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type GrpcGatewayBuilder struct {
	Endpoint       string //grpc连接地址
	Addr           string //http服务地址
	Registers      []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	Options        []runtime.ServeMuxOption
	ClientOpts     []grpc.DialOption
	MetadataFuncs  []MetadataFunc
	CustomHandlers map[HandlerKey]runtime.HandlerFunc
	Middlewares    []runtime.Middleware
}

func (c *GrpcGatewayBuilder) RegisterCustomHandler(handlers ...CustomerHandler) {
	if c.CustomHandlers == nil {
		c.CustomHandlers = make(map[HandlerKey]runtime.HandlerFunc)
	}
	for _, handler := range handlers {
		c.CustomHandlers[handler.Key()] = handler.Handler
	}
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

func (c *GrpcGatewayBuilder) SetMetadataFuncs(fns ...MetadataFunc) {
	c.MetadataFuncs = append(c.MetadataFuncs, fns...)
}

func (c *GrpcGatewayBuilder) SetMiddlewares(middlewares ...runtime.Middleware) {
	c.Middlewares = append(c.Middlewares, middlewares...)
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

	// 默认配置
	defaultOpts := []runtime.ServeMuxOption{
		runtime.WithForwardResponseRewriter(DefaultResponseRewriter),
		runtime.WithErrorHandler(DefaultErrorHandler),
		runtime.SetQueryParameterParser(&QueryParser{}),
	}
	c.Options = append(defaultOpts, c.Options...)

	if len(c.MetadataFuncs) > 0 {
		c.Options = append(c.Options, runtime.WithMetadata(ChainMetadataFuncs(c.MetadataFuncs...)))
	}

	c.Options = append(c.Options, runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseEnumNumbers:  true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	))

	if len(c.Middlewares) > 0 {
		c.Options = append(c.Options, runtime.WithMiddlewares(c.Middlewares...))
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
		err = mux.HandlePath(key.Method(), key.Path(), handler)
		if err != nil {
			err = fmt.Errorf("Failed to set custom handler(info=%+v): %w", key, err)
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

// ManualParseGrpcContext 手动解析grpc上下文, 并将上下文设置到http请求上下文
// 当ctx为nil时, 表示中间件未全部通过, 直接返回.(出错时,中间件自己负责返回响应)
func (c *GrpcGatewayBuilder) ManualParseGrpcContext(w http.ResponseWriter, r *http.Request, pathParams map[string]string) (ctx context.Context) {
	middleware := chainMiddlewares(nil)

	metaParser := ChainMetadataFuncs(c.MetadataFuncs...)

	middleware(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		ctx = r.Context()
		meta := metaParser(ctx, r)
		ctx = metadata.NewIncomingContext(ctx, meta)
	})(w, r, pathParams)

	return
}

func chainMiddlewares(mws []runtime.Middleware) runtime.Middleware {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		for i := len(mws); i > 0; i-- {
			next = mws[i-1](next)
		}
		return next
	}
}

type MetadataFunc func(context.Context, *http.Request) metadata.MD

func ChainMetadataFuncs(funcs ...MetadataFunc) MetadataFunc {
	return func(ctx context.Context, request *http.Request) metadata.MD {
		result := make(metadata.MD)
		for _, v := range funcs {
			md := v(ctx, request)
			if len(md) == 0 {
				continue
			}
			for k, v := range md {
				result[k] = v
			}
		}
		return result
	}
}
