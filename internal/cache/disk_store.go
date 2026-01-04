package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DiskStore implements a thread-safe, disk-based cache with TTL support.
type DiskStore struct {
	basePath string
	mu       sync.RWMutex
	entries  map[string]*Entry
	ticker   *time.Ticker
	done     chan struct{}
}

// NewDiskStore creates a new disk-based cache store.
// The cache directory will be created if it doesn't exist.
func NewDiskStore(cacheDir string) (*DiskStore, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	ds := &DiskStore{
		basePath: cacheDir,
		entries:  make(map[string]*Entry),
		done:     make(chan struct{}),
	}

	// Load existing entries from disk
	if err := ds.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load cache from disk: %w", err)
	}

	// Start background cleanup goroutine for expired entries
	ds.ticker = time.NewTicker(5 * time.Minute)
	go ds.cleanupExpired()

	return ds, nil
}

// Set stores a value with the given key and TTL.
func (ds *DiskStore) Set(key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	entry := &Entry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
		TTL:       ttl,
	}

	ds.entries[key] = entry

	// Write to disk
	if err := ds.writeEntryToDisk(key, entry); err != nil {
		// Remove from memory if write fails
		delete(ds.entries, key)
		return fmt.Errorf("failed to write cache entry to disk: %w", err)
	}

	return nil
}

// Get retrieves a value by key. Returns nil if not found or expired.
func (ds *DiskStore) Get(key string) (interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	entry, exists := ds.entries[key]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if entry.IsExpired() {
		return nil, nil
	}

	return entry.Value, nil
}

// Delete removes a key from the cache.
func (ds *DiskStore) Delete(key string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.entries, key)

	// Remove from disk
	cacheFile := filepath.Join(ds.basePath, hashKey(key)+".json")
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// Clear removes all entries from the cache.
func (ds *DiskStore) Clear() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.entries = make(map[string]*Entry)

	// Remove all cache files
	entries, err := os.ReadDir(ds.basePath)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(ds.basePath, entry.Name())); err != nil {
				return fmt.Errorf("failed to delete cache file: %w", err)
			}
		}
	}

	return nil
}

// Exists checks if a key exists and is not expired.
func (ds *DiskStore) Exists(key string) (bool, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	entry, exists := ds.entries[key]
	if !exists {
		return false, nil
	}

	return !entry.IsExpired(), nil
}

// GetExpired returns all expired entries.
func (ds *DiskStore) GetExpired() ([]*Entry, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var expired []*Entry
	for _, entry := range ds.entries {
		if entry.IsExpired() {
			expired = append(expired, entry)
		}
	}

	return expired, nil
}

// Close closes the store and releases resources.
func (ds *DiskStore) Close() error {
	close(ds.done)
	ds.ticker.Stop()
	return nil
}

// writeEntryToDisk writes a single cache entry to disk.
func (ds *DiskStore) writeEntryToDisk(key string, entry *Entry) error {
	cacheFile := filepath.Join(ds.basePath, hashKey(key)+".json")

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// loadFromDisk loads all cache entries from disk.
func (ds *DiskStore) loadFromDisk() error {
	entries, err := os.ReadDir(ds.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet, that's OK
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(ds.basePath, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Log but continue loading other entries
			continue
		}

		var cacheEntry Entry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			// Invalid entry, skip it
			continue
		}

		ds.entries[cacheEntry.Key] = &cacheEntry
	}

	return nil
}

// cleanupExpired periodically removes expired entries from disk.
func (ds *DiskStore) cleanupExpired() {
	for {
		select {
		case <-ds.ticker.C:
			expired, err := ds.GetExpired()
			if err != nil {
				continue
			}

			for _, entry := range expired {
				_ = ds.Delete(entry.Key)
			}

		case <-ds.done:
			return
		}
	}
}

// hashKey creates a filesystem-safe filename from a cache key.
func hashKey(key string) string {
	// Simple hash: use first 32 chars of key, replacing unsafe chars
	if len(key) > 32 {
		key = key[:32]
	}

	// Replace unsafe characters with underscores
	unsafe := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00'}
	result := make([]rune, len(key))
	for i, ch := range key {
		safe := true
		for _, u := range unsafe {
			if ch == u {
				result[i] = '_'
				safe = false
				break
			}
		}
		if safe {
			result[i] = ch
		}
	}

	return string(result)
}
