package client

import (
	"fmt"
	"swap/config"
	"swap/wallet"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

// PumpSwapClient 客户端对象
type PumpSwapClient struct {
	config    *config.SwapConfig
	wallet    wallet.SecureWallet
	rpcClient *rpc.Client
	logger    *logrus.Logger

	// 内部组件

}

// 用swapConfig和wallet创建一个客户端
func NewPumpSwapClient(swapConfig *config.SwapConfig, wallet *wallet.SecureWallet) (*PumpSwapClient, error) {
	if swapConfig == nil {
		swapConfig = config.DefaultConfig()
	}

	if err := swapConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid swap config: %w", err)
	}

	rpcClient := rpc.New(swapConfig.RPCEndpoint)

	logger := logrus.New()
	if swapConfig.EnableDebugLog {
		logger.SetLevel(logrus.DebugLevel)
	}

	client := &PumpSwapClient{
		config:    swapConfig,
		wallet:    wallet,
		rpcClient: rpcClient,
		logger:    logger,
	}
}
