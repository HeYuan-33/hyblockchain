package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
)

// GenerateKey 生成一个随机私钥（32字节）
func GenerateKey() ([]byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	if priv.D == nil {
		return nil, errors.New("invalid private key")
	}
	return priv.D.FillBytes(make([]byte, 32)), nil
}

// PubKeyFromPrivKey 返回未压缩的公钥字节（65字节: 0x04 + X + Y）
func PubKeyFromPrivKey(priv []byte) []byte {
	x, y := elliptic.P256().ScalarBaseMult(priv)
	pub := elliptic.Marshal(elliptic.P256(), x, y)
	return pub
}
