package version

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Strategy type for version selection
type Strategy string

const (
	StrategyMinor  Strategy = "minor"
	StrategyLatest Strategy = "latest"
)

// SelectVersion chooses the appropriate version from available versions
// based on the strategy, current version, and optional constraints.
// If constraints are provided, only versions satisfying all constraints are considered.
//
// Arguments:
//   - currentVersion: The currently installed version
//   - availableVersions: List of available versions (assumed to be sorted in descending order)
//   - strategy: The version selection strategy (StrategyLatest or StrategyMinor)
//   - constraints: Optional constraints to filter versions (nil = no filtering)
//
// Returns:
//   - The selected version string
//   - An error if no versions are available or match the criteria
func SelectVersion(currentVersion string, availableVersions []string, strategy Strategy, constraints Constraints) (string, error) {
	if len(availableVersions) == 0 {
		return "", fmt.Errorf("no available versions")
	}

	// Filter versions by constraints if provided
	candidateVersions := availableVersions
	if len(constraints) > 0 {
		candidateVersions = filterVersionsByConstraints(availableVersions, constraints)
		if len(candidateVersions) == 0 {
			return "", fmt.Errorf("no versions satisfy the constraints")
		}
	}

	switch strategy {
	case StrategyLatest:
		// Return the highest version (assumed to be first after sorting in descending order)
		return candidateVersions[0], nil

	case StrategyMinor:
		// Return latest version matching current major version
		return selectMinorVersion(currentVersion, candidateVersions)

	default:
		return "", fmt.Errorf("unknown version strategy: %s", strategy)
	}
}

// filterVersionsByConstraints filters a list of version strings through constraint evaluation.
// Returns only versions that satisfy all constraints (AND semantics).
func filterVersionsByConstraints(versions []string, constraints Constraints) []string {
	if len(constraints) == 0 {
		return versions
	}

	var filtered []string
	for _, versionStr := range versions {
		if constraints.MatchesString(versionStr) {
			filtered = append(filtered, versionStr)
		}
	}
	return filtered
}

// selectMinorVersion finds the latest version matching the current major version
// Example: currentVersion="1.2.3", availableVersions=["2.0.0", "1.5.2", "1.5.1", "0.9.0"]
// Returns: "1.5.2"
func selectMinorVersion(currentVersion string, availableVersions []string) (string, error) {
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", fmt.Errorf("invalid current version %q: %w", currentVersion, err)
	}

	currentMajor := current.Major()

	for _, availableStr := range availableVersions {
		available, err := semver.NewVersion(availableStr)
		if err != nil {
			// Skip invalid versions
			continue
		}

		if available.Major() == currentMajor {
			return availableStr, nil
		}
	}

	// No matching major version found
	return "", fmt.Errorf("no version found matching major version %d", currentMajor)
}

// IsValidStrategy checks if a strategy string is valid
func IsValidStrategy(s string) bool {
	return s == string(StrategyMinor) || s == string(StrategyLatest)
}
