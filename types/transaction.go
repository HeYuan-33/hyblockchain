package types

import (
	"fmt"
	"hyblockchain/crypto/secp256k1"
	"hyblockchain/crypto/sha3"
	"hyblockchain/utils/hexutil"
	"hyblockchain/utils/rlp"
	"math/big"
)

type Transaction struct {
	txdata
	signature
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

func (tx Transaction) GasPrice() uint64 {
	return tx.txdata.GasPrice
}

func (tx Transaction) Nonce() uint64 {
	return tx.txdata.Nonce
}

func (tx Transaction) From() Address {
	if tx.R == nil || tx.S == nil {
		panic("signature missing")
	}

	// 编码 txdata
	txdata := tx.txdata
	toSign, err := rlp.EncodeToBytes(txdata)
	if err != nil {
		panic(fmt.Errorf("RLP encoding failed: %v", err))
	}
	fmt.Println("Encoded txdata:", hexutil.Encode(toSign))

	// 计算消息哈希
	msg := sha3.Keccak256(toSign)

	// 构造签名字节数组 sig (R||S||V)
	sig := make([]byte, 65)
	rBytes := tx.R.Bytes()
	sBytes := tx.S.Bytes()

	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sig[64] = tx.V

	// 恢复公钥
	pubKey, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		panic(fmt.Errorf("recover pubkey failed: %v", err))
	}

	return PubKeyToAddress(pubKey)
}

func NewTransactionWithSigner(to Address, nonce, gas, value, gasPrice uint64, privKey []byte) (*Transaction, error) {
	tx := &Transaction{
		txdata: txdata{
			To:       to,
			Nonce:    nonce,
			Gas:      gas,
			Value:    value,
			GasPrice: gasPrice,
		},
	}

	txBytes, err := rlp.EncodeToBytes(tx.txdata)
	if err != nil {
		return nil, err
	}
	msgHash := sha3.Keccak256(txBytes)

	sig, err := secp256k1.Sign(msgHash, privKey)
	if err != nil {
		return nil, err
	}

	tx.R = new(big.Int).SetBytes(sig[:32])
	tx.S = new(big.Int).SetBytes(sig[32:64])
	tx.V = sig[64]

	return tx, nil
}
