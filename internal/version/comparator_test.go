package version

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name      string
		v1        string
		v2        string
		expected  int
		wantError bool
	}{
		{"v1 < v2", "1.0.0", "2.0.0", -1, false},
		{"v1 > v2", "2.0.0", "1.0.0", 1, false},
		{"v1 == v2", "1.0.0", "1.0.0", 0, false},
		{"patch version", "1.0.1", "1.0.0", 1, false},
		{"minor version", "1.1.0", "1.0.0", 1, false},
		{"prerelease", "1.0.0-alpha", "1.0.0", -1, false},
		{"invalid v1", "invalid", "1.0.0", 0, true},
		{"invalid v2", "1.0.0", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)
			if (err != nil) != tt.wantError {
				t.Errorf("CompareVersions() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("CompareVersions() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name      string
		v1        string
		v2        string
		expected  bool
		wantError bool
	}{
		{"2.0.0 is newer than 1.0.0", "1.0.0", "2.0.0", true, false},
		{"1.0.0 is not newer than 2.0.0", "2.0.0", "1.0.0", false, false},
		{"same version", "1.0.0", "1.0.0", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsNewer(tt.v1, tt.v2)
			if (err != nil) != tt.wantError {
				t.Errorf("IsNewer() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("IsNewer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"1.0.0", "2.1.0", "1.5.0", "2.0.0"}
	expected := []string{"2.1.0", "2.0.0", "1.5.0", "1.0.0"} // latest first

	result, err := SortVersions(versions)
	if err != nil {
		t.Errorf("SortVersions() error = %v", err)
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("SortVersions()[%d] = %s, want %s", i, v, expected[i])
		}
	}
}

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name      string
		versions  []string
		expected  string
		wantError bool
	}{
		{"normal case", []string{"1.0.0", "2.1.0", "1.5.0"}, "2.1.0", false},
		{"single version", []string{"1.0.0"}, "1.0.0", false},
		{"empty list", []string{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetLatestVersion(tt.versions)
			if (err != nil) != tt.wantError {
				t.Errorf("GetLatestVersion() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("GetLatestVersion() = %s, want %s", result, tt.expected)
			}
		})
	}
}
