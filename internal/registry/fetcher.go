package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/vdesjardins/terraform-module-versions/internal/source"
	"github.com/vdesjardins/terraform-module-versions/internal/version"
)

// VersionFetcher implements source.VersionFetcher interface
var _ source.VersionFetcher = (*VersionFetcher)(nil)

// VersionFetcher fetches versions from registries, with parallel support
type VersionFetcher struct {
	client    *Client
	workers   int
	results   map[string][]string
	resultsMu sync.RWMutex
	workerSem chan struct{}
	errors    map[string]error
	errorsMu  sync.RWMutex
}

// NewVersionFetcher creates a new fetcher with worker pool
func NewVersionFetcher(workerCount int) *VersionFetcher {
	if workerCount <= 0 {
		workerCount = 4 // Default worker pool size
	}

	return &VersionFetcher{
		client:    NewClient(),
		workers:   workerCount,
		results:   make(map[string][]string),
		errors:    make(map[string]error),
		workerSem: make(chan struct{}, workerCount),
	}
}

// NewVersionFetcherWithClient creates a new fetcher with a custom client and worker pool
func NewVersionFetcherWithClient(client *Client, workerCount int) *VersionFetcher {
	if workerCount <= 0 {
		workerCount = 4 // Default worker pool size
	}

	if client == nil {
		client = NewClient()
	}

	return &VersionFetcher{
		client:    client,
		workers:   workerCount,
		results:   make(map[string][]string),
		errors:    make(map[string]error),
		workerSem: make(chan struct{}, workerCount),
	}
}

// FetchVersions implements the source.VersionFetcher interface
func (f *VersionFetcher) FetchVersions(ctx context.Context, host, namespace, name, provider string) ([]string, error) {
	moduleKey := fmt.Sprintf("%s/%s/%s", namespace, name, provider)

	// Acquire worker slot
	select {
	case f.workerSem <- struct{}{}:
		defer func() { <-f.workerSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Fetch from registry
	module, err := f.client.FetchModuleVersions(ctx, host, namespace, name, provider)
	if err != nil {
		f.errorsMu.Lock()
		f.errors[moduleKey] = err
		f.errorsMu.Unlock()
		return nil, err
	}

	if module == nil {
		f.resultsMu.Lock()
		f.results[moduleKey] = []string{}
		f.resultsMu.Unlock()
		return []string{}, nil
	}

	// Fetch detailed info for versions
	if err := f.client.FetchModuleInfo(ctx, host, namespace, name, provider, module); err != nil {
		f.errorsMu.Lock()
		f.errors[moduleKey] = err
		f.errorsMu.Unlock()
		// Don't return error - we have version info even without detailed metadata
	}

	// Extract version strings and sort them (latest first)
	versionStrings := make([]string, len(module.Versions))
	for i, v := range module.Versions {
		versionStrings[i] = v.Version
	}

	// Sort versions - latest first
	sortedVersions, err := version.SortVersions(versionStrings)
	if err != nil {
		// If sorting fails, return unsorted versions
		sortedVersions = versionStrings
	}

	f.resultsMu.Lock()
	f.results[moduleKey] = sortedVersions
	f.resultsMu.Unlock()

	return sortedVersions, nil
}

// FetchMultipleVersions fetches versions for multiple modules in parallel
func (f *VersionFetcher) FetchMultipleVersions(ctx context.Context, modules []*source.Source) map[string][]string {
	var wg sync.WaitGroup

	for _, mod := range modules {
		if !mod.Supported {
			continue
		}

		wg.Add(1)
		go func(m *source.Source) {
			defer wg.Done()
			_, _ = f.FetchVersions(ctx, m.Host, m.Namespace, m.Name, m.Provider)
		}(mod)
	}

	wg.Wait()

	f.resultsMu.RLock()
	defer f.resultsMu.RUnlock()

	resultsCopy := make(map[string][]string)
	for k, v := range f.results {
		resultsCopy[k] = append([]string{}, v...)
	}

	return resultsCopy
}

// Errors returns all errors encountered during fetching
func (f *VersionFetcher) Errors() map[string]error {
	f.errorsMu.RLock()
	defer f.errorsMu.RUnlock()

	errorsCopy := make(map[string]error)
	for k, v := range f.errors {
		errorsCopy[k] = v
	}

	return errorsCopy
}

// GetResult returns the result for a specific module
func (f *VersionFetcher) GetResult(namespace, name, provider string) []string {
	moduleKey := fmt.Sprintf("%s/%s/%s", namespace, name, provider)
	f.resultsMu.RLock()
	defer f.resultsMu.RUnlock()

	return f.results[moduleKey]
}
