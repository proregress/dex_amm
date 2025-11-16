package slot

import (
	"context"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"richcode.cc/dex/consumer/internal/svc"
)

var ErrServiceStop = errors.New("slot service stopped")

type SlotService struct {
	Conn *websocket.Conn
	sc   *svc.ServiceContext // 服务上下文,用于获取配置和共享资源
	logx.Logger

	ctx     context.Context //生命周期控制上下文 1.控制goroutine的生命周期，用于启动与停止 2.通过done()通知goroutine退出
	cancel  func(err error)
	maxSlot uint64
}

func NewSlotService(sc *svc.ServiceContext) *SlotService {
	// context.WithCancelCause 返回：1. ctx：可取消的上下文 2.cancel func(error)：取消函数，可携带错误原因
	ctx, cancel := context.WithCancelCause(context.Background())
	return &SlotService{
		sc:     sc,
		Logger: logx.WithContext(context.Background()).WithFields(logx.Field("service", "slot")),
		ctx:    ctx,    // ctx 是上下文，goroutine 通过监听 ctx.Done() 感知取消
		cancel: cancel, // cancel函数用于取消上下文，调用后ctx.Done()会返回一个已关闭的channel，ctx.Err() 返回取消原因（通过 context.Cause(ctx) 获取）
	}
}

func (s *SlotService) Start() {}

func (s *SlotService) Stop() {
	s.Info("Stopping slot service")
	s.cancel(ErrServiceStop)
	if s.Conn != nil {
		_ = s.Conn.Close() // Close() 是物理层面的关闭：立即关闭底层的 WebSocket 连接,让正在阻塞的 ReadMessage() 立即返回错误, goroutine 可以快速退出
	}
}
