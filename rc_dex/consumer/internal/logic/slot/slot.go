package slot

import (
	"context"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"richcode.cc/dex/consumer/internal/svc"
)

var ErrServiceStop = errors.New("slot service stopped")

// SlotService 是slot服务的主要逻辑，它不直接启动，而是通过SlotWsService间接启动
// 它虽然叫slot服务，但它不是一个独立运行的服务（consumer是，SlotWsService是，原因：实现了 Service 接口，有真正的 Start() 实现）
// 它应该被理解为一个工具类、基础类，提供各种func给别人使用
type SlotService struct {
	Conn *websocket.Conn
	sc   *svc.ServiceContext // 服务上下文,用于获取配置和共享资源
	logx.Logger

	ctx     context.Context //生命周期控制上下文 1.控制goroutine的生命周期，用于启动与停止 2.通过done()通知goroutine退出
	cancel  func(err error)
	maxSlot uint64

	realtimeChannel chan uint64

	errChannel chan uint64
}

func NewSlotService(sc *svc.ServiceContext, slotChannel chan uint64, errChannel chan uint64) *SlotService {
	// context.WithCancelCause 返回：1. ctx：可取消的上下文 2.cancel func(error)：取消函数，可携带错误原因
	ctx, cancel := context.WithCancelCause(context.Background())
	return &SlotService{
		sc:              sc,
		Logger:          logx.WithContext(context.Background()).WithFields(logx.Field("service", "slot")),
		ctx:             ctx,    // ctx 是上下文，goroutine 通过监听 ctx.Done() 感知取消
		cancel:          cancel, // cancel函数用于取消上下文，调用后ctx.Done()会返回一个已关闭的channel，ctx.Err() 返回取消原因（通过 context.Cause(ctx) 获取）
		realtimeChannel: slotChannel,
		errChannel:      errChannel,
	}
}

// 这个Start()是空实现，因为SlotService不是一个需要启动的服务，可以删除
// 虽然是空，但也算实现了Start()方法
// 在 Go 中，只要方法签名匹配，即使方法体为空，也满足接口要求。

// 它只提供功能方法（MustConnect, ReadSlotMessage等）给其他服务使用
// 真正的服务（SlotWsService, SlotAndSlotWsService）会实现自己的 Start() 方法
// func (s *SlotService) Start() {}

func (s *SlotService) Stop() {
	s.Info("Stopping slot service")
	s.cancel(ErrServiceStop) // 取消 context，通知 goroutine 退出
	if s.Conn != nil {
		_ = s.Conn.Close() // Close() 是物理层面的关闭：立即关闭底层的 WebSocket 连接,让正在阻塞的 ReadMessage() 立即返回错误, goroutine 可以快速退出
	}
}
