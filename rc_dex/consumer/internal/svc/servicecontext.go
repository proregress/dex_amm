package svc

import "richcode.cc/dex/consumer/internal/config"

// ServiceContext - 服务上下文结构体
// 用于在整个服务中共享配置和依赖服务
// 通过依赖注入的方式，避免在业务逻辑中直接创建依赖
// 目前只包含应用配置，后续可扩展为包含数据库连接、RPC 客户端等。
type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
