package grpcx

import (
	"fmt"
	"net"
	"strings"

	"github.com/goslacker/slacker/core/grpcx/interceptor"
	"github.com/goslacker/slacker/core/registry"
	"github.com/goslacker/slacker/core/trace"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type TraceConfig struct {
	Type     trace.TraceType `mapstructure:"type"`
	Endpoint string          `mapstructure:"endpoint"`
}

type RegistryConfig struct {
	Type      string   `mapstructure:"type"`
	Endpoints []string `mapstructure:"endpoints"`
}

type GrpcServerBuilder struct {
	ServerOptions    []grpc.ServerOption           //grpc其他配置
	ServiceRegisters []func(grpc.ServiceRegistrar) // 服务注册器
	HealthCheck      bool                          // 是否开启健康检查
	Reflection       bool                          // 是否开启反射
	PprofPort        int                           // pprof端口,如果为1-65535,则开启pprof
	Network          string                        // 网络名称
	Addr             string                        // 监听地址
	TraceConfig      *TraceConfig                  // 链路追踪配置
	RegistryConfig   *RegistryConfig               // 服务注册配置
	RegistryDriver   registry.Driver               // 服务注册驱动
}

func (c *GrpcServerBuilder) Register(registers ...func(grpc.ServiceRegistrar)) {
	c.ServiceRegisters = append(c.ServiceRegisters, registers...)
}

func (c *GrpcServerBuilder) SetServerOptions(opts ...grpc.ServerOption) {
	c.ServerOptions = append(c.ServerOptions, opts...)
}

func (c *GrpcServerBuilder) Build() (server *Server, err error) {
	// 通过配置的监听地址和网络名称尝试获取真实的本机ip地址
	addr, err := DetectAddr(c.Addr, c.Network)
	if err != nil {
		err = fmt.Errorf("detect addr failed(network: %s, addr: %s): %w", c.Network, c.Addr, err)
		return
	}

	// 初始化链路追踪
	if c.TraceConfig != nil && c.TraceConfig.Endpoint != "" {
		c.ServerOptions = append(c.ServerOptions, grpc.UnaryInterceptor(trace.UnaryTraceServerInterceptor),
			grpc.StreamInterceptor(trace.StreamTraceServerInterceptor))
	}

	c.ServerOptions = append(c.ServerOptions, grpc.UnaryInterceptor(interceptor.UnaryErrorInterceptor),
		grpc.StreamInterceptor(interceptor.StreamErrorInterceptor))
	c.ServerOptions = append(c.ServerOptions, grpc.UnaryInterceptor(interceptor.UnaryValidateInterceptor),
		grpc.StreamInterceptor(interceptor.StreamValidateInterceptor))

	server = &Server{
		pprofPort: c.PprofPort,
		addr:      c.Addr,
	}
	server.Server = grpc.NewServer(c.ServerOptions...)
	// 注册服务
	for _, register := range c.ServiceRegisters {
		register(server.Server)
	}

	// 初始化链路追踪
	if c.TraceConfig != nil && c.TraceConfig.Endpoint != "" {
		svrMap := server.Server.GetServiceInfo()
		var providerDefer func()
		providerDefer, err = trace.InitTraceProviders(c.TraceConfig.Type, c.TraceConfig.Endpoint, maps.Keys(svrMap), addr)
		if err != nil {
			err = fmt.Errorf("init trace providers failed: %w", err)
			return
		}
		server.defers = append(server.defers, providerDefer)
	}

	var healthCheckServer *health.Server
	// 注册健康检查服务
	if c.HealthCheck {
		healthCheckServer = health.NewServer()
		healthgrpc.RegisterHealthServer(server.Server, healthCheckServer)
	}

	// 注册反射服务
	if c.Reflection {
		reflection.Register(server.Server)
	}

	// 初始化服务注册
	if c.RegistryConfig != nil {
		server.registrar = registry.NewDefaultRegistrar(addr, c.RegistryDriver)
	}

	return
}

func DetectAddr(oriAddr string, network string) (realAddr string, err error) {
	arr := strings.Split(oriAddr, ":")
	if len(arr) != 2 {
		err = fmt.Errorf("invalid oriAddr: %s, forget port?", oriAddr)
		return
	}
	hostname := arr[0]
	if hostname != "" && hostname != "localhost" && hostname != "127.0.0.1" && hostname != "0.0.0.0" {
		realAddr = oriAddr
		return
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		err = fmt.Errorf("get interfaces failed: %w", err)
		return
	}

	var ip string
	for _, iface := range interfaces {
		if network != "" && iface.Name != network {
			continue
		}
		var addrs []net.Addr
		addrs, err = iface.Addrs()
		if err != nil {
			err = fmt.Errorf("get iface %s's addrs failed: %w", iface.Name, err)
			return
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil { //目前只持支ipv4
				ip = ipnet.IP.String()
				break
			}
		}
	}

	if ip == "" {
		err = fmt.Errorf("no valid address found")
		return
	}

	realAddr = ip + ":" + arr[1]

	return
}
