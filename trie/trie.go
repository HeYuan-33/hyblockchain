package trie

import "hyblockchain/utils/hash"

type ITrie interface {
	Store(key []byte, value []byte) error
	Root() hash.Hash
	Load(key []byte) ([]byte, error)
}

type TrieNode struct {
	Path     string
	Children []Child
}

type Children []Child
type Child struct {
	Path string
	Hash hash.Hash
}
