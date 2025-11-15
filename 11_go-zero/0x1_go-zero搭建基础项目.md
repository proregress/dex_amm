# 0x1 go-zero 搭建基础项目

## 项目目标
- 掌握 go-zero 微服务框架的基本组成与工程结构。
- 能够使用 `goctl` 快速生成 API 与 RPC 服务骨架。
- 学会配置常用的开发环境变量，完成本地构建与运行。

## 环境准备
### 1. 安装 Go
- **macOS / Windows**：从 [Go 官方下载页](https://go.dev/dl/) 安装对应平台的安装包。
- **Debian / Ubuntu**：
  ```bash
  sudo apt update
  sudo apt install -y golang-go
  ```
- 安装完成后执行 `go version`，确保输出正确的版本号。

### 2. 配置全局环境
- 将 Go 安装目录（默认 `~/go/bin` 或 `C:\Users\<用户名>\go\bin`）加入 `PATH`。
- 为国内网络环境配置代理：
  ```bash
  go env -w GOPROXY=https://goproxy.cn,direct
  ```

### 3. 安装并配置 goctl
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```
- Go 会将可执行文件放在 `$(go env GOPATH)/bin`，需追加到 `PATH` 中：
  - Linux/macOS：`echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc && source ~/.bashrc`
  - Windows PowerShell：`[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;"+(go env GOPATH)+"\\bin", "User")`
- 验证安装：`goctl --version`

## 快速创建微服务骨架
```bash
goctl quickstart -t micro go_demo
```
- `go_demo` 为项目根目录名称，可按需替换。
- `-t micro` 表示生成一个包含 API 与 RPC 的微服务模板。

## 目录结构说明
```text
go_demo/
├── greet/                 # 主模块
│   ├── api/               # HTTP 网关服务（用户入口）
│   ├── rpc/               # gRPC 内部服务（业务实现）
│   ├── go.mod / go.sum    # 模块定义与依赖锁定
│   └── README.md
└── etc、scripts...        # 其他支持文件（按需生成）
```

## API 与 RPC 的协作流程
1. 客户端发送 HTTP 请求到 API 服务。
2. API 服务根据路由映射调用内部生成的 RPC 客户端。
3. RPC 服务完成核心业务逻辑处理并返回数据。
4. API 服务封装响应并返回给客户端。

## 本地运行与调试
1. 启动 RPC 服务：
   ```bash
   cd go_demo/greet/rpc
   go run greet.go
   ```
2. 另开终端启动 API 服务：
   ```bash
   cd go_demo/greet/api
   go run greet.go
   ```
3. 使用 `curl` 或 Postman 调用 API，验证服务链路是否正常。
```bash
MacBook-Pro:~$ curl -i http://127.0.0.1:8888/ping

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Traceparent: 00-b9f9a12a54e7cfdaa5cbf23106a28265-4c5f15030e9e7674-00
Date: Sat, 15 Nov 2025 03:17:50 GMT
Content-Length: 14
```

## 常见问题与排错
- **`command not found: goctl`**：确认 `$(go env GOPATH)/bin` 已加入 `PATH`。
- **拉取依赖缓慢**：再次确认 `GOPROXY` 设置为国内镜像。
- **端口被占用**：修改 `etc` 目录下的配置文件或停止占用端口的进程。
- **Go 版本过低**：go-zero 推荐 Go 1.18+，执行 `go version` 检查。

## 拓展阅读
- go-zero 源码仓库：<https://github.com/zeromicro/go-zero>
- 官方开发文档：<https://go-zero.dev/>
- goctl 使用指南：<https://go-zero.dev/#%E4%BB%A3%E7%A0%81%E8%87%AA%E5%8A%A8%E7%94%9F%E6%88%90>
- go-zero Awesome 资源汇总：<https://github.com/zeromicro/awesome-zero>

## 后续练习方向
- 自定义 API 路由与请求响应结构，生成代码后补全业务逻辑。
- 学习 go-zero 的配置中心、服务发现和链路追踪组件。
- 将生成的服务容器化，构建本地 Docker Compose 运行环境。
