package mod

import (
	"os"
	"testing"
)

// TestFindModulesWithNoTfFiles ensures that findModules returns no modules
// when the directory contains no .tf files.
func TestFindModulesWithNoTfFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-no-tf-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	mods, err := FindModules(dir)
	if err != nil {
		t.Fatalf("findModules returned error: %v", err)
	}
	if len(mods) != 0 {
		t.Fatalf("expected 0 modules, got %d", len(mods))
	}
}

// TestUpstreamVersionsUnsupportedSource ensures that UpstreamVersions returns an
// error for unsupported source types.
func TestUpstreamVersionsUnsupportedSource(t *testing.T) {
	_, err := UpstreamVersions("invalid/source")
	if err == nil {
		t.Fatalf("expected error for unsupported source")
	}
}

// TestUpstreamVersionsHashicorpInvalidFormat checks that an error is
// returned when the hashicorp source format is not as expected.
func TestUpstreamVersionsHashicorpInvalidFormat(t *testing.T) {
	_, err := UpstreamVersions("hashicorp/aws") // missing provider
	if err == nil {
		t.Fatalf("expected error for invalid hashicorp source")
	}
}
