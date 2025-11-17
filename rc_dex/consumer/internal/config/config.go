package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"richcode.cc/dex/pkg/constants"
)

var Cfg Config

// 轮询（round-robin）负载均衡计数器，用于在多个 Solana RPC 节点之间轮询选择。解决单个节点故障问题
var SolRpcUseFrequency int

type Config struct {
	zrpc.RpcServerConf

	Sol Chain `json:"Sol,optional"`

	ConsumerConfig ConsumerConfig `json:"ConsumerConfig,optional"`
}

type ConsumerConfig struct {
	Concurrency int `json:"Concurrency" env:"CONSUMER_CONCURRENCY"`
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

/*
(rpc string) ： 命名返回值，可以直接赋值，return 时无需显式写变量
*/
func FindChainRpcByChainId(chainId int) (rpc string) {
	var rpcs []string
	var useFrequency *int

	switch chainId {
	case constants.SolChainIdInt:
		rpcs = Cfg.Sol.NodeUrl
		useFrequency = &SolRpcUseFrequency
	default:
		logx.Error("No Rpc Config")
		return
	}

	if len(rpcs) == 0 {
		logx.Error("No Rpc Config")
		return
	}

	*useFrequency++
	index := *useFrequency % len(rpcs)
	rpc = rpcs[index]
	return
}
