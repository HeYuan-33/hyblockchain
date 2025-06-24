package blockchain

import (
	"hyblockchain/crypto/sha3"
	"hyblockchain/trie"
	"hyblockchain/txpool"
	"hyblockchain/types"
	"hyblockchain/utils/hash"
	"hyblockchain/utils/rlp"
)

type Header struct {
	Root       hash.Hash
	ParentHash hash.Hash
	Height     uint64
	Coinbase   types.Address
	Timestamp  uint64

	Nonce uint64
}

type Body struct {
	Transactions []types.Transaction
	Receiptions  []types.Receiption
}

func (header Header) Hash() hash.Hash {
	data, _ := rlp.EncodeToBytes(header)
	return sha3.Keccak256(data)
}

func NewHeader(parent Header) *Header {
	return &Header{
		Root:       parent.Root,
		ParentHash: parent.Hash(),
		Height:     parent.Height + 1,
	}
}

func NewBlock() *Body {
	return &Body{
		Transactions: make([]types.Transaction, 0),
		Receiptions:  make([]types.Receiption, 0),
	}
}

type Blockchain struct {
	CurrentHeader Header
	Statedb       trie.ITrie
	Txpool        txpool.TxPool
	Current       Header
}
