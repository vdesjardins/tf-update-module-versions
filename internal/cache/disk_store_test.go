package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) (*DiskStore, string) {
	tmpDir := t.TempDir()
	store, err := NewDiskStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return store, tmpDir
}

func TestDiskStore_Set_Get(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	testValue := "test_data"
	if err := store.Set("key1", testValue, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := store.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != testValue {
		t.Errorf("expected %v, got %v", testValue, retrieved)
	}
}

func TestDiskStore_Get_NotFound(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	retrieved, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil, got %v", retrieved)
	}
}

func TestDiskStore_Get_Expired(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	testValue := "expired_data"
	if err := store.Set("key1", testValue, 10*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	retrieved, err := store.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil (expired), got %v", retrieved)
	}
}

func TestDiskStore_Delete(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer store.Close()

	if err := store.Set("key1", "data", 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := store.Delete("key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	retrieved, err := store.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil after delete, got %v", retrieved)
	}

	// Verify file was deleted from disk
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			t.Errorf("cache file still exists after delete: %s", f.Name())
		}
	}
}

func TestDiskStore_Clear(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer store.Close()

	for i := 1; i <= 3; i++ {
		key := "key" + string(rune(i))
		if err := store.Set(key, "data"+string(rune(i)), 1*time.Hour); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all entries are gone
	for i := 1; i <= 3; i++ {
		key := "key" + string(rune(i))
		retrieved, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if retrieved != nil {
			t.Errorf("expected nil, got %v", retrieved)
		}
	}

	// Verify disk is clean
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			t.Errorf("cache file still exists after clear: %s", f.Name())
		}
	}
}

func TestDiskStore_Exists(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Non-existent key
	exists, err := store.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Errorf("expected false for nonexistent key")
	}

	// Store a key
	if err := store.Set("key1", "data", 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	exists, err = store.Exists("key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected true for existing key")
	}

	// Expired key
	if err := store.Set("key2", "data", 10*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	exists, err = store.Exists("key2")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Errorf("expected false for expired key")
	}
}

func TestDiskStore_GetExpired(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Add some non-expiring entries
	for i := 1; i <= 2; i++ {
		key := "key" + string(rune(i))
		if err := store.Set(key, "data"+string(rune(i)), 1*time.Hour); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Add some expiring entries
	for i := 3; i <= 5; i++ {
		key := "key" + string(rune(i))
		if err := store.Set(key, "data"+string(rune(i)), 10*time.Millisecond); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)

	expired, err := store.GetExpired()
	if err != nil {
		t.Fatalf("GetExpired failed: %v", err)
	}

	if len(expired) != 3 {
		t.Errorf("expected 3 expired entries, got %d", len(expired))
	}
}

func TestDiskStore_InvalidKey(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	if err := store.Set("", "data", 1*time.Hour); err == nil {
		t.Errorf("expected error for empty key")
	}
}

func TestDiskStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first store and add data
	store1, err := NewDiskStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store1: %v", err)
	}

	testValue := "persistent_data"
	if err := store1.Set("persistent_key", testValue, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	store1.Close()

	// Create second store using same directory
	store2, err := NewDiskStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store2: %v", err)
	}
	defer store2.Close()

	// Verify data persisted
	retrieved, err := store2.Get("persistent_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != testValue {
		t.Errorf("expected %v, got %v", testValue, retrieved)
	}
}

func TestDiskStore_ComplexValues(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	type ComplexData struct {
		ID   int
		Name string
	}

	complexValue := ComplexData{ID: 123, Name: "test"}
	if err := store.Set("complex_key", complexValue, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := store.Get("complex_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != complexValue {
		t.Errorf("expected %v, got %v", complexValue, retrieved)
	}
}

func TestDiskStore_NilValue(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Storing nil should work
	if err := store.Set("nil_key", nil, 1*time.Hour); err != nil {
		t.Fatalf("Set nil failed: %v", err)
	}

	retrieved, err := store.Get("nil_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved != nil {
		t.Errorf("expected nil, got %v", retrieved)
	}
}

func TestDiskStore_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Should handle empty directory gracefully
	store, err := NewDiskStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	if err := store.Set("key1", "data", 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
}

func TestDiskStore_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "nested", "cache", "dir")

	store, err := NewDiskStore(cacheDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Verify directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("cache directory was not created")
	}

	// Verify we can use it
	if err := store.Set("key1", "data", 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
}

func TestDiskStore_DeleteNonexistent(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Should not error when deleting non-existent key
	if err := store.Delete("nonexistent"); err != nil {
		t.Fatalf("Delete nonexistent failed: %v", err)
	}
}
