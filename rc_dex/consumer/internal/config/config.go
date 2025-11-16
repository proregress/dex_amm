package config

import "github.com/zeromicro/go-zero/zrpc"

// 不知道在哪里使用，先注释掉
// var Cfg Config

// var (
// 	SolRpcUseFrequency int
// )

type Config struct {
	zrpc.RpcServerConf

	Sol Chain `json:"Sol,optional"`
}

// json 标签 → 匹配 YAML/JSON 配置文件
// env 标签 → 指定环境变量名（用于覆盖配置）
type Chain struct {
	ChainId    int64    `json:"ChainId"             env:"SOL_CHAINID"`
	NodeUrl    []string `json:"NodeUrl"             env:"SOL_NODEURL"`
	WSUrl      string   `json:"WSUrl,optional"      env:"SOL_WSURL"`
	MEVNodeUrl string   `json:"MevNodeUrl,optional" env:"SOL_MEVNODEURL"`
	StartBlock uint64   `json:"StartBlock,optional" env:"SOL_STARTBLOCK"`
}
