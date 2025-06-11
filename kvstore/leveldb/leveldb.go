package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"hyblockchain/kvstore"
)

// LevelDB 实现了KVStore接口
type LevelDB struct {
	db *leveldb.DB
}

// NewLevelDB 创建一个新的LevelDB实例
func NewLevelDB(path string) (kvstore.KVStore, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDB{db: db}, nil
}

// Get 获取指定键的值
func (l *LevelDB) Get(key []byte) ([]byte, error) {
	return l.db.Get(key, nil)
}

// Put 存储键值对
func (l *LevelDB) Put(key, value []byte) error {
	return l.db.Put(key, value, nil)
}

// Delete 删除指定键
func (l *LevelDB) Delete(key []byte) error {
	return l.db.Delete(key, nil)
}

// Has 检查键是否存在
func (l *LevelDB) Has(key []byte) (bool, error) {
	return l.db.Has(key, nil)
}

// Batch 创建新的批量操作
func (l *LevelDB) Batch() kvstore.Batch {
	return &levelDBBatch{
		batch: new(leveldb.Batch),
	}
}

// Write 执行批量操作
func (l *LevelDB) Write(batch kvstore.Batch) error {
	b, ok := batch.(*levelDBBatch)
	if !ok {
		return nil
	}
	return l.db.Write(b.batch, nil)
}

// NewIterator 创建新的迭代器
func (l *LevelDB) NewIterator(prefix []byte) kvstore.Iterator {
	iter := l.db.NewIterator(util.BytesPrefix(prefix), nil)
	return &levelDBIterator{iter: iter}
}

// Close 关闭数据库连接
func (l *LevelDB) Close() error {
	return l.db.Close()
}

// levelDBBatch 实现了Batch接口
type levelDBBatch struct {
	batch *leveldb.Batch
}

func (b *levelDBBatch) Put(key, value []byte) {
	b.batch.Put(key, value)
}

func (b *levelDBBatch) Delete(key []byte) {
	b.batch.Delete(key)
}

func (b *levelDBBatch) Reset() {
	b.batch.Reset()
}

func (b *levelDBBatch) Len() int {
	return b.batch.Len()
}

// levelDBIterator 实现了Iterator接口
type levelDBIterator struct {
	iter iterator.Iterator
}

func (i *levelDBIterator) Next() bool {
	return i.iter.Next()
}

func (i *levelDBIterator) Key() []byte {
	return i.iter.Key()
}

func (i *levelDBIterator) Value() []byte {
	return i.iter.Value()
}

func (i *levelDBIterator) Error() error {
	return i.iter.Error()
}

func (i *levelDBIterator) Release() {
	i.iter.Release()
}
