package version

import (
	"testing"
)

func TestSelectVersion_Latest(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		availableVersions []string
		want              string
		wantErr           bool
	}{
		{
			name:              "select latest from multiple versions",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "1.5.0", "1.0.0"},
			want:              "2.0.0",
			wantErr:           false,
		},
		{
			name:              "select latest with single version",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0"},
			want:              "1.0.0",
			wantErr:           false,
		},
		{
			name:              "empty available versions",
			currentVersion:    "1.0.0",
			availableVersions: []string{},
			want:              "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectVersion(tt.currentVersion, tt.availableVersions, StrategyLatest, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SelectVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectVersion_Minor(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		availableVersions []string
		want              string
		wantErr           bool
	}{
		{
			name:              "select latest minor matching major",
			currentVersion:    "1.2.3",
			availableVersions: []string{"2.0.0", "1.5.2", "1.5.1", "1.0.0"},
			want:              "1.5.2",
			wantErr:           false,
		},
		{
			name:              "no higher minor, same version available",
			currentVersion:    "1.5.0",
			availableVersions: []string{"2.0.0", "1.5.0", "0.9.0"},
			want:              "1.5.0",
			wantErr:           false,
		},
		{
			name:              "multiple patch versions same minor",
			currentVersion:    "1.2.0",
			availableVersions: []string{"1.2.8", "1.2.5", "1.2.3"},
			want:              "1.2.8",
			wantErr:           false,
		},
		{
			name:              "no matching major version",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "2.1.0", "3.0.0"},
			want:              "",
			wantErr:           true,
		},
		{
			name:              "invalid current version",
			currentVersion:    "invalid",
			availableVersions: []string{"1.0.0"},
			want:              "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectVersion(tt.currentVersion, tt.availableVersions, StrategyMinor, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SelectVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectVersion_InvalidStrategy(t *testing.T) {
	got, err := SelectVersion("1.0.0", []string{"2.0.0", "1.0.0"}, Strategy("invalid"), nil)
	if err == nil {
		t.Errorf("SelectVersion() expected error for invalid strategy, got nil")
	}
	if got != "" {
		t.Errorf("SelectVersion() expected empty string, got %v", got)
	}
}

func TestSelectVersion_WithConstraints_Latest(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		availableVersions []string
		constraints       string
		want              string
		wantErr           bool
	}{
		{
			name:              "select highest matching constraint",
			currentVersion:    "2.0.0",
			availableVersions: []string{"3.0.0", "2.5.0", "1.9.0"},
			constraints:       ">= 2.0, < 3.0",
			want:              "2.5.0",
			wantErr:           false,
		},
		{
			name:              "single constraint lower bound",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "1.5.0", "1.0.0", "0.9.0"},
			constraints:       ">= 1.0",
			want:              "2.0.0",
			wantErr:           false,
		},
		{
			name:              "single constraint upper bound",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "1.5.0", "1.0.0", "0.9.0"},
			constraints:       "< 2.0",
			want:              "1.5.0",
			wantErr:           false,
		},
		{
			name:              "pessimistic constraint",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.5.0", "2.0.0", "1.9.0", "1.5.0", "1.0.0"},
			constraints:       "~> 1",
			want:              "1.9.0",
			wantErr:           false,
		},
		{
			name:              "no versions match constraint",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "3.0.0"},
			constraints:       "< 2.0",
			want:              "",
			wantErr:           true,
		},
		{
			name:              "exact match constraint",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.0.0", "1.5.0", "1.0.0"},
			constraints:       "= 1.5.0",
			want:              "1.5.0",
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := ParseConstraints(tt.constraints)
			if err != nil {
				t.Fatalf("failed to parse constraints: %v", err)
			}

			got, err := SelectVersion(tt.currentVersion, tt.availableVersions, StrategyLatest, constraints)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SelectVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectVersion_WithConstraints_Minor(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		availableVersions []string
		constraints       string
		want              string
		wantErr           bool
	}{
		{
			name:              "minor strategy respects constraints",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.5.0", "2.0.0", "1.9.0", "1.5.0", "1.0.0"},
			constraints:       "~> 1",
			want:              "1.9.0",
			wantErr:           false,
		},
		{
			name:              "minor strategy with range constraint",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.9.0", "1.5.0", "1.0.0"},
			constraints:       ">= 1.2, < 1.8",
			want:              "1.5.0",
			wantErr:           false,
		},
		{
			name:              "minor strategy no matching major",
			currentVersion:    "1.0.0",
			availableVersions: []string{"2.5.0", "2.0.0"},
			constraints:       ">= 1.0",
			want:              "",
			wantErr:           true,
		},
		{
			name:              "minor strategy exact match in same major",
			currentVersion:    "1.2.0",
			availableVersions: []string{"1.3.0", "1.2.5", "1.2.0"},
			constraints:       ">= 1.2.0",
			want:              "1.3.0",
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := ParseConstraints(tt.constraints)
			if err != nil {
				t.Fatalf("failed to parse constraints: %v", err)
			}

			got, err := SelectVersion(tt.currentVersion, tt.availableVersions, StrategyMinor, constraints)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SelectVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterVersionsByConstraints(t *testing.T) {
	tests := []struct {
		name        string
		versions    []string
		constraints string
		want        []string
	}{
		{
			name:        "filter range constraint",
			versions:    []string{"2.5.0", "2.0.0", "1.9.0", "1.5.0", "1.0.0"},
			constraints: ">= 1.5, < 2.0",
			want:        []string{"1.9.0", "1.5.0"},
		},
		{
			name:        "filter pessimistic constraint major",
			versions:    []string{"2.5.0", "2.0.0", "1.9.0", "1.5.0", "1.0.0"},
			constraints: "~> 1",
			want:        []string{"1.9.0", "1.5.0", "1.0.0"},
		},
		{
			name:        "filter exact match",
			versions:    []string{"2.0.0", "1.5.0", "1.0.0"},
			constraints: "= 1.5.0",
			want:        []string{"1.5.0"},
		},
		{
			name:        "filter not equal",
			versions:    []string{"2.0.0", "1.5.0", "1.0.0"},
			constraints: "!= 1.5.0",
			want:        []string{"2.0.0", "1.0.0"},
		},
		{
			name:        "no versions match",
			versions:    []string{"2.0.0", "3.0.0"},
			constraints: "< 2.0",
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := ParseConstraints(tt.constraints)
			if err != nil {
				t.Fatalf("failed to parse constraints: %v", err)
			}

			got := filterVersionsByConstraints(tt.versions, constraints)
			if len(got) != len(tt.want) {
				t.Errorf("filterVersionsByConstraints() returned %d versions, want %d", len(got), len(tt.want))
				return
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("filterVersionsByConstraints() version %d = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestIsValidStrategy(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid minor",
			s:    "minor",
			want: true,
		},
		{
			name: "valid latest",
			s:    "latest",
			want: true,
		},
		{
			name: "invalid strategy",
			s:    "invalid",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "case sensitive",
			s:    "Minor",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidStrategy(tt.s)
			if got != tt.want {
				t.Errorf("IsValidStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}
