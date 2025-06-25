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

	mpt := NewMPT(db)

	// 测试空树根哈希不为空且一致
	emptyHash := mpt.RootHash()
	if emptyHash == nil {
		t.Fatal("Empty tree root hash should not be nil")
	}
	t.Logf("Empty root hash: %x", emptyHash)

	// 插入单个键值对，测试Put和Get
	key1 := []byte("key1")
	val1 := []byte("value1")

	if err := mpt.Put(key1, val1); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	afterPutHash := mpt.RootHash()
	if afterPutHash == nil {
		t.Fatal("Root hash after put should not be nil")
	}
	if bytes.Equal(emptyHash, afterPutHash) {
		t.Fatal("Root hash should change after insertion")
	}
	t.Logf("Root hash after inserting key1: %x", afterPutHash)

	got, err := mpt.Get(key1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(got, val1) {
		t.Fatalf("Get returned wrong value: want %s, got %s", val1, got)
	}

	// 插入第二个键值对，测试多键支持
	key2 := []byte("key2")
	val2 := []byte("value2")

	if err := mpt.Put(key2, val2); err != nil {
		t.Fatalf("Put key2 failed: %v", err)
	}
	t.Logf("Root hash after inserting key2: %x", mpt.RootHash())

	got2, err := mpt.Get(key2)
	if err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}
	if !bytes.Equal(got2, val2) {
		t.Fatalf("Get key2 returned wrong value: want %s, got %s", val2, got2)
	}

	// 更新 key1 的值
	val1Updated := []byte("value1_updated")
	if err := mpt.Put(key1, val1Updated); err != nil {
		t.Fatalf("Put update key1 failed: %v", err)
	}
	t.Logf("Root hash after updating key1: %x", mpt.RootHash())

	gotUpdated, err := mpt.Get(key1)
	if err != nil {
		t.Fatalf("Get updated key1 failed: %v", err)
	}
	if !bytes.Equal(gotUpdated, val1Updated) {
		t.Fatalf("Get updated key1 returned wrong value: want %s, got %s", val1Updated, gotUpdated)
	}

	// 删除 key2
	if err := mpt.Delete(key2); err != nil {
		t.Fatalf("Delete key2 failed: %v", err)
	}
	t.Logf("Root hash after deleting key2: %x", mpt.RootHash())

	// 删除后查询 key2 应失败
	_, err = mpt.Get(key2)
	if err == nil {
		t.Fatalf("Expected error when getting deleted key2")
	}

	// 删除不存在的key，应该不报错且根哈希不变
	nonExistKey := []byte("not_exist")
	rootBeforeDelNonExist := mpt.RootHash()
	if err := mpt.Delete(nonExistKey); err != nil {
		t.Fatalf("Delete non-existing key returned error: %v", err)
	}
	rootAfterDelNonExist := mpt.RootHash()
	if !bytes.Equal(rootBeforeDelNonExist, rootAfterDelNonExist) {
		t.Fatalf("Root hash should not change when deleting non-existing key")
	}

	// 删除 key1，树应恢复为空树状态（根哈希与初始相同）
	if err := mpt.Delete(key1); err != nil {
		t.Fatalf("Delete key1 failed: %v", err)
	}
	t.Logf("Root hash after deleting key1: %x", mpt.RootHash())

	if !bytes.Equal(mpt.RootHash(), emptyHash) {
		t.Fatalf("Root hash after deleting all keys should equal empty tree root hash")
	}

	// 连续插入和删除多键
	keys := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	values := [][]byte{[]byte("va"), []byte("vb"), []byte("vc")}

	for i, k := range keys {
		if err := mpt.Put(k, values[i]); err != nil {
			t.Fatalf("Put key %s failed: %v", k, err)
		}
	}
	t.Logf("Root hash after batch insert: %x", mpt.RootHash())

	for _, k := range keys {
		got, err := mpt.Get(k)
		if err != nil {
			t.Fatalf("Get key %s failed: %v", k, err)
		}
		t.Logf("Get key %s = %s", k, got)
	}

	for _, k := range keys {
		if err := mpt.Delete(k); err != nil {
			t.Fatalf("Delete key %s failed: %v", k, err)
		}
	}
	t.Logf("Root hash after batch delete: %x", mpt.RootHash())

	if !bytes.Equal(mpt.RootHash(), emptyHash) {
		t.Fatalf("Root hash after deleting all batch keys should equal empty tree root hash")
	}
}
