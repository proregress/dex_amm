package config

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

// SwapConfig 统一配置管理，其实最重要的是RPC端点
type SwapConfig struct {
	// 网络配置
	RPCEndpoint string             `json:"rpc_endpoint"`
	Commitment  rpc.CommitmentType `json:"commitment"` // 交易确认级别：finalized、confirmed、processed

	// 交易配置
	ComputeUnitprice uint64        `json:"compute_unit_price"` // todo ： what?
	ComputeUnitLimit uint32        `json:"compute_unit_limit"` // todo ： what?
	Maxretries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`

	// 安全配置
	/*
		*big.Int的说明
			-使用指针可以区分“未设置”和“已设置”：
			    --nil = 未设置（不限制）
			    --非 nil = 已设置（有阈值限制）
			-big.Int 是结构体，不是基本类型：
			    --零值 big.Int{} 表示 0，但无法表示“未设置”
			    --指针 *big.Int 的零值是 nil，可以表示“未设置”
	*/
	MaxSlippageBP      uint32   `json:"max_slippage_bp"`      // 最大允许滑点：10000 = 100%，例如1000 = 10%
	MinAmountThreshold *big.Int `json:"min_amount_threshold"` // 最小交易金额
	MaxAmountThreshold *big.Int `json:"max_amount_threshold"` // 最大交易金额

	// 调试配置
	EnableDebugLog bool   `json:"enable_debug_log"` // 是否启用调试日志
	LogLevel       string `json:"log_level"`        // 日志级别：debug、info、warn、error

}

func DefaultConfig() *SwapConfig {
	return &SwapConfig{
		RPCEndpoint:        "https://api.devnet.solana.com",
		Commitment:         rpc.CommitmentProcessed,
		ComputeUnitprice:   1000,
		ComputeUnitLimit:   200_000,
		Maxretries:         3,
		RetryDelay:         time.Second * 2,
		MaxSlippageBP:      5000,
		MinAmountThreshold: big.NewInt(1000),
		MaxAmountThreshold: big.NewInt(100_000_000_000),
		EnableDebugLog:     false,
		LogLevel:           "info",
	}
}

func (c *SwapConfig) Validate() error {
	if c.RPCEndpoint == "" {
		return fmt.Errorf("rpc endpoint is required")
	}

	if c.ComputeUnitLimit == 0 {
		return fmt.Errorf("compute unit limit must be greater than 0")
	}

	if c.MaxSlippageBP > 10000 {
		return fmt.Errorf("max slippage must be less than 10000")
	}

	if c.Maxretries < 0 {
		return fmt.Errorf("max retries must be greater than 0")
	}

	return nil
}
