package statdb

import (
	"hash"
	"hyblockchain/types"
)

type StatDB interface {
	SetStatRoot(root hash.Hash)
	Load(addr types.Address) *types.Account
	Store(addr types.Address, account types.Account)
	SetRoot(root hash.Hash)
}
