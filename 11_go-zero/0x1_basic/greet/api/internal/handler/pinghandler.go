// pinghandler.go - Ping 接口的 HTTP 处理器
// 这个文件实现了 /ping 路由的 HTTP 处理器函数
// 处理器负责：
// 1. 接收 HTTP 请求
// 2. 创建业务逻辑层实例
// 3. 调用业务逻辑处理请求
// 4. 将处理结果序列化为 JSON 并返回给客户端

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"greet/api/internal/logic"
	"greet/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// pingHandler - 创建并返回 /ping 路由的处理器函数
// 参数：
//   - svcCtx: 服务上下文，包含配置和依赖服务（如 RPC 客户端）
//
// 返回：
//   - http.HandlerFunc: 标准的 HTTP 处理器函数
//
// 处理流程：
//  1. 创建 PingLogic 业务逻辑实例（传入请求上下文和服务上下文）
//  2. 调用业务逻辑的 Ping() 方法处理请求
//  3. 如果出错，返回错误响应
//  4. 如果成功，将响应数据序列化为 JSON 并返回（HTTP 200）
func pingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 创建业务逻辑层实例，传入请求上下文和服务上下文
		l := logic.NewPingLogic(r.Context(), svcCtx)
		// 调用业务逻辑处理请求
		resp, err := l.Ping()
		if err != nil {
			// 如果处理出错，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 如果处理成功，将响应数据序列化为 JSON 并返回（HTTP 200 OK）
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func xufanHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 创建业务逻辑层实例，传入请求上下文和服务上下文
		l := logic.NewPingLogic(r.Context(), svcCtx)
		// 调用业务逻辑处理请求
		resp, err := l.Xufan()
		if err != nil {
			// 如果处理出错，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 如果处理成功，将响应数据序列化为 JSON 并返回（HTTP 200 OK）
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
