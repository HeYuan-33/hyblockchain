package kvstore

import (
	"io"
)

// KVStore 定义了键值存储的接口
type KVStore interface {
	// 基本操作
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Has(key []byte) (bool, error)

	// 批量操作
	Batch() Batch
	Write(batch Batch) error

	// 迭代器
	NewIterator(prefix []byte) Iterator

	// 关闭存储
	io.Closer
}

// Batch 定义了批量操作的接口
type Batch interface {
	Put(key []byte, value []byte)
	Delete(key []byte)
	Reset()
	Len() int
}

// Iterator 定义了迭代器的接口
type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
	Release()
}
