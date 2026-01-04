package updater

import (
	"fmt"
	"regexp"
	"strings"
)

// versionUpdatePattern matches a module block with a specific source and captures the version attribute
// This pattern:
// 1. Matches module "name" {
// 2. Finds source = "..." within the block
// 3. Finds version = "..." within the block
// 4. Captures the version value to replace it
func buildVersionUpdatePattern(source string) (*regexp.Regexp, error) {
	// Escape special regex characters in the source
	escapedSource := regexp.QuoteMeta(source)

	// Pattern explanation:
	// (?:module\s+"[^"]+"\s+\{) - Matches opening of module block (non-capturing)
	// (.*?) - Captures any content between module opening and source (non-greedy)
	// (?:source\s*=\s*") - Matches the source attribute opening
	// (?:escapedSource) - Matches the exact module source
	// (.*?) - Captures any content between source and version (non-greedy)
	// (?:version\s*=\s*") - Matches the version attribute opening
	// ([^"]*) - Captures the current version value
	// This allows us to replace just the version while preserving everything else

	pattern := fmt.Sprintf(
		`(?:module\s+"[^"]+"\s+\{[\s\S]*?source\s*=\s*"%s"[\s\S]*?version\s*=\s*")([^"]+)(")`,
		escapedSource,
	)

	return regexp.Compile(pattern)
}

// ReplaceVersion replaces the version of a specific module source in the content
// Returns the updated content or error if pattern compilation fails
func ReplaceVersion(content, source, oldVersion, newVersion string) (string, error) {
	pattern, err := buildVersionUpdatePattern(source)
	if err != nil {
		return "", fmt.Errorf("failed to compile pattern for %s: %w", source, err)
	}

	// Find all matches for this source
	matches := pattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		// No matches found - this could be ok if the version is already updated
		return content, nil
	}

	// Replace matches in reverse order to preserve indices
	result := content
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		// match[0:2] = full match start and end
		// match[2:4] = group 1 (current version) start and end
		// match[4:6] = group 2 (closing quote) start and end

		start := match[2]
		end := match[3]

		currentVersion := content[start:end]

		// Only replace if it matches the old version
		if currentVersion == oldVersion {
			result = result[:start] + newVersion + result[end:]
		}
	}

	return result, nil
}

// SimpleVersionReplacer does simple string-based replacement of versions
// This is a more conservative approach that's less likely to have false matches
// Pattern: module "name" { ... source = "source" ... version = "oldVersion" ... }
type SimpleVersionReplacer struct {
	source     string
	oldVersion string
	newVersion string
}

// NewSimpleVersionReplacer creates a new simple replacer
func NewSimpleVersionReplacer(source, oldVersion, newVersion string) *SimpleVersionReplacer {
	return &SimpleVersionReplacer{
		source:     source,
		oldVersion: oldVersion,
		newVersion: newVersion,
	}
}

// Replace performs a conservative replacement
// It looks for the source first, then the version attribute near it
func (r *SimpleVersionReplacer) Replace(content string) (string, error) {
	// Find all module blocks that use this source
	pattern := regexp.MustCompile(fmt.Sprintf(
		`module\s+"([^"]+)"\s+\{([^}]*source\s*=\s*"%s"[^}]*)\}`,
		regexp.QuoteMeta(r.source),
	))

	result := content
	matches := pattern.FindAllStringSubmatchIndex(content, -1)

	// Process in reverse to preserve indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		blockContent := content[match[4]:match[5]] // Group 2: block content

		// Find and replace version in this block
		versionPattern := regexp.MustCompile(fmt.Sprintf(
			`version\s*=\s*"%s"`,
			regexp.QuoteMeta(r.oldVersion),
		))

		newBlockContent := versionPattern.ReplaceAllString(blockContent, fmt.Sprintf(`version = "%s"`, r.newVersion))

		if newBlockContent != blockContent {
			// Replace this module block in the result
			blockPrefixEnd := match[2]
			blockNameEnd := match[3]
			blockName := content[blockPrefixEnd:blockNameEnd]

			oldBlock := fmt.Sprintf("module \"%s\" {%s}", blockName, blockContent)
			newBlock := fmt.Sprintf("module \"%s\" {%s}", blockName, newBlockContent)

			result = strings.Replace(result, oldBlock, newBlock, 1)
		}
	}

	return result, nil
}
