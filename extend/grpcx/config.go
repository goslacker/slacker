package grpcx

import (
	"github.com/goslacker/slacker/serviceregistry/registry"
)

type Config struct {
	Trace       bool                     //是否开启链路追踪
	HealthCheck bool                     //是否开启健康检查
	Reflection  bool                     //是否开启反射服务
	Addr        string                   //服务地址
	Registry    *registry.RegistryConfig //服务注册中心配置
}
