package updater

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/vdesjardins/terraform-module-versions/internal/report"
)

// FileUpdater handles updating .tf files with new module versions
type FileUpdater struct {
}

// NewFileUpdater creates a new file updater
func NewFileUpdater() *FileUpdater {
	return &FileUpdater{}
}

// Update updates all occurrences of a module source from oldVersion to newVersion in a file
// Returns the number of replacements made
func (u *FileUpdater) Update(filePath, source, oldVersion, newVersion string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)

	// Count occurrences before replacement
	countBefore := countOccurrences(contentStr, source, oldVersion)
	if countBefore == 0 {
		return 0, nil
	}

	// Perform replacement
	replacer := NewSimpleVersionReplacer(source, oldVersion, newVersion)
	updated, err := replacer.Replace(contentStr)
	if err != nil {
		return 0, fmt.Errorf("failed to replace version: %w", err)
	}

	if updated == contentStr {
		// No changes made
		return 0, nil
	}

	// Write back atomically
	if err := u.writeAtomically(filePath, []byte(updated)); err != nil {
		return 0, fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return countBefore, nil
}

// WriteDiff outputs a unified diff for the requested update.
func (u *FileUpdater) WriteDiff(writer io.Writer, dirPath, source, oldVersion, newVersion string) error {
	if writer == nil {
		writer = os.Stdout
	}

	return u.walkTerraformFiles(dirPath, func(path string) error {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		contentStr := string(content)
		if countOccurrences(contentStr, source, oldVersion) == 0 {
			return nil
		}

		replacer := NewSimpleVersionReplacer(source, oldVersion, newVersion)
		updated, err := replacer.Replace(contentStr)
		if err != nil {
			return fmt.Errorf("failed to replace version: %w", err)
		}
		if updated == contentStr {
			return nil
		}

		diffOutput, err := report.FormatUnifiedDiff(path, contentStr, updated)
		if err != nil {
			return err
		}
		if diffOutput == "" {
			return nil
		}

		formatted, err := report.RenderOutput(diffOutput)
		if err != nil {
			return err
		}

		_, err = io.WriteString(writer, formatted)
		return err
	})
}

// WriteDiffWithTool outputs diff through an external tool.
func (u *FileUpdater) WriteDiffWithTool(writer io.Writer, dirPath, source, oldVersion, newVersion, tool string) error {
	if err := report.SetOutput(tool); err != nil {
		return err
	}
	return u.WriteDiff(writer, dirPath, source, oldVersion, newVersion)
}

// Count counts occurrences of a module source+version in a file without updating it
// Returns the number of matches found
func (u *FileUpdater) Count(filePath, source, oldVersion string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)
	return countOccurrences(contentStr, source, oldVersion), nil
}

// UpdateDirectory updates all .tf files in a directory tree
// Returns a map of file paths to number of replacements made
func (u *FileUpdater) UpdateDirectory(dirPath, source, oldVersion, newVersion string) (map[string]int, error) {
	results := make(map[string]int)

	err := u.walkTerraformFiles(dirPath, func(path string) error {
		count, err := u.Update(path, source, oldVersion, newVersion)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", path, err)
			return nil
		}

		if count > 0 {
			results[path] = count
		}

		return nil
	})

	return results, err
}

// CountDirectory counts matches in all .tf files in a directory tree without updating
// Returns a map of file paths to number of matches found
func (u *FileUpdater) CountDirectory(dirPath, source, oldVersion string) (map[string]int, error) {
	results := make(map[string]int)

	err := u.walkTerraformFiles(dirPath, func(path string) error {
		count, err := u.Count(path, source, oldVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", path, err)
			return nil
		}

		if count > 0 {
			results[path] = count
		}

		return nil
	})

	return results, err
}

func (u *FileUpdater) walkTerraformFiles(dirPath string, handler func(path string) error) error {
	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !isTerraformFile(path) {
			return nil
		}

		return handler(path)
	})
}

// writeAtomically writes to a file atomically using temp file + rename
func (u *FileUpdater) writeAtomically(filePath string, content []byte) error {
	dir := filepath.Dir(filePath)
	tempFile, err := os.CreateTemp(dir, ".tf-tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up if something fails

	// Write content
	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to ensure data is written
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	tempFile.Close()

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// isTerraformFile checks if a file is a terraform file
func isTerraformFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".tf"
}

// countOccurrences counts how many times a module with a specific version appears in content
func countOccurrences(content, source, version string) int {
	// Simple counting: look for module blocks with this source and version
	// This is approximate but good enough for counting updates
	count := 0
	pos := 0

	for {
		idx := findModuleBlock(content, pos, source, version)
		if idx == -1 {
			break
		}
		count++
		pos = idx + 1
	}

	return count
}

// findModuleBlock finds the next module block with the given source and version
// Returns the position or -1 if not found
func findModuleBlock(content string, startPos int, source, version string) int {
	// Look for a simple pattern: source = "..." version = "..."
	searchContent := content[startPos:]

	sourceIdx := -1
	for i := 0; i < len(searchContent); {
		idx := findStringInContent(searchContent[i:], "source", "\""+source+"\"")
		if idx == -1 {
			break
		}

		realIdx := i + idx
		// Look for version attribute near this source
		// Search within 500 characters
		checkStart := realIdx
		checkEnd := realIdx + 500
		if checkEnd > len(searchContent) {
			checkEnd = len(searchContent)
		}

		checkContent := searchContent[checkStart:checkEnd]
		if findStringInContent(checkContent, "version", "\""+version+"\"") != -1 {
			sourceIdx = realIdx
			break
		}

		i = realIdx + 1
	}

	if sourceIdx == -1 {
		return -1
	}

	return startPos + sourceIdx
}

// findStringInContent finds both attribute and value in content
// Returns position of attribute or -1
func findStringInContent(content, attr, value string) int {
	// Look for attribute = value
	_ = regexp.QuoteMeta(value) // For future use with regex matching
	idx := -1

	for i := 0; i < len(content); {
		if i+len(attr) <= len(content) && content[i:i+len(attr)] == attr {
			// Found attribute, now check if value follows
			checkStart := i + len(attr)
			remaining := content[checkStart:]

			// Skip whitespace and =
			j := 0
			for j < len(remaining) && (remaining[j] == ' ' || remaining[j] == '\t' || remaining[j] == '=') {
				j++
			}

			// Check if value follows
			if j < len(remaining) && remaining[j:j+len(value)] == value {
				idx = i
				break
			}
		}
		i++
	}

	return idx
}
