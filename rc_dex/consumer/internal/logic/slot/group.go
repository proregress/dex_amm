/**
分析 SlotAndSlotWsService 是否必要：
结论：SlotAndSlotWsService 是多余的，可以直接使用 SlotWsService。
原因：
Start() 只是转发调用
Stop() 已通过嵌入 *SlotService 获得
没有额外功能

这里这个文件暂且保留，只在consumer.go中进行修改，使其不调用这里的方法

后续如果需要（比如要组合多个服务）再改回来
**/

package slot

import "richcode.cc/dex/consumer/internal/svc"

// SlotAndSlotWsService 和 SlotWsService 共享同一个 SlotService 实例
// 设计目的：共享状态，统一管理，避免数据不一致
// 这是共享模式（Shared Instance Pattern）
type SlotAndSlotWsService struct {
	*SlotService // 这里通过嵌入 *SlotService 获得了它的 Stop() 方法
	Ws           *SlotWsService
}

func NewSlotAndSlotWsService(sc *svc.ServiceContext, slotChannel chan uint64) *SlotAndSlotWsService {
	slotService := NewSlotService(sc, slotChannel)
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
