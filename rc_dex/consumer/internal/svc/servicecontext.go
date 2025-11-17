package svc

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	solclient "github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/zeromicro/go-zero/core/logx"
	"richcode.cc/dex/consumer/internal/config"
)

// ServiceContext - 服务上下文结构体
// 用于在整个服务中共享配置和依赖服务
// 通过依赖注入的方式，避免在业务逻辑中直接创建依赖
// 目前只包含应用配置，后续可扩展为包含数据库连接、RPC 客户端等。
type ServiceContext struct {
	Config config.Config

	/* Solana RPC 客户端 */
	solClientLock  sync.Mutex
	solClientIndex int
	solClient      *solclient.Client
	solClients     []*solclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func NewSolServiceContext(c config.Config) *ServiceContext {
	logx.MustSetup(c.Log)

	logx.Infof("newSolServiceContext: config:%#v", c)

	var solClients []*solclient.Client
	for _, node := range c.Sol.NodeUrl {
		solClients = append(solClients, client.New(rpc.WithEndpoint(node), rpc.WithHTTPClient(&http.Client{
			Timeout: 10 * time.Second,
		})))
	}
	fmt.Println("solClients: ", c.Sol.NodeUrl)
	return &ServiceContext{
		Config:     c,
		solClients: solClients,
	}
}

// 通过轮询的方式将请求分散到不同的客户端进行处理
func (sc *ServiceContext) GetSolClient() *client.Client {
	sc.solClientLock.Lock()
	defer sc.solClientLock.Unlock()
	sc.solClientIndex++
	index := sc.solClientIndex % len(sc.solClients)
	sc.solClient = sc.solClients[index]
	return sc.solClients[index]
}
