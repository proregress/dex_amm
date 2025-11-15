package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"swap/config"
	"swap/wallet"

	"github.com/mr-tron/base58"
)

func main() {
	fmt.Println("🚀 Pump.fun AMM Swap Start!")
	fmt.Println(strings.Repeat("=", 50))

	// 1. 准备钱包信息
	rpcEndpoint := getEnvOrDefault("RPC_ENDPOINT", "http://api.devnet.solana.com")
	privateKeyArr := []byte{172, 254, 142, 221, 147, 137, 233, 182, 189, 100, 45, 12, 9, 141, 74, 187, 22, 151, 243, 72, 227, 34, 224, 218, 92, 211, 206, 167, 125, 152, 100, 129, 235, 135, 85, 90, 106, 132, 39, 123, 201, 171, 57, 209, 164, 200, 109, 9, 76, 241, 19, 19, 135, 28, 127, 247, 211, 221, 190, 87, 162, 8, 203, 50}
	privateKeyStr := getEnvOrDefault("PRIVATE_KEY", base58.Encode(privateKeyArr))

	if privateKeyStr == "" {
		log.Fatal("❌ 需要设置 PRIVATE_KEY 环境变量")
	}

	// 2. 客户端配置，用于初始化交换客户端
	cfg := config.DefaultConfig()
	cfg.RPCEndpoint = rpcEndpoint
	cfg.EnableDebugLog = true

	fmt.Printf("📢 RPC 节点：%s.\n", cfg.RPCEndpoint)

	// 3. 根据私钥恢复/加载钱包对象（用于签名交易）
	// 基于1中的私钥字符串得到的钱包对象，w.publicKey()是钱包的地址，todo 打印出来看下是否与Solana address得到的地址一致GrQVv3uEobKDrfXbVaP6qvEA6ioX5yZUASV3FCz7xoFw
	w, err := wallet.NewMemoryWalletFromBase58(privateKeyStr)
	if err != nil {
		log.Fatalf("❌ 获取钱包对象失败：%v.\n", err) //%v 是 Go 的通用格式化动词，适用于大多数类型
	}

	fmt.Printf("👛 钱包地址：%s.\n", w.PublicKey().String())
	// 4. 创建客户端连接对象
	// 如果函数名首字母大写（公开），可以直接通过包名调用

	// 5. 构建交易请求

}

// os.Getenv(key) 直接从操作系统的环境变量中读取配置
/*
方法一：在终端中临时设置（当前会话有效）
export RPC_ENDPOINT="https://api.mainnet-beta.solana.com"
export PRIVATE_KEY="your-private-key"
*/
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
