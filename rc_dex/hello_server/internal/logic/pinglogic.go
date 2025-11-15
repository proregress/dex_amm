package logic

import (
	"context"

	"richcode.cc/dex/hello_server/hello"
	"richcode.cc/dex/hello_server/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Ping 方法：用于测试服务是否可用
func (l *PingLogic) Ping(in *hello.Request) (*hello.Response, error) {
	// todo: add your logic here and delete this line

	return &hello.Response{
		Msg: "hello server of logic layer",
	}, nil
}
