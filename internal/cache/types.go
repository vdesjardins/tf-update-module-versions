package cache

import (
	"time"
)

// Entry represents a cached item with metadata.
type Entry struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	TTL       time.Duration `json:"ttl"`
}

// IsExpired returns true if the entry has expired.
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Store defines the interface for cache storage backends.
type Store interface {
	// Set stores a value with the given key and TTL.
	Set(key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value by key. Returns nil if not found or expired.
	Get(key string) (interface{}, error)

	// Delete removes a key from the cache.
	Delete(key string) error

	// Clear removes all entries from the cache.
	Clear() error

	// Exists checks if a key exists and is not expired.
	Exists(key string) (bool, error)

	// GetExpired returns all expired entries (for cleanup).
	GetExpired() ([]*Entry, error)

	// Close closes the store and releases resources.
	Close() error
}
