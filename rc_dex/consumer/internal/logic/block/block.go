package block

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"richcode.cc/dex/consumer/internal/config"
	"richcode.cc/dex/consumer/internal/svc"
	"richcode.cc/dex/model/solmodel"
	"richcode.cc/dex/pkg/constants"
)

var ErrServiceStop = errors.New("service stop")

var ErrUnknowProgram = errors.New("unknow program")

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
	ctx := s.ctx
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
			// 再创建一个单独的协程，专门用来处理区块数据
			// RunSafe：异步执行，不会阻塞当前协程，继续执行下一个
			// RunSafe和GoSafe的区别？？？
			threading.RunSafe(func() {
				s.ProcessBlock(ctx, int64(slot)) // 直接调用的话是同步执行，会阻塞在这一步，直到处理完才会继续下一个
			})
		}
	}
}

func (s *BlockService) ProcessBlock(ctx context.Context, slot int64) {
	// 监听
	beginTime := time.Now()
	if slot == 0 {
		return
	}

	// 创建block对象
	block := &solmodel.Block{
		Slot: slot,
	}

	blockInfo, err := GetSolBlockInfoDelay(s.sc.GetSolClient(), ctx, uint64(slot))
	if err != nil || blockInfo == nil {
		fmt.Println("err :", err)
		return
	}
	// 从上面拿到的blockInfo将信息设置进block对象中
	if blockInfo.BlockTime != nil {
		block.BlockTime = *blockInfo.BlockTime
		blockTime := blockInfo.BlockTime.Format("2006-01-02 15:04:05")
		s.Infof("processBlock:%v getBlockInfo blockTime: %v,cur: %v, dur: %v, queue size: %v", slot, blockTime, time.Now().Format("15:04:05"), time.Since(beginTime), len(s.slotChannel))
	} else {
		s.Infof("processBlock:%v getBlockInfo blockTime is nil,cur: %v, dur: %v, queue size: %v", slot, time.Now().Format("15:04:05"), time.Since(beginTime), len(s.slotChannel))
	}

	if blockInfo.BlockHeight != nil {
		block.BlockHeight = *blockInfo.BlockHeight
	}
	block.Status = constants.BlockProcessed

	// TODO: 获取 sol 价格
	block.SolPrice = 0

	// 通过slice组件遍历transactions，拿到每一个交易对象tx
	slice.ForEach(blockInfo.Transactions, func(index int, tx client.BlockTransaction) {
		DecodeTx(&tx)
	})

	// 将block对象插入数据库
	err = s.sc.BlockModel.Insert(ctx, block)
	if err != nil {
		s.Error("insert block error", err)
	}
}

func DecodeTx(tx *client.BlockTransaction) {
	if tx == nil {
		return
	}

	// Instructions是交易中的所有指令，遍历每一个指令inst
	for i := range tx.Transaction.Message.Instructions {
		inst := &tx.Transaction.Message.Instructions[i]
		err := DecodeInstruction(inst, tx)
		if err != nil {
			return
		}
	}
}

func DecodeInstruction(inst *types.CompiledInstruction, tx *client.BlockTransaction) (err error) {
	if inst == nil {
		return errors.New("instruction is null")
	}

	if len(tx.AccountKeys) == 0 {
		return errors.New("account keys is empty")
	}

	// AccountKeys是当前交易中的所有指令的账户id
	// ProgramIDIndex是当前指令的账户id索引
	program := tx.AccountKeys[inst.ProgramIDIndex].String()
	// 根据programID，调用不同的解码函数

	// 过滤
	switch program {
	case ProgramStrPumpFun:
		return DecodePumpFunInstruction(inst, tx)
	case ProgramStrPumpFunAMM:
		return DecodePumpFunAMMInstruction(inst, tx)
	case ProgramStrRaydium:
		return DecodeRaydiumInstruction(inst, tx)
	default:
		return ErrUnknowProgram
	}
}
