package wallet

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// SecureWallet 安全钱包接口
type SecureWallet interface {
	PublicKey() solana.PublicKey
	Sign(message []byte) ([]byte, error)
	SignTransaction(transaction *solana.Transaction) error
}

// MemoryWallet 内存钱包，用于存储私钥
type MemoryWallet struct {
	privateKey solana.PrivateKey
}

func NewMemoryWallet(privateKey solana.PrivateKey) *MemoryWallet {
	return &MemoryWallet{privateKey: privateKey}
}

func NewMemoryWalletFromBase58(privateKeyStr string) (*MemoryWallet, error) {
	privateKey, err := solana.PrivateKeyFromBase58(privateKeyStr) // todo 这里解析出来应该和我配置的私钥二进制一致
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	fmt.Printf("👛 钱包地址：%s.\n", privateKey)
	return NewMemoryWallet(privateKey), nil
}

// publicKey 返回钱包的公钥
func (w *MemoryWallet) PublicKey() solana.PublicKey {
	return w.privateKey.PublicKey()
}

// Sign 签名消息
func (w *MemoryWallet) Sign(message []byte) ([]byte, error) {
	signature, err := w.privateKey.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	return signature[:], nil
}

// SignTransaction 签名交易
func (w *MemoryWallet) SignTransaction(tx *solana.Transaction) error {
	_, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(w.PublicKey()) {
			return &w.privateKey
		}
		return nil
	})
	return err
}
