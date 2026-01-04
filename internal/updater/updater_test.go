package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplacer(t *testing.T) {
	content := `module "example" {
  source  = "hashicorp/vault/aws"
  version = "0.1.0"
}

module "another" {
  source  = "hashicorp/vault/aws"
  version = "0.1.0"
}`

	replacer := NewSimpleVersionReplacer("hashicorp/vault/aws", "0.1.0", "0.2.0")
	result, err := replacer.Replace(content)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	// Should have replaced both versions
	if result == content {
		t.Error("Replace() made no changes")
	}

	// Check that new version is present
	if !contains(result, "0.2.0") {
		t.Error("Replace() did not update version to 0.2.0")
	}
}

func TestFileUpdater(t *testing.T) {
	// Create temp dir with a .tf file
	tmpDir, err := os.MkdirTemp("", "test-updater-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test .tf file
	tfFile := filepath.Join(tmpDir, "main.tf")
	content := []byte(`module "example" {
  source  = "hashicorp/vault/aws"
  version = "0.1.0"
}`)

	if err := os.WriteFile(tfFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	updater := NewFileUpdater()
	count, err := updater.Update(tfFile, "hashicorp/vault/aws", "0.1.0", "0.2.0")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if count != 1 {
		t.Errorf("Update() = %d replacements, want 1", count)
	}

	// Verify file was updated
	updatedContent, err := os.ReadFile(tfFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	if !contains(string(updatedContent), "0.2.0") {
		t.Error("File was not updated with new version")
	}
}

func TestIsTerraformFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"main.tf", true},
		{"variables.tf", true},
		{"outputs.tf", true},
		{"main.json", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		result := isTerraformFile(tt.path)
		if result != tt.expected {
			t.Errorf("isTerraformFile(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
