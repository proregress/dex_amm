package block

import (
	"fmt"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/mr-tron/base58"
)

func DecodePumpFunInstruction(inst *types.CompiledInstruction, tx *client.BlockTransaction) (err error) {
	fmt.Println("pump.fun transactions", base58.Encode(tx.Transaction.Signatures[0]))
	return
}

func DecodeRaydiumInstruction(inst *types.CompiledInstruction, tx *client.BlockTransaction) (err error) {
	fmt.Println("pump.fun transactions", base58.Encode(tx.Transaction.Signatures[0]))
	return
}

/*
指令的data的前8位是discriminator，用于区分不同的指令
*/
func GetInstructionDiscriminator(data []byte) []byte {
	if len(data) < 8 || data == nil {
		return nil
	}
	return data[:8]
}
