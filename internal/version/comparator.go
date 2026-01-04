package version

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
)

// CompareVersions compares two semantic versions.
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
//
// Returns error if versions are not valid semver
func CompareVersions(v1, v2 string) (int, error) {
	sv1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %w", v1, err)
	}

	sv2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %w", v2, err)
	}

	return sv1.Compare(sv2), nil
}

// IsNewer returns true if v2 is newer than v1
func IsNewer(v1, v2 string) (bool, error) {
	cmp, err := CompareVersions(v1, v2)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

// IsSameOrNewer returns true if v2 is same or newer than v1
func IsSameOrNewer(v1, v2 string) (bool, error) {
	cmp, err := CompareVersions(v1, v2)
	if err != nil {
		return false, err
	}
	return cmp <= 0, nil
}

// SortVersions sorts a slice of version strings in descending order (latest first)
// Returns error if any version is invalid semver
func SortVersions(versions []string) ([]string, error) {
	// Convert to semver
	semverVersions := make([]*semver.Version, len(versions))
	for i, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			return nil, fmt.Errorf("invalid version %s: %w", v, err)
		}
		semverVersions[i] = sv
	}

	// Sort using semver collection (ascending order)
	sort.Sort(semver.Collection(semverVersions))

	// Reverse to get latest first (descending order)
	result := make([]string, len(semverVersions))
	for i, sv := range semverVersions {
		result[len(result)-1-i] = sv.String()
	}

	return result, nil
}

// GetLatestVersion returns the latest version from a slice
// Returns error if slice is empty or contains invalid versions
func GetLatestVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions provided")
	}

	sorted, err := SortVersions(versions)
	if err != nil {
		return "", err
	}

	return sorted[0], nil
}
