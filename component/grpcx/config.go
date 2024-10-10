package grpcx

import (
	"github.com/goslacker/slacker/core/serviceregistry/registry"
	"github.com/goslacker/slacker/core/trace"
)

type Config struct {
	HealthCheck bool                     //是否开启健康检查
	Reflection  bool                     //是否开启反射服务
	Addr        string                   //服务地址
	Trace       *trace.TraceConfig       //启链路追踪配置
	Registry    *registry.RegistryConfig //服务注册中心配置
}
