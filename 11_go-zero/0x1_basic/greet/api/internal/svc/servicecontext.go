// servicecontext.go - 服务上下文定义文件
// 这个文件定义了 ServiceContext，用于在整个服务中共享配置和依赖服务
// ServiceContext 是一个依赖注入容器，包含：
// 1. 应用配置
// 2. RPC 客户端（用于调用其他微服务）
// 3. 数据库连接（如果有）
// 4. 其他共享资源

package svc

import (
	"greet/api/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
	"greet/rpc/greet"
)

// ServiceContext - 服务上下文结构体
// 用于在整个服务中共享配置和依赖服务
// 通过依赖注入的方式，避免在业务逻辑中直接创建依赖
type ServiceContext struct {
	Config   config.Config  // 应用配置
	GreetRpc greet.Greet    // Greet RPC 服务客户端，用于调用 RPC 服务
}

// NewServiceContext - 创建服务上下文实例
// 参数：
//   - c: 应用配置
// 返回：
//   - *ServiceContext: 服务上下文实例
//
// 初始化流程：
//   1. 创建 RPC 客户端连接到 RPC 服务（127.0.0.1:8080）
//   2. 创建 Greet RPC 服务客户端
//   3. 返回包含配置和 RPC 客户端的服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 RPC 客户端，连接到 RPC 服务
	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Target: "127.0.0.1:8080", // RPC 服务地址
	})
	return &ServiceContext{
		Config:   c,                      // 保存应用配置
		GreetRpc: greet.NewGreet(client), // 创建 Greet RPC 服务客户端
	}
}
