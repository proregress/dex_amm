# Go-Zero 微服务项目说明文档

## 目录
- [项目概述](#项目概述)
- [架构设计](#架构设计)
- [文件结构说明](#文件结构说明)
- [请求流程详解](#请求流程详解)
- [核心概念说明](#核心概念说明)
- [配置文件说明](#配置文件说明)
- [启动和测试](#启动和测试)

---

## 项目概述

这是一个基于 go-zero 框架的微服务项目，采用分层架构设计：
- **API 服务** (`greet/api`): 提供 HTTP REST API，面向外部客户端
- **RPC 服务** (`greet/rpc`): 提供 gRPC 服务，封装核心业务逻辑

### 技术栈
- **框架**: go-zero v1.9.2
- **API 协议**: HTTP/1.1 (RESTful)
- **RPC 协议**: gRPC (HTTP/2)
- **语言**: Go 1.24.7

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                   客户端（浏览器/App）                    │
│              curl -i http://127.0.0.1:8888/ping          │
└─────────────────────────────────────────────────────────┘
                          ↓ HTTP REST API
┌─────────────────────────────────────────────────────────┐
│              API 服务 (greet/api)                        │
│              - 端口: 8888                                 │
│              - 协议: HTTP/1.1                            │
│              - 作用: 对外提供 RESTful API                 │
│              - 职责: 接口适配、参数校验、HTTP 处理        │
└─────────────────────────────────────────────────────────┘
                          ↓ gRPC 调用
┌─────────────────────────────────────────────────────────┐
│              RPC 服务 (greet/rpc)                        │
│              - 端口: 8080                                │
│              - 协议: gRPC (HTTP/2)                       │
│              - 作用: 内部业务逻辑服务                     │
│              - 职责: 业务逻辑、数据处理、数据库操作        │
└─────────────────────────────────────────────────────────┘
```

### 架构特点

1. **分层设计**
   - API 层：负责对外接口和协议转换
   - RPC 层：负责核心业务逻辑

2. **服务解耦**
   - API 服务和 RPC 服务可以独立部署和扩展
   - 多个 API 服务可以共享同一个 RPC 服务

3. **性能优化**
   - 外部使用 HTTP REST API（易调试、易集成）
   - 内部使用 gRPC（高性能、支持流式传输）

---

## 文件结构说明

### 项目目录结构

```
greet/
├── api/                          # API 服务目录
│   ├── etc/
│   │   └── greet.yaml           # API 服务配置文件
│   ├── greet.api                # API 接口定义文件
│   ├── greet.go                 # API 服务主入口
│   └── internal/
│       ├── config/
│       │   └── config.go        # 配置结构体定义
│       ├── handler/
│       │   ├── routes.go        # 路由注册文件
│       │   └── pinghandler.go   # Ping 接口处理器
│       ├── logic/
│       │   └── pinglogic.go     # Ping 业务逻辑
│       ├── svc/
│       │   └── servicecontext.go # 服务上下文（依赖注入）
│       └── types/
│           └── types.go         # 类型定义
├── rpc/                          # RPC 服务目录
│   ├── etc/
│   │   └── greet.yaml           # RPC 服务配置文件
│   ├── greet.proto              # gRPC 服务定义文件
│   ├── greet.go                 # RPC 服务主入口
│   ├── greet/
│   │   └── greet.go             # RPC 客户端代码
│   ├── pb/                      # 自动生成的 protobuf 代码
│   │   ├── greet.pb.go
│   │   └── greet_grpc.pb.go
│   └── internal/
│       ├── config/
│       │   └── config.go        # RPC 配置结构体
│       ├── logic/
│       │   └── pinglogic.go     # RPC 业务逻辑
│       ├── server/
│       │   └── greetserver.go   # RPC 服务实现
│       └── svc/
│           └── servicecontext.go # RPC 服务上下文
├── go.mod                        # Go 模块依赖文件
└── go.sum                        # Go 模块校验文件
```

### 核心文件说明

#### API 服务文件

| 文件 | 作用 | 说明 |
|------|------|------|
| `greet.api` | API 定义文件 | 定义 REST API 接口规范，go-zero 根据此文件生成代码 |
| `greet.go` | 主入口文件 | 启动 HTTP 服务器，注册路由，监听请求 |
| `etc/greet.yaml` | 配置文件 | 定义服务端口、主机地址、日志配置等 |
| `internal/config/config.go` | 配置结构体 | 定义配置数据结构 |
| `internal/handler/routes.go` | 路由注册 | 将路由映射到对应的处理器 |
| `internal/handler/pinghandler.go` | HTTP 处理器 | 接收 HTTP 请求，调用业务逻辑，返回响应 |
| `internal/logic/pinglogic.go` | 业务逻辑层 | 处理业务规则，调用 RPC 服务，组装响应数据 |
| `internal/svc/servicecontext.go` | 服务上下文 | 依赖注入容器，包含配置和 RPC 客户端 |
| `internal/types/types.go` | 类型定义 | 定义请求和响应的数据结构 |

#### RPC 服务文件

| 文件 | 作用 | 说明 |
|------|------|------|
| `greet.proto` | gRPC 服务定义 | 定义 RPC 服务接口和消息格式 |
| `greet.go` | RPC 服务入口 | 启动 gRPC 服务器，注册服务，监听 RPC 请求 |
| `etc/greet.yaml` | RPC 配置文件 | 定义 RPC 服务端口、日志配置等 |
| `internal/server/greetserver.go` | RPC 服务实现 | 实现 proto 中定义的 RPC 方法 |
| `internal/logic/pinglogic.go` | RPC 业务逻辑 | 实现核心业务逻辑（数据库操作、业务计算等） |
| `pb/*.pb.go` | 自动生成代码 | 根据 proto 文件生成的 Go 代码 |

---

## 请求流程详解

### 完整请求流程

当执行 `curl -i http://127.0.0.1:8888/ping` 时，请求的完整流程如下：

```
┌─────────────────────────────────────────────────────────────┐
│ 1. HTTP 请求到达                                            │
│    curl -i http://127.0.0.1:8888/ping                      │
│    ↓                                                         │
│    请求: GET /ping                                           │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. 路由匹配 (routes.go)                                     │
│    go-zero 框架根据 HTTP 方法和路径匹配路由                  │
│    匹配到: GET /ping -> pingHandler                         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. 处理器层 (pinghandler.go)                                │
│    pingHandler 函数被调用                                    │
│    ↓                                                         │
│    3.1 创建业务逻辑实例                                      │
│        l := logic.NewPingLogic(r.Context(), svcCtx)         │
│        - 传入请求上下文 (用于超时控制、日志追踪)            │
│        - 传入服务上下文 (包含配置和 RPC 客户端)             │
│    ↓                                                         │
│    3.2 调用业务逻辑                                          │
│        resp, err := l.Ping()                                │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. 业务逻辑层 (pinglogic.go)                                 │
│    Ping() 方法执行                                           │
│    ↓                                                         │
│    4.1 调用 RPC 服务 (可选)                                 │
│        l.svcCtx.GreetRpc.Ping(ctx, placeholder)            │
│        - 通过 gRPC 协议调用 RPC 服务 (127.0.0.1:8080)        │
│        - 如果 RPC 服务不可用，会返回错误但不会阻止响应       │
│    ↓                                                         │
│    4.2 创建响应对象                                          │
│        resp = new(types.Resp)                               │
│        resp.Msg = "pong"                                     │
│    ↓                                                         │
│    4.3 返回响应数据                                          │
│        return resp, err                                      │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. RPC 服务处理 (greetserver.go -> pinglogic.go)            │
│    (如果调用了 RPC 服务)                                     │
│    ↓                                                         │
│    5.1 RPC Server 接收 gRPC 请求                           │
│    5.2 调用 RPC Logic 执行业务逻辑                          │
│    5.3 返回 RPC 响应                                         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. 处理器层处理响应 (pinghandler.go)                        │
│    ↓                                                         │
│    6.1 判断是否有错误                                        │
│        if err != nil                                         │
│          → httpx.ErrorCtx() 返回错误响应                    │
│        else                                                  │
│          → httpx.OkJsonCtx() 返回成功响应                   │
│    ↓                                                         │
│    6.2 序列化响应为 JSON                                     │
│        {"msg": "pong"}                                       │
│    ↓                                                         │
│    6.3 设置 HTTP 状态码 200 OK                              │
│    ↓                                                         │
│    6.4 写入 HTTP 响应体                                      │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. HTTP 响应返回给客户端                                    │
│    HTTP/1.1 200 OK                                          │
│    Content-Type: application/json                           │
│    {"msg": "pong"}                                          │
└─────────────────────────────────────────────────────────────┘
```

### 数据流转图

```
HTTP 请求 (GET /ping)
  ↓
路由匹配 (GET /ping -> pingHandler)
  ↓
Handler (pingHandler)
  ↓ 创建 Logic 实例
Logic (PingLogic.Ping())
  ↓ 调用 RPC (可选)
RPC 服务 (127.0.0.1:8080)
  ↓ gRPC 调用
RPC Server (GreetServer.Ping())
  ↓
RPC Logic (PingLogic.Ping())
  ↓ 返回
Logic 组装响应 (Resp{Msg: "pong"})
  ↓ 返回
Handler 序列化为 JSON
  ↓
HTTP 响应 (200 OK, {"msg": "pong"})
```

---

## 核心概念说明

### 1. HTTP 和 REST 的关系

#### HTTP (HyperText Transfer Protocol)
- **本质**: 通信协议
- **作用**: 定义客户端和服务器之间的通信规则
- **特点**: 
  - 基于请求-响应模型
  - 无状态协议
  - 支持多种方法（GET、POST、PUT、DELETE 等）

#### REST (Representational State Transfer)
- **本质**: 架构风格
- **作用**: 定义如何设计和使用 Web API
- **特点**:
  - 资源导向（Resource-Oriented）
  - 使用标准 HTTP 方法操作资源
  - 无状态
  - 统一接口

#### 关系
```
REST（架构风格）
    ↓ 使用
HTTP（通信协议）
    ↓ 实现
Web API 服务
```

**简单理解**:
- HTTP = 工具（协议）
- REST = 使用工具的方法（架构风格）
- RESTful API = 用 HTTP 实现的 REST 风格 API

### 2. RPC 服务的作用

#### 为什么需要 RPC 服务？

1. **服务解耦**
   - API 服务：负责对外接口、参数校验、HTTP 协议处理
   - RPC 服务：负责核心业务逻辑、数据处理、数据库操作

2. **高性能内部通信**
   - HTTP REST API：面向外部，使用 HTTP/1.1，易调试
   - gRPC：面向内部，使用 HTTP/2，性能更好，支持流式传输

3. **支持微服务架构**
   - 多个 API 服务可以调用同一个 RPC 服务
   - RPC 服务可以被多个服务复用
   - 便于服务拆分和扩展

#### RPC 服务在项目中的位置

```
API 服务 (对外接口)
    ↓ 调用
RPC 服务 (业务逻辑)
    ↓ 可能调用
数据库/其他服务
```

### 3. 服务上下文 (ServiceContext)

**作用**: 依赖注入容器，在整个服务中共享配置和依赖服务

**包含内容**:
- 应用配置
- RPC 客户端（用于调用其他微服务）
- 数据库连接（如果有）
- 其他共享资源

**优势**:
- 避免在业务逻辑中直接创建依赖
- 便于测试（可以注入 mock 对象）
- 统一管理依赖资源

---

## 配置文件说明

### API 服务配置 (`api/etc/greet.yaml`)

```yaml
Name: ping        # 服务名称
Host: 127.0.0.1   # 服务监听的主机地址
Port: 8888        # 服务监听的端口号
Log:              # 日志配置
  Encoding: plain # 日志编码格式，plain 表示纯文本格式
```

### RPC 服务配置 (`rpc/etc/greet.yaml`)

```yaml
Name: greet.rpc    # RPC 服务名称
ListenOn: 127.0.0.1:8080  # RPC 服务监听地址和端口
Log:               # 日志配置
  Encoding: plain  # 日志编码格式
```

### API 服务中的 RPC 客户端配置

在 `api/internal/svc/servicecontext.go` 中：

```go
client := zrpc.MustNewClient(zrpc.RpcClientConf{
    Target: "127.0.0.1:8080", // RPC 服务地址
})
```

---

## 启动和测试

### 1. 启动 RPC 服务

```bash
cd greet/rpc
go run greet.go
```

**预期输出**:
```
Starting rpc server at 127.0.0.1:8080...
```

### 2. 启动 API 服务

```bash
cd greet/api
go run greet.go
```

**预期输出**:
```
Starting server at 127.0.0.1:8888...
```

### 3. 测试 API 接口

```bash
curl -i http://127.0.0.1:8888/ping
```

**预期响应**:
```
HTTP/1.1 200 OK
Content-Type: application/json

{"msg":"pong"}
```

### 4. 使用 BloomRPC 测试 RPC 服务

1. 下载并安装 BloomRPC: https://github.com/bloomrpc/bloomrpc/releases
2. 导入 `greet/rpc/greet.proto` 文件
3. 配置服务地址: `127.0.0.1:8080`
4. 调用 `ping` 方法进行测试

---

## 总结

### 项目特点

1. **分层架构**: API 层和 RPC 层职责清晰，便于维护和扩展
2. **服务解耦**: API 服务和 RPC 服务可以独立部署和扩展
3. **性能优化**: 外部使用 HTTP REST API，内部使用高性能 gRPC
4. **易于扩展**: 支持多个 API 服务共享同一个 RPC 服务

### 关键文件

- **API 服务入口**: `api/greet.go`
- **API 定义**: `api/greet.api`
- **RPC 服务入口**: `rpc/greet.go`
- **RPC 服务定义**: `rpc/greet.proto`

### 请求流程总结

1. 客户端发送 HTTP 请求到 API 服务 (8888端口)
2. API 服务路由匹配，调用对应的 Handler
3. Handler 创建 Logic 实例，调用业务逻辑
4. Logic 可选调用 RPC 服务 (8080端口) 获取数据
5. RPC 服务处理业务逻辑，返回结果
6. API 服务组装响应，返回给客户端

---

## 参考资源

- [go-zero 官方文档](https://go-zero.dev/)
- [gRPC 官方文档](https://grpc.io/)
- [Protocol Buffers 文档](https://protobuf.dev/)

---

*文档生成时间: 2024年*
*项目版本: go-zero v1.9.2*

