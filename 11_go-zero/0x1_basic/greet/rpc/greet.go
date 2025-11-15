// greet.go - RPC 服务主入口文件
// 这个文件启动 gRPC 服务器，提供内部服务间通信的能力
// RPC 服务的作用：
// 1. 提供高性能的内部服务接口（使用 gRPC 协议）
// 2. 封装核心业务逻辑，供多个 API 服务调用
// 3. 实现服务解耦：API 层负责对外接口，RPC 层负责业务逻辑
// 4. 支持服务间通信，实现微服务架构

package main

import (
	"flag"
	"fmt"

	"greet/rpc/internal/config"
	"greet/rpc/internal/server"
	"greet/rpc/internal/svc"
	"greet/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/greet.yaml", "the config file")

// main - RPC 服务入口函数
// 启动流程：
// 1. 加载配置文件
// 2. 创建服务上下文
// 3. 创建 gRPC 服务器并注册服务
// 4. 启动服务器监听 RPC 请求（默认端口 8080）
func main() {
	flag.Parse()

	// 加载配置文件
	var c config.Config
	conf.MustLoad(*configFile, &c)
	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 创建 gRPC 服务器
	// zrpc.MustNewServer 是 go-zero 封装的 gRPC 服务器创建函数
	// 它提供了服务发现、负载均衡、限流等微服务特性
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册 Greet 服务到 gRPC 服务器
		// 这样其他服务就可以通过 gRPC 协议调用这个服务
		pb.RegisterGreetServer(grpcServer, server.NewGreetServer(ctx))

		// 在开发或测试模式下启用 gRPC 反射
		// 反射允许客户端动态发现服务的方法和参数
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动 RPC 服务器，开始监听 gRPC 请求
	// 默认监听地址：127.0.0.1:8080
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start() // 阻塞运行，直到服务器关闭
}
