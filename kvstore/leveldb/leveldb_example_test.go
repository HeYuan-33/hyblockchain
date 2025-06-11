package leveldb

import (
	"os"
	"testing"
)

func TestLevelDB(t *testing.T) {
	dbPath := "testdb_example"
	defer os.RemoveAll(dbPath)

	db, err := NewLevelDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open leveldb: %v", err)
	}
	defer db.Close()

	key := []byte("hello")
	value := []byte("world")

	// Put
	err = db.Put(key, value)
	if err != nil {
		t.Errorf("Put failed: %v", err)
	}

	// Has
	has, err := db.Has(key)
	if err != nil {
		t.Errorf("Has failed: %v", err)
	}
	if !has {
		t.Errorf("Expected key to exist")
	}

	// Get
	got, err := db.Get(key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Expected %s, got %s", value, got)
	}

	// Delete
	err = db.Delete(key)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Has after delete
	has, err = db.Has(key)
	if err != nil {
		t.Errorf("Has (after delete) failed: %v", err)
	}
	if has {
		t.Errorf("Expected key to be deleted")
	}
}
