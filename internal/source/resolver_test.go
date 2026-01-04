package source

import (
	"testing"
)

func TestResolverTerraformRegistry(t *testing.T) {
	resolver := NewResolver()

	tests := []struct {
		name         string
		source       string
		expectedType SourceTypeEnum
		expectedHost string
		expectedNS   string
		expectedName string
		expectedProv string
		supported    bool
	}{
		{
			name:         "Terraform Registry",
			source:       "hashicorp/vault-starter/aws",
			expectedType: SourceTypeTerraformRegistry,
			expectedHost: "registry.terraform.io",
			expectedNS:   "hashicorp",
			expectedName: "vault-starter",
			expectedProv: "aws",
			supported:    true,
		},
		{
			name:         "Custom Registry",
			source:       "my-registry.com/team/module/aws",
			expectedType: SourceTypeCustomRegistry,
			expectedHost: "my-registry.com",
			expectedNS:   "team",
			expectedName: "module",
			expectedProv: "aws",
			supported:    true,
		},
		{
			name:         "GitHub source",
			source:       "github.com/hashicorp/example",
			expectedType: SourceTypeGitHub,
			expectedHost: "github.com",
			supported:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.Resolve(tt.source)
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}

			if result.Host != tt.expectedHost {
				t.Errorf("Host = %v, want %v", result.Host, tt.expectedHost)
			}

			if tt.expectedType == SourceTypeTerraformRegistry || tt.expectedType == SourceTypeCustomRegistry {
				if result.Namespace != tt.expectedNS {
					t.Errorf("Namespace = %v, want %v", result.Namespace, tt.expectedNS)
				}
				if result.Name != tt.expectedName {
					t.Errorf("Name = %v, want %v", result.Name, tt.expectedName)
				}
				if result.Provider != tt.expectedProv {
					t.Errorf("Provider = %v, want %v", result.Provider, tt.expectedProv)
				}
			}

			if result.Supported != tt.supported {
				t.Errorf("Supported = %v, want %v", result.Supported, tt.supported)
			}
		})
	}
}

func TestResolverInvalidSources(t *testing.T) {
	resolver := NewResolver()

	tests := []struct {
		name        string
		source      string
		expectError bool
	}{
		{"empty source", "", true},
		{"too few parts", "namespace/name", true},
		{"invalid github", "github.com/owner", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolver.Resolve(tt.source)
			if (err != nil) != tt.expectError {
				t.Errorf("Resolve() error = %v, wantError = %v", err, tt.expectError)
			}
		})
	}
}

func TestRegistryPath(t *testing.T) {
	resolver := NewResolver()

	source, _ := resolver.Resolve("hashicorp/vault/aws")
	path := source.RegistryPath()

	expected := "hashicorp/vault/aws"
	if path != expected {
		t.Errorf("RegistryPath() = %s, want %s", path, expected)
	}
}
