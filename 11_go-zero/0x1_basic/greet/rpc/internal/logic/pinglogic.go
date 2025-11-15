// pinglogic.go - RPC 服务的业务逻辑层
// 这个文件实现了 RPC 服务的核心业务逻辑
// 当 API 服务调用 RPC 服务时，实际执行的是这里的业务逻辑

package logic

import (
	"context"

	"greet/rpc/internal/svc"
	"greet/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// PingLogic - RPC 服务业务逻辑结构体
type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewPingLogic - 创建 PingLogic 实例
func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Ping - RPC 服务的业务逻辑实现
// 这个方法会被 API 服务通过 gRPC 调用
// 可以在这里实现：
//   - 数据库操作
//   - 复杂的业务计算
//   - 调用其他服务
//   - 缓存操作等
//
// 参数：
//   - in: 输入参数（当前为 Placeholder，可扩展为具体业务参数）
// 返回：
//   - *pb.Placeholder: 返回结果（当前为 Placeholder，可扩展为具体业务结果）
//   - error: 错误信息
func (l *PingLogic) Ping(in *pb.Placeholder) (*pb.Placeholder, error) {
	// todo: add your logic here and delete this line
	// 在这里添加实际的业务逻辑，例如：
	// - 查询数据库
	// - 处理业务规则
	// - 调用其他服务等

	return &pb.Placeholder{}, nil
}
