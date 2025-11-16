package slot

import "richcode.cc/dex/consumer/internal/svc"

type SlotAndSlotWsService struct {
	*SlotService // 这里通过嵌入 *SlotService 获得了它的 Start() 和 Stop() 方法：
	Ws           *SlotWsService
}

func NewSlotAndSlotWsService(sc *svc.ServiceContext) *SlotAndSlotWsService {
	slotService := NewSlotService(sc)
	return &SlotAndSlotWsService{
		SlotService: slotService,
		Ws:          NewSlotWsService(slotService),
	}
}

// Start 启动 SlotAndSlotWsService
// 这里需要显式实现 Start() 方法，因为嵌入的 SlotService.Start() 是空实现
// 我们需要调用 Ws.Start() 来启动 WebSocket 服务
func (s *SlotAndSlotWsService) Start() {
	s.Ws.Start()
}
