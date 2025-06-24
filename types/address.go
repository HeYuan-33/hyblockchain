package types

import "hyblockchain/crypto/sha3"

type Address [20]byte

func PubKeyToAddress(pub []byte) Address {
	// 去掉未压缩公钥的前缀 0x04（如果有）
	if len(pub) == 65 && pub[0] == 0x04 {
		pub = pub[1:]
	}

	hash := sha3.Keccak256(pub) // 对公钥哈希
	var addr Address
	copy(addr[:], hash[12:]) // 取最后20字节（160位）
	return addr
}
