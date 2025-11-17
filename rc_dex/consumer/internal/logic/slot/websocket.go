package slot

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/threading"
)

type SlotWsService struct {
	*SlotService
}

func NewSlotWsService(slotService *SlotService) *SlotWsService {
	return &SlotWsService{
		SlotService: slotService,
	}
}

/*
重要‼️
作用1:注册关闭监听器：程序关闭时调用 s.cancel() 取消 context，通知 goroutine 退出
作用2:启动 WebSocket 服务：调用 s.SlotWs() 建立连接并开始读取消息
*/
func (s *SlotWsService) Start() {
	// 是 go-zero 提供的关闭监听器注册函数。当程序收到关闭信号（如 SIGTERM、SIGINT）时，会执行注册的回调，用于优雅关闭。
	proc.AddShutdownListener(func() {
		s.Infof("SlotWsService : ShutdownListener")
		s.cancel(errors.New("close slot websocket service"))
	})

	// 等价于s.SlotService.SlotWsServiceStart()，因为SlotWsService嵌入了SlotService，【方法提升】
	s.SlotWsServiceStart()
}

/*
作用:启动 WebSocket 服务：调用 s.MustConnect() 建立连接并开始读取消息
*/
func (s *SlotService) SlotWsServiceStart() {
	s.MustConnect()
	s.Infof("SlotWs: MustConnect success")

	threading.GoSafe(func() {
		// 开始消费增量
		for {
			select {
			case <-s.ctx.Done(): // 上下文是否已经关闭？
				s.Info("slotWs stop succeed") // 已关闭：退出循环
				return
			default:
			}
			s.ReadSlotMessage() // 没关闭：不断读取消息
		}
	})
}

/*
读取消息
*/
func (s *SlotService) ReadSlotMessage() {
	// defer func是Go 的 panic 恢复机制，用于捕获并处理 ReadSlotMessage() 中可能发生的 panic。
	// 相当于try-catch-finally中的finally块，无论是否发生panic，都会执行。
	defer func() {
		cause := recover() // 捕获 panic，返回 panic 的值；无 panic 时返回 nil；
		if cause != nil {  // cause不为空，说明有Panic发生，需要恢复
			s.Errorf("ReadSlotMessage: panic: %v", cause)
			s.MustConnect()
		}
	}()
	// ReadMessage() 会阻塞，因为：
	// 1. 它在等待网络数据
	// 2. 如果对方没发消息，它会一直等待
	// 3. 在等待期间，这行代码不会返回，后面的代码不会执行
	_, message, err := s.Conn.ReadMessage() // 读取websocket连接中的消息
	if err != nil {
		s.Errorf("ReadSlotMessage: ReadMessage err: %v", err)
		// 如果错误是连接关闭或管道破裂，说明是异常断开，需要重新连接
		if strings.Contains(err.Error(), "close") {
			s.MustConnect()
		}
		if strings.Contains(err.Error(), "broken pipe") {
			s.MustConnect()
		}
		return
	}

	// 声明slot的响应对象
	var resp SlotResp
	// 通过json反序列化，将数据解析道resp对象中
	err = json.Unmarshal(message, &resp)
	if err != nil {
		s.Errorf("ReadSlotMessage: Unmarshal err: %v", err)
		return
	}
	// 如果slot为0，说明是无效的响应？，不需要处理
	if resp.Params.Result.Slot == 0 {
		return
	}
	// slot不为0，需要处理
	s.maxSlot = resp.Params.Result.Slot
	// 往channel中写入数据
	s.realtimeChannel <- s.maxSlot // 若channel中写入了缓冲区大小的数据，比如缓冲区大小为5，这里写入了5个，同时没有被消费，则会阻塞，无法再写入数据
	s.Infof("ReadSlotMessage: message: %v", string(message))
	fmt.Println("lastest slot: ", s.maxSlot)
}

/*
连接websocket
*/
func (s *SlotService) MustConnect() {
	dialer := websocket.DefaultDialer
	for {
		s.Infof("MustConnect:slot ws url: %v", s.sc.Config.Sol.WSUrl)
		dialer.HandshakeTimeout = time.Second * 5
		// c 是创建出来的websocket连接对象
		c, _, err := dialer.Dial(s.sc.Config.Sol.WSUrl, nil) // WSUrl来自yaml文件配置的helius的websocket地址
		if err != nil {
			s.Errorf("MustConnect:slot ws Dial err: %v", err)
		} else {
			s.Conn = c
			for i := 0; i < 10; i++ { // 重试
				// 发送订阅请求
				err = c.WriteMessage(websocket.TextMessage, []byte("{\"id\":1,\"jsonrpc\":\"2.0\",\"method\": \"slotSubscribe\"}\n"))
				if err != nil {
					s.Error("slot ws slotSubscribe err: %v", err)
				} else {
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// slot的响应对, 根据helius的websocket响应API结构定义的
type SlotResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Slot   uint64 `json:"slot"`
			Parent uint64 `json:"parent"`
			Root   uint64 `json:"root"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}
