// greet.go - 应用程序主入口文件
// 这个文件负责：
// 1. 解析命令行参数（配置文件路径）
// 2. 加载配置文件
// 3. 创建 HTTP 服务器
// 4. 初始化服务上下文
// 5. 注册路由处理器
// 6. 启动服务器监听请求

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"greet/api/internal/config"
	"greet/api/internal/handler"
	"greet/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// configFile - 命令行参数，指定配置文件路径，默认为 "etc/greet.yaml"
var configFile = flag.String("f", "etc/greet.yaml", "the config file")

// main - 程序入口函数
// 执行流程：
// 1. 解析命令行参数获取配置文件路径
// 2. 加载配置文件到 Config 结构体
// 3. 根据配置创建 REST 服务器
// 4. 创建服务上下文（包含配置和 RPC 客户端等）
// 5. 注册所有路由处理器
// 6. 启动服务器开始监听 HTTP 请求
func main() {
	flag.Parse()

	// 加载配置文件
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建 HTTP REST API 服务器实例
	// server 的类型是 *rest.Server，这是 go-zero 框架提供的 HTTP 服务器
	// rest.Server 是对标准库 net/http 的封装，提供了：
	//   - 路由管理（支持 RESTful API）
	//   - 中间件支持
	//   - 优雅关闭（Graceful Shutdown）
	//   - 请求日志和监控
	//   - CORS 支持
	//   - 限流和熔断等特性
	// MustNewServer 会根据配置文件创建服务器，如果出错会直接退出程序
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop() // 确保程序退出时优雅关闭服务器

	// 创建服务上下文，包含配置和依赖服务（如 RPC 客户端）
	ctx := svc.NewServiceContext(c)
	// 注册所有路由处理器，将路由映射到对应的处理器函数
	handler.RegisterHandlers(server, ctx)

	// 启动服务器，开始监听并处理 HTTP 请求
	// server.Start() 会：
	//   1. 根据配置文件中的 Host 和 Port（127.0.0.1:8888）启动 HTTP 服务器
	//   2. 绑定所有已注册的路由
	//   3. 开始监听 HTTP 请求
	//   4. 支持优雅关闭（收到 SIGTERM/SIGINT 信号时会优雅关闭）
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start() // 这是一个阻塞调用，会一直运行直到服务器关闭
}
