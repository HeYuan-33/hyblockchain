package leveldb

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkLevelDB_Put(b *testing.B) {
	dbPath := "testdb_benchmark_put"
	defer os.RemoveAll(dbPath)

	db, err := NewLevelDB(dbPath)
	if err != nil {
		b.Fatalf("failed to open leveldb: %v", err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value := []byte(fmt.Sprintf("value%d", i))
		if err := db.Put(key, value); err != nil {
			b.Errorf("Put failed: %v", err)
		}
	}
}

func BenchmarkLevelDB_Get(b *testing.B) {
	dbPath := "testdb_benchmark_get"
	defer os.RemoveAll(dbPath)

	db, err := NewLevelDB(dbPath)
	if err != nil {
		b.Fatalf("failed to open leveldb: %v", err)
	}
	defer db.Close()

	// 先写入数据
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value := []byte("value")
		db.Put(key, value)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		_, err := db.Get(key)
		if err != nil {
			b.Errorf("Get failed: %v", err)
		}
	}
}
