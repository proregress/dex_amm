// pinglogic.go - Ping 接口的业务逻辑层
// 这个文件实现了 /ping 接口的核心业务逻辑
// 业务逻辑层负责：
// 1. 处理业务规则和数据验证
// 2. 调用 RPC 服务或其他依赖服务
// 3. 组装响应数据
// 注意：这里还调用了 RPC 服务，但即使 RPC 服务不可用，也会返回 "pong" 响应

package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"greet/api/internal/svc"
	"greet/api/internal/types"
	"greet/rpc/greet"
)

// PingLogic - Ping 业务逻辑结构体
// 包含日志记录器、请求上下文和服务上下文
type PingLogic struct {
	logx.Logger                     // 日志记录器，用于记录日志
	ctx         context.Context     // 请求上下文，用于传递请求相关的信息（如超时控制）
	svcCtx      *svc.ServiceContext // 服务上下文，包含配置和依赖服务（如 RPC 客户端）
}

// NewPingLogic - 创建 PingLogic 实例
// 参数：
//   - ctx: 请求上下文
//   - svcCtx: 服务上下文
//
// 返回：
//   - *PingLogic: 业务逻辑实例
func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx), // 创建带上下文的日志记录器
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Ping - 处理 Ping 请求的业务逻辑
// 返回：
//   - resp: 响应数据，包含 msg 字段
//   - err: 错误信息（如果有）
//
// 处理流程：
//  1. 调用 RPC 服务的 Ping 方法（可选，用于健康检查或服务间通信）
//  2. 创建响应对象，设置 msg 为 "pong"
//  3. 返回响应数据
func (l *PingLogic) Ping() (resp *types.Resp, err error) {
	// 调用 RPC 服务的 Ping 方法（这里用于演示，实际可能用于服务间通信）
	// 注意：如果 RPC 服务不可用，这里会返回错误，但不会影响最终响应
	if _, err = l.svcCtx.GreetRpc.Ping(l.ctx, new(greet.Placeholder)); err != nil {
		return
	}

	// 创建响应对象并设置消息内容
	resp = new(types.Resp)
	resp.Msg = "pong of api layer"

	return
}

func (l *PingLogic) Xufan() (resp *types.Resp, err error) {
	// 调用 RPC 服务的 Ping 方法（这里用于演示，实际可能用于服务间通信）
	// 注意：如果 RPC 服务不可用，这里会返回错误，但不会影响最终响应
	if _, err = l.svcCtx.GreetRpc.Ping(l.ctx, new(greet.Placeholder)); err != nil {
		return
	}

	// 创建响应对象并设置消息内容
	resp = new(types.Resp)
	resp.Msg = "xufan of api layer"

	return
}
