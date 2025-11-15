// config.go - 配置结构体定义文件
// 这个文件定义了应用程序的配置结构
// Config 结构体嵌入了 go-zero 的 RestConf，包含了 HTTP 服务的所有配置项
// 配置文件 greet.yaml 会被解析到这个结构体中

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

// Config - 应用程序配置结构体
// 嵌入了 rest.RestConf，包含了服务名称、主机、端口、日志等配置
// 这些配置项对应 greet.yaml 文件中的配置
type Config struct {
	rest.RestConf
}
