package main

import (
	"flag"
	"fmt"

	"richcode.cc/dex/consumer/consumer"
	"richcode.cc/dex/consumer/internal/config"
	"richcode.cc/dex/consumer/internal/logic/slot"
	"richcode.cc/dex/consumer/internal/server"
	"richcode.cc/dex/consumer/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/consumer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// 此s是consumer服务的server
	// 创建service组：实现多个服务的统一管理
	group := service.NewServiceGroup()
	defer group.Stop()

	// 此s是consumer服务的server
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		consumer.RegisterConsumerServer(grpcServer, server.NewConsumerServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	group.Add(s)

	{
		// 增加生产者服务：获取最新的slot
		// 依赖注入：把 ctx 作为参数传进去，而不是让这些服务自己创建
		// group.Add(slot.NewSlotAndSlotWsService(ctx)) // 目前NewSlotAndSlotWsService没啥作用，可以跳过它直接使用SlotWsService，因此先注释掉
		slotService := slot.NewSlotService(ctx)
		group.Add(slot.NewSlotWsService(slotService))
	}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	group.Start()
}
