package mpt

import (
	"bytes"
	"hyblockchain/kvstore/leveldb"
	"testing"
)

func TestMPT(t *testing.T) {
	// 创建测试数据库
	db, err := leveldb.NewLevelDB("testdb")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	// 创建MPT
	mpt := NewMPT(db)

	// 测试Put和Get
	key := []byte("test_key")
	value := []byte("test_value")

	// 先测试空树的根哈希
	emptyHash := mpt.RootHash()
	if emptyHash == nil {
		t.Error("Empty tree root hash should not be nil")
	}

	err = mpt.Put(key, value)
	if err != nil {
		t.Errorf("Put failed: %v", err)
	}

	// 测试插入后的根哈希
	afterPutHash := mpt.RootHash()
	if afterPutHash == nil {
		t.Error("Root hash after put should not be nil")
	}
	if bytes.Equal(emptyHash, afterPutHash) {
		t.Error("Root hash should change after insertion")
	}

	got, err := mpt.Get(key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Expected %s, got %s", value, got)
	}

	// 测试Delete
	err = mpt.Delete(key)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// 测试删除后的根哈希
	afterDeleteHash := mpt.RootHash()
	if afterDeleteHash == nil {
		t.Error("Root hash after delete should not be nil")
	}
	if !bytes.Equal(emptyHash, afterDeleteHash) {
		t.Error("Root hash should be same as empty tree after deletion")
	}

	_, err = mpt.Get(key)
	if err == nil {
		t.Error("Expected error after delete")
	}
}
