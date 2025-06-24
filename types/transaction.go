package types

import (
	"fmt"
	"hash"
	"hyblockchain/crypto/secp256k1"
	"hyblockchain/crypto/sha3"
	"hyblockchain/utils/hexutil"
	"hyblockchain/utils/rlp"
	"math/big"
)

type Receiption struct {
	TxHash hash.Hash
	Status int
	// GasUsed int
	// Logs
}

type Transaction struct {
	txdata
	signature
}

func (tx Transaction) GasPrice() uint64 {
	return tx.txdata.GasPrice
}

func (tx Transaction) Push(tx1 *Transaction) {

}

func (tx Transaction) Pop() *Transaction {
	return nil
}

func (tx Transaction) Nonce() uint64 {
	return tx.txdata.Nonce
}

type txdata struct {
	To       Address
	Nonce    uint64
	Value    uint64
	Gas      uint64
	GasPrice uint64
	Input    []byte
}

type signature struct {
	R, S *big.Int
	V    uint8
}

func (tx Transaction) From() Address {
	// 1. 编码交易数据
	txdata := tx.txdata
	toSign, err := rlp.EncodeToBytes(txdata)
	if err != nil {
		panic(fmt.Errorf("RLP encoding failed: %v", err))
	}
	fmt.Println("Encoded txdata:", hexutil.Encode(toSign))

	// 2. 计算消息哈希
	msg := sha3.Keccak256(toSign)

	// 3. 构造签名 sig: R(32字节) || S(32字节) || V(1字节)
	sig := make([]byte, 65)

	rBytes := tx.R.Bytes()
	sBytes := tx.S.Bytes()

	// 注意：Bytes() 可能不足 32 字节，需要左边补 0
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sig[64] = tx.V

	// 4. 恢复公钥
	pubKey, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		panic(fmt.Errorf("recover pubkey failed: %v", err))
	}

	// 5. 计算地址（例如 keccak(pubkey)[12:]）
	return PubKeyToAddress(pubKey)
}
