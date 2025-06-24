package txpool

import (
	"hyblockchain/types"
	"hyblockchain/utils/hash"
)

type TxPool interface {
	NewTx(tx *types.Transaction) TxPool

	Pop() *types.Transaction

	SetStatRoot(h hash.Hash)

	NotifyTxEvent(txs []*types.Transaction)
}
