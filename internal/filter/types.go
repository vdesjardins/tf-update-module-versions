package filter

// ModuleFilter defines filtering and version update strategy
type ModuleFilter struct {
	// Module-specific filters (from --module flag)
	ModulePatterns map[string]string // pattern → version_type (minor/latest)

	// Global version strategy (from --version flag)
	GlobalVersion string // "minor", "latest", or ""

	// If true, warn when no modules match a pattern
	WarnUnmatched bool
}

// GetVersionStrategy returns the version strategy for a given module source
// Returns (version_type, matched bool)
func (mf *ModuleFilter) GetVersionStrategy(moduleSource string) (string, bool) {
	// If global version is set, use it for all modules
	if mf.GlobalVersion != "" {
		return mf.GlobalVersion, true
	}

	// Otherwise check module-specific patterns
	for pattern, versionType := range mf.ModulePatterns {
		matched, _ := MatchModule(moduleSource, pattern)
		if matched {
			return versionType, true
		}
	}

	return "", false
}
