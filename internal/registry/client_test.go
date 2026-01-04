package registry

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vdesjardins/terraform-module-versions/internal/cache"
)

// MockStore is a simple in-memory cache store for testing
type MockStore struct {
	data map[string]interface{}
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]interface{}),
	}
}

func (m *MockStore) Set(key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockStore) Get(key string) (interface{}, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (m *MockStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStore) Clear() error {
	m.data = make(map[string]interface{})
	return nil
}

func (m *MockStore) Exists(key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockStore) GetExpired() ([]*cache.Entry, error) {
	return []*cache.Entry{}, nil
}

func (m *MockStore) Close() error {
	return nil
}

var _ cache.Store = (*MockStore)(nil)

func TestNewClientWithCache(t *testing.T) {
	mockStore := NewMockStore()
	client := NewClientWithCache(mockStore)

	if client.store == nil {
		t.Error("Client should have cache store set")
	}

	if client.store != mockStore {
		t.Error("Client should use provided cache store")
	}
}

func TestNewClientWithoutCache(t *testing.T) {
	client := NewClient()

	if client.store != nil {
		t.Error("Client should not have cache store when created with NewClient")
	}
}

func TestCacheStorageAndRetrieval(t *testing.T) {
	mockStore := NewMockStore()

	// Test storing a module
	testModule := &Module{
		Source: "terraform-aws-modules/vpc/aws",
		Versions: []*Version{
			{
				Version: "5.0.0",
				Root: RootModule{
					Providers: []Provider{},
				},
			},
		},
	}

	cacheKey := "module_versions:terraform-aws-modules:vpc:aws"
	moduleData, err := json.Marshal(testModule)
	if err != nil {
		t.Fatalf("Failed to marshal module: %v", err)
	}

	// Store in cache
	if err := mockStore.Set(cacheKey, moduleData, 24*time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Retrieve from cache
	cachedData, err := mockStore.Get(cacheKey)
	if err != nil {
		t.Fatalf("Failed to get from cache: %v", err)
	}

	if cachedData == nil {
		t.Error("Expected data to be in cache")
	}

	// Verify we can unmarshal the cached data (it's already bytes)
	var retrievedModule Module
	if err := json.Unmarshal(cachedData.([]byte), &retrievedModule); err != nil {
		t.Fatalf("Failed to unmarshal cached data: %v", err)
	}

	if retrievedModule.Source != testModule.Source {
		t.Errorf("Expected source %s, got %s", testModule.Source, retrievedModule.Source)
	}
}

func TestCacheCheckWithoutStore(t *testing.T) {
	// Client without store should still work
	client := NewClient()

	if client.store != nil {
		t.Error("Client created without store should have nil store")
	}

	// Store field should be nil but accessible
	if client.httpClient == nil {
		t.Error("Client should have httpClient")
	}
}
