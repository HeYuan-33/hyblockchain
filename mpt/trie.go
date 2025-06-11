package mpt

type Trie interface {
	Put(key, value string)
	Get(key string) (string, bool)
	Delete(key string) bool
	RootHash() string
}
