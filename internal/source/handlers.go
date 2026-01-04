package source

import (
	"context"
	"fmt"
)

// RegistryHandler implements SourceHandler for registry sources (both Terraform and Custom)
type RegistryHandler struct {
	source  *Source
	fetcher VersionFetcher
}

// VersionFetcher is an interface for fetching versions from registries
// This allows for dependency injection in tests
type VersionFetcher interface {
	FetchVersions(ctx context.Context, host, namespace, name, provider string) ([]string, error)
}

// NewRegistryHandler creates a handler for registry sources
func NewRegistryHandler(source *Source, fetcher VersionFetcher) *RegistryHandler {
	return &RegistryHandler{
		source:  source,
		fetcher: fetcher,
	}
}

func (h *RegistryHandler) Type() SourceTypeEnum {
	return h.source.Type
}

func (h *RegistryHandler) Host() string {
	return h.source.Host
}

func (h *RegistryHandler) IsSupported() bool {
	return true
}

func (h *RegistryHandler) FetchLatestVersions(ctx context.Context) ([]string, error) {
	if h.fetcher == nil {
		return nil, fmt.Errorf("no fetcher configured")
	}
	return h.fetcher.FetchVersions(ctx, h.source.Host, h.source.Namespace, h.source.Name, h.source.Provider)
}

func (h *RegistryHandler) Source() *Source {
	return h.source
}

// GitHubHandler implements SourceHandler for GitHub sources (not supported)
type GitHubHandler struct {
	source *Source
}

// NewGitHubHandler creates a handler for GitHub sources
func NewGitHubHandler(source *Source) *GitHubHandler {
	return &GitHubHandler{
		source: source,
	}
}

func (h *GitHubHandler) Type() SourceTypeEnum {
	return h.source.Type
}

func (h *GitHubHandler) Host() string {
	return h.source.Host
}

func (h *GitHubHandler) IsSupported() bool {
	return false
}

func (h *GitHubHandler) FetchLatestVersions(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("GitHub sources not yet supported (git refs, not semantic versions)")
}

func (h *GitHubHandler) Source() *Source {
	return h.source
}
