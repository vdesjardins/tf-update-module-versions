package source

import (
	"fmt"
	"strings"
)

// Resolver detects the type of a module source and creates appropriate handler
type Resolver struct{}

// NewResolver creates a new source resolver
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve analyzes a module source string and returns a Source with type detection
func (r *Resolver) Resolve(sourceStr string) (*Source, error) {
	if sourceStr == "" {
		return nil, fmt.Errorf("empty source string")
	}

	// Handle GitHub URLs
	if strings.HasPrefix(sourceStr, "github.com/") {
		parts := strings.Split(strings.TrimPrefix(sourceStr, "github.com/"), "//")
		paths := strings.Split(parts[0], "/")
		if len(paths) < 2 {
			return nil, fmt.Errorf("invalid github source format: %s", sourceStr)
		}
		return &Source{
			Original:  sourceStr,
			Type:      SourceTypeGitHub,
			Host:      "github.com",
			Path:      parts[0],
			Supported: false,
		}, nil
	}

	// Handle registry sources
	// Format: [host/]namespace/name/provider[//path]
	return r.parseRegistrySource(sourceStr)
}

// parseRegistrySource handles both Terraform registry and custom registries
func (r *Resolver) parseRegistrySource(sourceStr string) (*Source, error) {
	// Split by // to separate module path from subdirectory
	parts := strings.Split(sourceStr, "//")
	modulePath := parts[0]
	var subPath string
	if len(parts) > 1 {
		subPath = strings.Join(parts[1:], "//")
	}

	// Split by / to get components
	pathParts := strings.Split(modulePath, "/")

	var host, namespace, name, provider string

	if len(pathParts) == 3 {
		// Terraform Registry format: namespace/name/provider
		host = "registry.terraform.io"
		namespace = pathParts[0]
		name = pathParts[1]
		provider = pathParts[2]
	} else if len(pathParts) >= 4 {
		// Custom Registry format: host/namespace/name/provider[/...]
		host = pathParts[0]
		namespace = pathParts[1]
		name = pathParts[2]
		provider = pathParts[3]
	} else {
		return nil, fmt.Errorf("invalid registry source format: %s", sourceStr)
	}

	source := &Source{
		Original:  sourceStr,
		Type:      SourceTypeCustomRegistry,
		Host:      host,
		Namespace: namespace,
		Name:      name,
		Provider:  provider,
		Path:      subPath,
		Supported: true, // All registry sources are supported
	}

	// Mark as Terraform Registry if using official host
	if host == "registry.terraform.io" {
		source.Type = SourceTypeTerraformRegistry
	}

	return source, nil
}

// String returns the canonical source string
func (s *Source) String() string {
	return s.Original
}

// RegistryPath returns the registry API path (namespace/name/provider)
// Only valid for registry sources
func (s *Source) RegistryPath() string {
	if s.Namespace == "" || s.Name == "" || s.Provider == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s.Namespace, s.Name, s.Provider)
}
