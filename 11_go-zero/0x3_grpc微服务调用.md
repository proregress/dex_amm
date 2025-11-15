# 0x3 gRPC 微服务调用

## 学习目标
- 理解 go-zero gRPC 网关如何把 HTTP 请求映射到微服务
- 独立搭建并调试一个最小可用的 gRPC 网关 Demo
- 掌握常见配置项、调试方式与排错思路

## 背景速览
在 go-zero 微服务体系中，gRPC 负责服务间通信，而对外往往仍需提供 RESTful API。gRPC 网关承担协议转换与能力治理工作，使前端能保持熟悉的 HTTP 调用方式，同时复用 gRPC 的强类型、高效传输与中台治理。

## 工作原理
1. **Proto 描述解析**：读取 `.proto` 文件，获得服务、消息及 `google.api.http` 注解定义。
2. **HTTP 映射加载**：读取网关配置（如 `yaml/json`），补充服务发现、负载均衡与治理策略。
3. **生成处理器**：基于 proto 与配置生成 HTTP → gRPC 适配器（代码生成或运行时注册）。
4. **启动 HTTP 网关**：监听客户端 REST 请求，并附加鉴权、限流、日志等中间件。
5. **请求转换**：将 HTTP Path、Method、Query、Body 映射为 gRPC 请求消息。
6. **调用下游服务**：通过 gRPC Stub 调用真实的微服务节点。
7. **响应转换**：把 gRPC 响应封装为 HTTP 状态码、Header 与 JSON Body。
8. **返回结果**：向调用方返回结果，同时上报链路日志、指标与监控。

> 建议：统一维护 proto 文件，使网关与下游服务共享定义，避免接口不一致。

## 上手实践
### 1. 编写 Proto
```proto
syntax = "proto3";
package user;

import "google/api/annotations.proto";

service UserService {
  rpc GetProfile(GetProfileReq) returns (GetProfileResp) {
    option (google.api.http) = {
      get: "/api/v1/users/{id}"
    };
  }
}

message GetProfileReq {
  string id = 1;
}

message GetProfileResp {
  string id = 1;
  string nickname = 2;
}
```

### 2. 生成代码
```bash
protoc --go_out=. --go-grpc_out=. \
       --grpc-gateway_out=. --grpc-gateway_opt logtostderr=true \
       api/user.proto
```
该命令会生成 gRPC 服务端、客户端 Stub 与 HTTP 网关适配器代码。

### 3. 准备配置
```yaml
Name: user.rpc
Host: 0.0.0.0
Port: 8080

Timeout: 3000
```
网关利用此配置完成服务发现、超时控制与治理策略加载。

### 4. 启动流程
- 运行 RPC 服务：`go run service/user/rpc/main.go`
- 运行 HTTP 网关：`go run service/user/api/main.go`

### 5. 发起调用
```bash
curl -i http://localhost:8080/api/v1/users/10001
```
若返回 JSON 用户信息，说明请求已经成功完成 HTTP ↔ gRPC 转换。

## 调试与排错
- **日志定位**：打开 go-zero 默认日志，重点关注请求耗时与 gRPC Status。
- **链路追踪**：接入 OpenTelemetry，快速定位慢调用与跨服务故障点。
- **超时设置**：保持 HTTP 与 gRPC 客户端超时一致，避免请求悬挂。
- **错误映射**：定义好 gRPC Status 与 HTTP 状态码的对应关系，便于前端处理。
- **兼容性校验**：确保网关与服务使用同一份 proto，并重新生成兼容的代码。

## 进阶建议
- 借助 goctl 生成 API/RPC 模板，提高接口一致性与开发效率。
- 下沉公共能力（鉴权、限流、监控）到网关，避免服务重复实现。
- 大体量传输（文件、流）优先评估直接 gRPC 或对象存储加速方案。

## 参考资料
- go-zero gRPC 网关教程：<https://go-zero.dev/docs/tutorials/gateway/grpc>
- grpc-gateway 项目主页：<https://github.com/grpc-ecosystem/grpc-gateway>
- Google API HTTP 映射规范：<https://cloud.google.com/endpoints/docs/grpc-service-config/reference/rpc/google.api#http>
