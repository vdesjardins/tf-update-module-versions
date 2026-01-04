package filter

import (
	"fmt"
	"regexp"
	"strings"
)

type MatchMode int

const (
	MatchModeExact MatchMode = iota
	MatchModeRegex
)

// Matcher determines if a module source matches a pattern
type Matcher struct {
	Pattern string
	Mode    MatchMode
	regex   *regexp.Regexp // Compiled regex if Mode == MatchModeRegex
}

// NewMatcher creates a matcher from a pattern string
// Detects regex patterns by checking for regex metacharacters
// Otherwise treats as exact match
func NewMatcher(pattern string) (*Matcher, error) {
	m := &Matcher{Pattern: pattern}

	// Check if pattern contains regex metacharacters
	if hasRegexMetacharacters(pattern) {
		// Attempt regex compilation
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
		}
		m.Mode = MatchModeRegex
		m.regex = compiled
	} else {
		// No metacharacters, treat as exact match
		m.Mode = MatchModeExact
	}

	return m, nil
}

// hasRegexMetacharacters checks if a pattern contains regex metacharacters
func hasRegexMetacharacters(pattern string) bool {
	metacharacters := []string{"*", "+", "?", "[", "]", "(", ")", "{", "}", "^", "$", "|", "\\", "."}
	for _, char := range metacharacters {
		if strings.Contains(pattern, char) {
			// "." might be part of a domain name, but if combined with other chars, it's likely regex
			// For now, we consider it a metacharacter
			return true
		}
	}
	return false
}

// Matches returns true if moduleSource matches the pattern
func (m *Matcher) Matches(moduleSource string) bool {
	switch m.Mode {
	case MatchModeRegex:
		return m.regex.MatchString(moduleSource)
	case MatchModeExact:
		return m.Pattern == moduleSource
	default:
		return false
	}
}

// MatchModule is a convenience function
func MatchModule(moduleSource, pattern string) (bool, error) {
	matcher, err := NewMatcher(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return matcher.Matches(moduleSource), nil
}
