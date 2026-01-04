package source

import "context"

// SourceTypeEnum represents the type of module source
type SourceTypeEnum int

const (
	SourceTypeTerraformRegistry SourceTypeEnum = iota // registry.terraform.io
	SourceTypeCustomRegistry                          // custom.registry.com
	SourceTypeGitHub                                  // github.com/...
	SourceTypeUnknown
)

func (e SourceTypeEnum) String() string {
	switch e {
	case SourceTypeTerraformRegistry:
		return "Terraform Registry"
	case SourceTypeCustomRegistry:
		return "Custom Registry"
	case SourceTypeGitHub:
		return "GitHub"
	default:
		return "Unknown"
	}
}

// Source represents a parsed module source with type information
type Source struct {
	Original  string         // Original source string as specified in terraform
	Type      SourceTypeEnum // Detected source type
	Host      string         // Registry host (e.g., "registry.terraform.io" or "github.com")
	Namespace string         // e.g., "hashicorp" (for registries)
	Name      string         // e.g., "vault-starter" (for registries)
	Provider  string         // e.g., "aws" (for registries)
	Path      string         // Registry subdirectory or repo path
	Supported bool           // Whether we can fetch versions from this source
}

// SourceHandler is the interface for different module source types
type SourceHandler interface {
	// Type returns the enum type of this source
	Type() SourceTypeEnum

	// Host returns the registry/service host
	Host() string

	// IsSupported returns whether we can fetch versions from this source
	IsSupported() bool

	// FetchLatestVersions fetches available versions from the source
	// Returns versions in descending order (latest first)
	FetchLatestVersions(ctx context.Context) ([]string, error)

	// Source returns the underlying Source struct
	Source() *Source
}
