package finder

import (
	"os"
	"testing"
)

func TestFindModulesWithVersions(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-finder-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	mods, err := FindModulesWithVersions(dir, nil)
	if err != nil {
		t.Fatalf("FindModulesWithVersions returned error: %v", err)
	}

	if len(mods) != 0 {
		t.Fatalf("expected 0 modules in empty dir, got %d", len(mods))
	}
}

func TestFindAllModules(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-finder-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	mods, err := FindAllModules(dir)
	if err != nil {
		t.Fatalf("FindAllModules returned error: %v", err)
	}

	if len(mods) != 0 {
		t.Fatalf("expected 0 modules in empty dir, got %d", len(mods))
	}
}
