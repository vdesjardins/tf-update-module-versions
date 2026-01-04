package source

import (
	"context"
	"testing"
)

// MockVersionFetcher is a mock implementation of VersionFetcher for testing
type MockVersionFetcher struct {
	versions map[string][]string
	err      error
}

func (m *MockVersionFetcher) FetchVersions(ctx context.Context, host, namespace, name, provider string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := namespace + "/" + name + "/" + provider
	return m.versions[key], nil
}

func TestRegistryHandler(t *testing.T) {
	source := &Source{
		Original:  "hashicorp/vault/aws",
		Type:      SourceTypeTerraformRegistry,
		Host:      "registry.terraform.io",
		Namespace: "hashicorp",
		Name:      "vault",
		Provider:  "aws",
		Supported: true,
	}

	mockFetcher := &MockVersionFetcher{
		versions: map[string][]string{
			"hashicorp/vault/aws": {"1.0.0", "1.1.0", "2.0.0"},
		},
	}

	handler := NewRegistryHandler(source, mockFetcher)

	if !handler.IsSupported() {
		t.Error("Registry handler should be supported")
	}

	if handler.Host() != "registry.terraform.io" {
		t.Errorf("Host = %s, want registry.terraform.io", handler.Host())
	}

	versions, err := handler.FetchLatestVersions(context.Background())
	if err != nil {
		t.Fatalf("FetchLatestVersions error = %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("FetchLatestVersions returned %d versions, want 3", len(versions))
	}
}

func TestGitHubHandler(t *testing.T) {
	source := &Source{
		Original:  "github.com/hashicorp/example",
		Type:      SourceTypeGitHub,
		Host:      "github.com",
		Supported: false,
	}

	handler := NewGitHubHandler(source)

	if handler.IsSupported() {
		t.Error("GitHub handler should not be supported")
	}

	_, err := handler.FetchLatestVersions(context.Background())
	if err == nil {
		t.Error("FetchLatestVersions should return an error for GitHub sources")
	}
}
