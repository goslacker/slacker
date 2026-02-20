package grpcx

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/goslacker/slacker/core/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

type Server struct {
	*grpc.Server                         // grpc服务器
	healthCheckServer *health.Server     // 健康检查服务
	registrar         registry.Registrar //服务中心注册者
	defers            []func()           // 延迟执行函数
	pprofPort         int                // pprof端口,如果为1-65535,则开启pprof
	pprofHttpServer   *http.Server       // pprof http 服务器
	addr              string             //grpc服务端口
}

func (s *Server) Start(ctx context.Context) {
	if s.pprofPort > 0 {
		go func() {
			mux := http.NewServeMux()
			prefix := ""
			godebug := os.Getenv("GODEBUG")
			if strings.Contains(godebug, "httpmuxgo121=1") {
				prefix = "GET "
			}
			mux.HandleFunc(prefix+"/debug/pprof/", pprof.Index)
			mux.HandleFunc(prefix+"/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc(prefix+"/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc(prefix+"/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc(prefix+"/debug/pprof/trace", pprof.Trace)
			s.pprofHttpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", s.pprofPort),
				Handler: mux,
			}
			err := s.pprofHttpServer.ListenAndServe()
			if err != nil {
				slog.Error("pprof start failed", "error", err)
			}
		}()
	}

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(fmt.Errorf("tcp listen port failed: %w", err))
	}

	defer func() {
		if s.pprofHttpServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.pprofHttpServer.Shutdown(ctx)
		}
		for i := len(s.defers) - 1; i >= 0; i-- {
			s.defers[i]()
		}
	}()

	slog.Info("Serving gRPC on " + s.addr)
	err = s.Server.Serve(lis)
	if err != nil {
		slog.Error("grpc server shutdown", "error", err)
	} else {
		slog.Info("grpc server shutdown")
	}
}

func (s *Server) Stop(ctx context.Context) {
	s.Server.Stop()
}
