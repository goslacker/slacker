package grpcx

import (
	"fmt"
	"net"
	"strings"

	"github.com/goslacker/slacker/core/registry"
	"github.com/goslacker/slacker/core/trace"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type traceConfig struct {
	Type     trace.TraceType `mapstructure:"type"`
	Endpoint string          `mapstructure:"endpoint"`
}

type registryConfig struct {
	Type      string   `mapstructure:"type"`
	Endpoints []string `mapstructure:"endpoints"`
}

type GrpcServerBuilder struct {
	unaryServerInterceptors  []grpc.UnaryServerInterceptor  //一元拦截器
	streamServerInterceptors []grpc.StreamServerInterceptor //流拦截器
	otherServerOptions       []grpc.ServerOption            //grpc其他配置
	serviceRegisters         []func(grpc.ServiceRegistrar)  // 服务注册器
	healthCheck              bool                           // 是否开启健康检查
	reflection               bool                           // 是否开启反射
	pprofPort                int                            // pprof端口,如果为1-65535,则开启pprof
	network                  string                         // 网络名称
	addr                     string                         // 监听地址
	traceConfig              *traceConfig                   // 链路追踪配置
	registryConfig           *registryConfig                // 服务注册配置
	registryDriver           registry.Driver                // 服务注册驱动
}

func (c *GrpcServerBuilder) RegisterUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	c.unaryServerInterceptors = append(c.unaryServerInterceptors, interceptors...)
}

func (c *GrpcServerBuilder) RegisterStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) {
	c.streamServerInterceptors = append(c.streamServerInterceptors, interceptors...)
}

func (c *GrpcServerBuilder) Register(registers ...func(grpc.ServiceRegistrar)) {
	c.serviceRegisters = append(c.serviceRegisters, registers...)
}

func (c *GrpcServerBuilder) SetOtherServerOptions(opts ...grpc.ServerOption) {
	c.otherServerOptions = append(c.otherServerOptions, opts...)
}

func (c *GrpcServerBuilder) Build() (server *Server, err error) {
	server = &Server{
		pprofPort: c.pprofPort,
		addr:      c.addr,
	}
	opts := make([]grpc.ServerOption, 0, len(c.otherServerOptions)+2)
	opts = append(opts, c.otherServerOptions...)
	opts = append(opts, []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(c.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(c.streamServerInterceptors...),
	}...)
	server.Server = grpc.NewServer(opts...)
	// 注册服务
	for _, register := range c.serviceRegisters {
		register(server.Server)
	}

	var healthCheckServer *health.Server
	// 注册健康检查服务
	if c.healthCheck {
		healthCheckServer = health.NewServer()
		healthgrpc.RegisterHealthServer(server.Server, healthCheckServer)
	}

	// 注册反射服务
	if c.reflection {
		reflection.Register(server.Server)
	}

	// 通过配置的监听地址和网络名称尝试获取真实的本机ip地址
	addr, err := DetectAddr(c.addr, c.network)
	if err != nil {
		err = fmt.Errorf("detect addr failed(network: %s, addr: %s): %w", c.network, c.addr, err)
		return
	}

	// 初始化链路追踪
	if c.traceConfig != nil {
		svrMap := server.Server.GetServiceInfo()
		var providerDefer func()
		providerDefer, err = trace.InitTraceProviders(c.traceConfig.Type, c.traceConfig.Endpoint, maps.Keys(svrMap), addr)
		if err != nil {
			err = fmt.Errorf("init trace providers failed: %w", err)
			return
		}
		server.defers = append(server.defers, providerDefer)
	}

	// 初始化服务注册
	if c.registryConfig != nil {
		server.registrar = registry.NewDefaultRegistrar(addr, c.registryDriver)
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
