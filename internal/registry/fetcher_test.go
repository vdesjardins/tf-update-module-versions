package registry

import (
	"context"
	"testing"
)

func TestNewVersionFetcher(t *testing.T) {
	fetcher := NewVersionFetcher(4)
	if fetcher == nil {
		t.Fatal("NewVersionFetcher returned nil")
	}
	if fetcher.workers != 4 {
		t.Errorf("workers = %d, want 4", fetcher.workers)
	}
}

func TestNewVersionFetcherDefaultWorkers(t *testing.T) {
	fetcher := NewVersionFetcher(0)
	if fetcher.workers != 4 {
		t.Errorf("workers = %d, want 4", fetcher.workers)
	}
}

func TestVersionFetcherImplementsInterface(t *testing.T) {
	var _ = (*VersionFetcher)(nil)
	// This test just ensures the type compiles
}

func TestGetResult(t *testing.T) {
	fetcher := NewVersionFetcher(4)

	// Manually set a result
	fetcher.resultsMu.Lock()
	fetcher.results["test/module/aws"] = []string{"1.0.0", "2.0.0"}
	fetcher.resultsMu.Unlock()

	result := fetcher.GetResult("test", "module", "aws")
	if len(result) != 2 {
		t.Errorf("GetResult returned %d versions, want 2", len(result))
	}
	if result[0] != "1.0.0" {
		t.Errorf("GetResult[0] = %s, want 1.0.0", result[0])
	}
}

func TestErrors(t *testing.T) {
	fetcher := NewVersionFetcher(4)

	// Manually set an error
	fetcher.errorsMu.Lock()
	fetcher.errors["test/module/aws"] = context.Canceled
	fetcher.errorsMu.Unlock()

	errors := fetcher.Errors()
	if _, ok := errors["test/module/aws"]; !ok {
		t.Error("Errors() did not return expected error")
	}
}
