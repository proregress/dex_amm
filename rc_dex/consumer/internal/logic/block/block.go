package block

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"richcode.cc/dex/consumer/internal/config"
	"richcode.cc/dex/consumer/internal/svc"
	"richcode.cc/dex/pkg/constants"
)

var ErrServiceStop = errors.New("service stop")

type BlockService struct {
	Name string
	sc   *svc.ServiceContext
	/* Solana Go SDK 的客户端，用于与 Solana 区块链节点通信。
	Solana Go SDK 是 Solana 区块链的官方 Go 语言 SDK，提供了与 Solana 网络交互的工具和库。
	它包含了与 Solana 区块链节点通信的各种功能，包括获取区块数据、发送交易、查询账户信息等。
	*/
	c *client.Client
	logx.Logger
	workerPool  *ants.Pool
	ctx         context.Context
	cancel      func(err error)
	slotChannel chan uint64
	Conn        *websocket.Conn
	name        string
}

// Start implements service.Service.
func (s *BlockService) Start() {
	s.GetBlockFromHttp()
}

// Stop implements service.Service.
func (s *BlockService) Stop() {
	s.cancel(ErrServiceStop)
	if s.Conn != nil {
		/* 通过 WebSocket 发送 JSON-RPC 消息，调用 blockUnsubscribe 取消订阅（订阅 ID 为 0）
		-服务停止前需要取消订阅，避免资源泄漏
		-通知 Solana 节点停止推送区块更新
		-确保 WebSocket 连接正确关闭
		*/
		err := s.Conn.WriteMessage(websocket.TextMessage, []byte("{\"id\":1,\"jsonrpc\":\"2.0\",\"method\": \"blockUnsubscribe\", \"params\": [0]}\n"))
		if err != nil {
			s.Error("programUnsubscribe", err)
		}
		_ = s.Conn.Close()
	}
}

func NewBlockService(sc *svc.ServiceContext, name string, slotChnanel chan uint64, index int) *BlockService {
	ctx, cancel := context.WithCancelCause(context.Background())
	pool, _ := ants.NewPool(5)
	blockService := &BlockService{
		c: client.New(rpc.WithEndpoint(config.FindChainRpcByChainId(constants.SolChainIdInt)), rpc.WithHTTPClient(&http.Client{
			Timeout: 5 * time.Second,
		})),
		sc:          sc,
		name:        name,
		workerPool:  pool,
		slotChannel: slotChnanel,
		Logger:      logx.WithContext(context.Background()).WithFields(logx.Field("service", fmt.Sprintf("%s-%v", name, index))),
		ctx:         ctx,
		cancel:      cancel,
	}
	return blockService
}

/*基于HTTP接口，通过传入slot编号，获取区块数据*/
func (s *BlockService) GetBlockFromHttp() {
	fmt.Print("block service is started")
	for {
		select {
		case <-s.ctx.Done():
			fmt.Print("block service is stopped")
			return
		case slot, ok := <-s.slotChannel: // 异步地从channel中获取slot数据
			if !ok {
				fmt.Print("slotChan is closed")
				return
			}
			//打印当前最新slot
			fmt.Println("consume slot is:", slot)
		}
	}
}
