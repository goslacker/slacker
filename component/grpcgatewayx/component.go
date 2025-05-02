package grpcgatewayx

import (
	"context"
	"fmt"
	"github.com/goslacker/slacker/component/grpcgatewayx/annotator"
	"github.com/goslacker/slacker/component/grpcgatewayx/middleware"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/goslacker/slacker/core/app"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
)

func WithRegisters(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) func(*Component) {
	return func(c *Component) {
		c.registers = registers
	}
}

func NewComponent(opts ...func(*Component)) *Component {
	c := &Component{
		handlers:     make(map[string]runtime.HandlerFunc),
		metadataFunc: []func(context.Context, *http.Request) metadata.MD{annotator.PassAuthResult},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.forwardResponseRewriter = func(ctx context.Context, response proto.Message) (any, error) {
		resp := make(map[string]any)
		if s, ok := response.(*status.Status); ok {
			if http.StatusText(int(s.Code)) == "" && (s.Code < 0 && s.Code > 17) {
				resp["code"] = s.Code
			}
			resp["message"] = s.Message
			resp["data"] = nil
		} else {
			if _, ok := response.(*emptypb.Empty); ok {
				resp["data"] = nil
			} else {
				resp["data"] = response
			}
			resp["message"] = ""
		}
		return resp, nil
	}
	return c
}

type Component struct {
	app.Component
	cancel                  context.CancelFunc
	registers               []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	handlers                map[string]runtime.HandlerFunc
	gwServer                *http.Server
	forwardResponseRewriter runtime.ForwardResponseRewriter
	errorHandler            runtime.ErrorHandlerFunc
	middleware              []runtime.Middleware
	metadataFunc            []func(context.Context, *http.Request) metadata.MD
	queryParser             runtime.QueryParameterParser
	ignoreLogPaths          []string
}

func (c *Component) IgnoreLogPaths(paths ...string) {
	c.ignoreLogPaths = append(c.ignoreLogPaths, paths...)
}

func (c *Component) SetForwardResponseRewriter(f runtime.ForwardResponseRewriter) {
	c.forwardResponseRewriter = f
}

func (c *Component) SetErrorHandler(f runtime.ErrorHandlerFunc) {
	c.errorHandler = f
}

func (c *Component) SetQueryParser(p runtime.QueryParameterParser) {
	c.queryParser = p
}

func (c *Component) Register(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	c.registers = append(c.registers, registers...)
}

func (c *Component) RegisterCustomerHandler(method string, path string, handler runtime.HandlerFunc) {
	c.handlers[fmt.Sprintf("%s|%s", method, path)] = handler
}

func (c *Component) RegisterMiddleware(m ...runtime.Middleware) {
	c.middleware = append(c.middleware, m...)
}

func (c *Component) RegisterMetadataFunc(m ...func(context.Context, *http.Request) metadata.MD) {
	c.metadataFunc = append(c.metadataFunc, m...)
}

func (c *Component) Init() error {
	return app.Bind[*Component](c)
}

// Start 启动服务并阻塞, 框架一般会将这个方法作为协程调用, 报错应打日志记录
func (c *Component) Start() {
	if len(c.registers) <= 0 {
		slog.Warn("no gateway register")
		return
	}
	conf := viper.Sub("grpcgatewayx")
	if conf == nil {
		slog.Error("no grpc gateway config")
		return
	}

	var ctx context.Context
	ctx, c.cancel = context.WithCancel(context.Background())

	endpoint := conf.GetString("endpoint")
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32), grpc.MaxCallSendMsgSize(math.MaxInt32)),
	)
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

	options := make([]runtime.ServeMuxOption, 0, 5)
	if c.forwardResponseRewriter != nil {
		options = append(options, runtime.WithForwardResponseRewriter(c.forwardResponseRewriter))
	}
	if c.errorHandler != nil {
		options = append(options, runtime.WithErrorHandler(c.errorHandler))
	}

	authMiddleware := middleware.NewJwtAuthMiddlewareBuilder()
	err = app.Bind[*middleware.JwtAuthMiddlewareBuilder](func() *middleware.JwtAuthMiddlewareBuilder {
		return authMiddleware
	})
	if err != nil {
		return
	}

	middlewares := append([]runtime.Middleware{
		middleware.GenLogReqAndRespMiddleware(c.ignoreLogPaths),
		authMiddleware.Build,
	}, c.middleware...)
	options = append(options, runtime.WithMiddlewares(middlewares...))
	options = append(options, runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
		result := make(metadata.MD)
		for _, v := range c.metadataFunc {
			md := v(ctx, request)
			if md == nil {
				continue
			}
			for k, v := range md {
				result[k] = v
			}
		}
		return result
	}))
	if c.queryParser != nil {
		options = append(options, runtime.SetQueryParameterParser(c.queryParser))
	}
	mux := runtime.NewServeMux(options...)
	for _, register := range c.registers {
		err := register(ctx, mux, conn)
		if err != nil {
			slog.Error("Failed to register gateway", "err", err)
		}
	}

	for key, handler := range c.handlers {
		info := strings.Split(key, "|")
		err := mux.HandlePath(info[0], info[1], handler)
		if err != nil {
			panic(err)
		}
	}

	withCors := cors.AllowAll().Handler(mux)

	c.gwServer = &http.Server{
		Addr:    conf.GetString("addr"),
		Handler: withCors,
	}

	slog.Info("Serving gRPC-Gateway on " + conf.GetString("addr"))
	slog.Error("grpc gateway server shutdown", "err", c.gwServer.ListenAndServe())
}

// Stop 停止服务并阻塞, 报错应打日志记录
func (c *Component) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	c.gwServer.Shutdown(ctx)
	if c.cancel != nil {
		c.cancel()
	}
}
