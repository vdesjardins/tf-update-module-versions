package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDiskStore_Concurrent_Read_Write(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	var wg sync.WaitGroup
	numGoroutines := 50
	operationsPerGoroutine := 100

	// Launch concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", goroutineID, j)
				value := fmt.Sprintf("value_%d_%d", goroutineID, j)
				if err := store.Set(key, value, 1*time.Hour); err != nil {
					t.Errorf("concurrent Set failed: %v", err)
				}
			}
		}(i)
	}

	// Launch concurrent readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", goroutineID, j)
				_, err := store.Get(key)
				if err != nil {
					t.Errorf("concurrent Get failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestDiskStore_Concurrent_Delete(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	numKeys := 100
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := store.Set(key, fmt.Sprintf("value_%d", i), 1*time.Hour); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(keyID int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", keyID)
			if err := store.Delete(key); err != nil {
				t.Errorf("concurrent Delete failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all keys are deleted
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		retrieved, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if retrieved != nil {
			t.Errorf("expected nil for key %s, got %v", key, retrieved)
		}
	}
}

func TestDiskStore_Concurrent_Exists_Check(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	key := "concurrent_key"
	if err := store.Set(key, "data", 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exists, err := store.Exists(key)
			if err != nil {
				t.Errorf("Exists failed: %v", err)
			}
			if !exists {
				t.Errorf("expected key to exist")
			}
		}()
	}

	wg.Wait()
}

func TestDiskStore_Concurrent_Mixed_Operations(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	var wg sync.WaitGroup
	numGoroutines := 30
	operationsPerGoroutine := 50

	// Writers
	for i := 0; i < numGoroutines/3; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d", goroutineID*operationsPerGoroutine+j)
				if err := store.Set(key, fmt.Sprintf("value_%d", j), 1*time.Hour); err != nil {
					t.Errorf("concurrent Set failed: %v", err)
				}
			}
		}(i)
	}

	// Readers
	for i := numGoroutines / 3; i < 2*numGoroutines/3; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d", goroutineID*operationsPerGoroutine+j)
				_, err := store.Get(key)
				if err != nil {
					t.Errorf("concurrent Get failed: %v", err)
				}
			}
		}(i)
	}

	// Deleters
	for i := 2 * numGoroutines / 3; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d", goroutineID*operationsPerGoroutine+j)
				_ = store.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestDiskStore_Race_Condition_Detection(t *testing.T) {
	// This test should be run with -race flag to detect race conditions
	// go test -race ./internal/cache/...

	store, _ := setupTestStore(t)
	defer store.Close()

	var wg sync.WaitGroup
	const iterations = 1000

	// Multiple goroutines performing concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("race_key_%d", id%5)
				value := fmt.Sprintf("race_value_%d", j)

				// Mix of operations
				switch j % 4 {
				case 0:
					_ = store.Set(key, value, 1*time.Hour)
				case 1:
					_, _ = store.Get(key)
				case 2:
					_ = store.Delete(key)
				case 3:
					_, _ = store.Exists(key)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestDiskStore_Concurrent_Expiration_Cleanup(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Add many short-lived entries
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("expiring_key_%d", i)
		if err := store.Set(key, fmt.Sprintf("value_%d", i), 50*time.Millisecond); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Concurrent reads while entries are expiring
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("expiring_key_%d", j)
				_, _ = store.Get(key)
			}
		}()
	}

	wg.Wait()

	// All entries should be expired
	expired, err := store.GetExpired()
	if err != nil {
		t.Fatalf("GetExpired failed: %v", err)
	}
	if len(expired) != 100 {
		t.Errorf("expected 100 expired entries, got %d", len(expired))
	}
}

func TestDiskStore_Concurrent_Clear_And_Set(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Pre-populate cache
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := store.Set(key, fmt.Sprintf("value_%d", i), 1*time.Hour); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	var wg sync.WaitGroup

	// Goroutine that clears the cache
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		if err := store.Clear(); err != nil {
			t.Errorf("Clear failed: %v", err)
		}
	}()

	// Goroutines that continue adding data
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				if err := store.Set(key, fmt.Sprintf("value_%d_%d", id, j), 1*time.Hour); err != nil {
					t.Errorf("Set failed: %v", err)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}
