package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vdesjardins/terraform-module-versions/internal/cache"
	"github.com/vdesjardins/terraform-module-versions/internal/version"
)

// Client is an HTTP client for registry API calls
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
	store      cache.Store
}

// NewClient creates a new registry client with configured timeout
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		timeout:    time.Duration(version.RegistryTimeout) * time.Second,
		store:      nil,
	}
}

// NewClientWithCache creates a new registry client with a cache store
func NewClientWithCache(store cache.Store) *Client {
	return &Client{
		httpClient: &http.Client{},
		timeout:    time.Duration(version.RegistryTimeout) * time.Second,
		store:      store,
	}
}

// FetchModuleVersions fetches all versions for a module from the registry
func (c *Client) FetchModuleVersions(ctx context.Context, registryHost, namespace, name, provider string) (*Module, error) {
	cacheKey := fmt.Sprintf("module_versions:%s:%s:%s:%s", registryHost, namespace, name, provider)

	// Check cache first if store is available
	if c.store != nil {
		if cachedData, err := c.store.Get(cacheKey); err == nil && cachedData != nil {
			// Convert interface{} to JSON bytes for unmarshaling
			if jsonBytes, err := json.Marshal(cachedData); err == nil {
				var module Module
				if err := json.Unmarshal(jsonBytes, &module); err == nil {
					return &module, nil
				}
			}
		}
	}

	apiURL := fmt.Sprintf("https://%s/v1/modules/%s/%s/%s/versions", registryHost, namespace, name, provider)

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry API returned %d for %s/%s/%s", resp.StatusCode, namespace, name, provider)
	}

	var payload registryResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(payload.Modules) == 0 {
		return nil, nil
	}

	module := &payload.Modules[0]

	// Store in cache if available
	if c.store != nil {
		if data, err := json.Marshal(module); err == nil {
			// Cache for 24 hours
			c.store.Set(cacheKey, data, 24*time.Hour)
		}
	}

	return module, nil
}

// FetchModuleInfo fetches detailed info for specific module versions
func (c *Client) FetchModuleInfo(ctx context.Context, registryHost, namespace, name, provider string, module *Module) error {
	for _, v := range module.Versions {
		cacheKey := fmt.Sprintf("module_info:%s:%s:%s:%s:%s", registryHost, namespace, name, provider, v.Version)

		// Check cache first if store is available
		var info ModuleInfo
		var cacheHit bool
		if c.store != nil {
			if cachedData, err := c.store.Get(cacheKey); err == nil && cachedData != nil {
				// Convert interface{} to JSON bytes for unmarshaling
				if jsonBytes, err := json.Marshal(cachedData); err == nil {
					if err := json.Unmarshal(jsonBytes, &info); err == nil {
						v.RegistryModuleInfo = &info
						cacheHit = true
					}
				}
			}
		}

		if cacheHit {
			continue
		}

		apiURL := fmt.Sprintf("https://%s/v1/modules/%s/%s/%s/%s", registryHost, namespace, name, provider, v.Version)

		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctxWithTimeout, "GET", apiURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("registry API call failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("registry API returned %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return fmt.Errorf("failed to decode module info: %w", err)
		}

		v.RegistryModuleInfo = &info

		// Store in cache if available
		if c.store != nil {
			if data, err := json.Marshal(info); err == nil {
				// Cache for 24 hours
				c.store.Set(cacheKey, data, 24*time.Hour)
			}
		}
	}

	return nil
}
