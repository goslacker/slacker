package grpcgatewayx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func WithRegisters(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) func(*Component) {
	return func(c *Component) {
		c.registers = registers
	}
}

func NewComponent(opts ...func(*Component)) *Component {
	c := &Component{
		handlers: make(map[string]runtime.HandlerFunc),
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
}

func (c *Component) SetForwardResponseRewriter(f runtime.ForwardResponseRewriter) {
	c.forwardResponseRewriter = f
}

func (c *Component) SetErrorHandler(f runtime.ErrorHandlerFunc) {
	c.errorHandler = f
}

func (c *Component) Register(registers ...func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	c.registers = append(c.registers, registers...)
}

func (c *Component) RegisterCustomerHandler(method string, path string, handler runtime.HandlerFunc) {
	c.handlers[fmt.Sprintf("%s|%s", method, path)] = handler
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

	options := make([]runtime.ServeMuxOption, 0, 5)
	if c.forwardResponseRewriter != nil {
		options = append(options, runtime.WithForwardResponseRewriter(c.forwardResponseRewriter))
	}
	if c.errorHandler != nil {
		options = append(options, runtime.WithErrorHandler(c.errorHandler))
	}
	options = append(options, runtime.WithMiddlewares(LogReqAndRespMiddleware))
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

type ResponseLogger struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (r *ResponseLogger) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseLogger) Write(body []byte) (int, error) {
	r.Body = body
	return r.ResponseWriter.Write(body)
}

// LogReqAndRespMiddleware 是一个中间件，用于打印请求和响应的日志
func LogReqAndRespMiddleware(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		ctx := metadata.NewOutgoingContext(
			r.Context(),
			metadata.MD{
				"grpcgateway-test": []string{"test1"},
			},
		)
		r = r.WithContext(ctx)

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		w = &ResponseLogger{ResponseWriter: w}

		// 调用下一个处理函数
		next(w, r, pathParams)

		rMap := map[string]any{
			"method": r.Method,
			"url":    r.URL.String(),
			"header": r.Header,
		}
		{
			var b map[string]any
			err = json.Unmarshal(reqBody, &b)
			if err != nil {
				rMap["body"] = string(reqBody)
			} else {
				rMap["body"] = b
			}
		}

		respMap := map[string]any{
			"status": w.(*ResponseLogger).StatusCode,
			"header": w.Header(),
		}
		{
			var b map[string]any
			err = json.Unmarshal(w.(*ResponseLogger).Body, &b)
			if err != nil {
				respMap["body"] = string(w.(*ResponseLogger).Body)
			} else {
				respMap["body"] = b
			}
		}

		slog.Debug(
			"request log",
			"req",
			rMap,
			"resp",
			respMap,
		)
	}
}
