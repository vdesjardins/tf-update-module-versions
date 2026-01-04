package version

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    *Constraint
		wantErr bool
	}{
		{
			name: "exact match",
			expr: "= 1.0.0",
			want: &Constraint{
				Operator: "=",
				Version:  semver.MustParse("1.0.0"),
			},
		},
		{
			name: "greater than",
			expr: "> 1.2.3",
			want: &Constraint{
				Operator: ">",
				Version:  semver.MustParse("1.2.3"),
			},
		},
		{
			name: "greater than or equal",
			expr: ">= 2.0.0",
			want: &Constraint{
				Operator: ">=",
				Version:  semver.MustParse("2.0.0"),
			},
		},
		{
			name: "less than",
			expr: "< 1.5.0",
			want: &Constraint{
				Operator: "<",
				Version:  semver.MustParse("1.5.0"),
			},
		},
		{
			name: "less than or equal",
			expr: "<= 3.0.0",
			want: &Constraint{
				Operator: "<=",
				Version:  semver.MustParse("3.0.0"),
			},
		},
		{
			name: "not equal",
			expr: "!= 1.1.0",
			want: &Constraint{
				Operator: "!=",
				Version:  semver.MustParse("1.1.0"),
			},
		},
		{
			name: "pessimistic constraint",
			expr: "~> 1.2.3",
			want: &Constraint{
				Operator: "~>",
				Version:  semver.MustParse("1.2.3"),
			},
		},
		{
			name:    "invalid operator",
			expr:    ">> 1.0.0",
			wantErr: true,
		},
		{
			name:    "invalid version",
			expr:    ">= invalid",
			wantErr: true,
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name: "whitespace handling",
			expr: "  >=   1.0.0  ",
			want: &Constraint{
				Operator: ">=",
				Version:  semver.MustParse("1.0.0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConstraint(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Operator != tt.want.Operator {
					t.Errorf("ParseConstraint() operator = %v, want %v", got.Operator, tt.want.Operator)
				}
				if !got.Version.Equal(tt.want.Version) {
					t.Errorf("ParseConstraint() version = %v, want %v", got.Version, tt.want.Version)
				}
			}
		})
	}
}

func TestConstraintMatches(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		// Exact match tests
		{"exact match equal", "= 1.0.0", "1.0.0", true},
		{"exact match not equal", "= 1.0.0", "1.0.1", false},

		// Greater than tests
		{"gt true", "> 1.0.0", "1.0.1", true},
		{"gt false", "> 1.0.0", "1.0.0", false},
		{"gt false less", "> 1.0.0", "0.9.0", false},

		// Greater than or equal tests
		{"gte equal", ">= 1.0.0", "1.0.0", true},
		{"gte greater", ">= 1.0.0", "1.0.1", true},
		{"gte less", ">= 1.0.0", "0.9.0", false},

		// Less than tests
		{"lt true", "< 2.0.0", "1.9.0", true},
		{"lt false equal", "< 2.0.0", "2.0.0", false},
		{"lt false greater", "< 2.0.0", "2.0.1", false},

		// Less than or equal tests
		{"lte equal", "<= 2.0.0", "2.0.0", true},
		{"lte less", "<= 2.0.0", "1.9.0", true},
		{"lte greater", "<= 2.0.0", "2.0.1", false},

		// Not equal tests
		{"ne true", "!= 1.0.0", "1.0.1", true},
		{"ne false", "!= 1.0.0", "1.0.0", false},

		// Pessimistic constraint tests
		{"pessimistic exact patch", "~> 1.2.3", "1.2.3", true},
		{"pessimistic patch up", "~> 1.2.3", "1.2.4", true},
		{"pessimistic minor up", "~> 1.2.3", "1.3.0", false},
		{"pessimistic patch down", "~> 1.2.3", "1.2.2", false},

		{"pessimistic minor exact", "~> 1.2", "1.2.0", true},
		{"pessimistic minor up", "~> 1.2", "1.3.0", false},
		{"pessimistic minor major up", "~> 1.2", "2.0.0", false},

		{"pessimistic major exact", "~> 1", "1.0.0", true},
		{"pessimistic major up", "~> 1", "1.5.0", true},
		{"pessimistic major next", "~> 1", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() error = %v", err)
			}

			version := semver.MustParse(tt.version)
			got := constraint.Matches(version)

			if got != tt.want {
				t.Errorf("Constraint.Matches(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestParseConstraints(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		wantLen    int
		wantErr    bool
		wantString string
	}{
		{
			name:       "single constraint",
			expr:       ">= 1.0.0",
			wantLen:    1,
			wantString: ">= 1.0.0",
		},
		{
			name:       "multiple constraints",
			expr:       ">= 1.0.0, < 2.0.0",
			wantLen:    2,
			wantString: ">= 1.0.0, < 2.0.0",
		},
		{
			name:       "complex constraints",
			expr:       "~> 2.0, != 2.0.1",
			wantLen:    2,
			wantString: "~> 2.0.0, != 2.0.1",
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "invalid constraint",
			expr:    ">= 1.0.0, >> invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := ParseConstraints(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(constraints) != tt.wantLen {
					t.Errorf("ParseConstraints() len = %d, want %d", len(constraints), tt.wantLen)
				}
				if constraints.String() != tt.wantString {
					t.Errorf("ParseConstraints() string = %q, want %q", constraints.String(), tt.wantString)
				}
			}
		})
	}
}

func TestConstraintsMatches(t *testing.T) {
	tests := []struct {
		name        string
		constraints string
		version     string
		want        bool
	}{
		{
			name:        "single constraint match",
			constraints: ">= 1.0.0",
			version:     "1.0.0",
			want:        true,
		},
		{
			name:        "multiple constraints all match",
			constraints: ">= 1.0.0, < 2.0.0",
			version:     "1.5.0",
			want:        true,
		},
		{
			name:        "multiple constraints one fails",
			constraints: ">= 1.0.0, < 2.0.0",
			version:     "2.0.0",
			want:        false,
		},
		{
			name:        "multiple constraints one fails other way",
			constraints: ">= 1.0.0, < 2.0.0",
			version:     "0.9.0",
			want:        false,
		},
		{
			name:        "range constraint",
			constraints: ">= 1.2.0, < 1.3.0, != 1.2.5",
			version:     "1.2.3",
			want:        true,
		},
		{
			name:        "range constraint excluded",
			constraints: ">= 1.2.0, < 1.3.0, != 1.2.5",
			version:     "1.2.5",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := ParseConstraints(tt.constraints)
			if err != nil {
				t.Fatalf("ParseConstraints() error = %v", err)
			}

			got := constraints.MatchesString(tt.version)

			if got != tt.want {
				t.Errorf("Constraints.MatchesString(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestConstraintString(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
		want       string
	}{
		{
			name: "nil constraint",
			want: "<nil>",
		},
		{
			name: "normal constraint",
			constraint: &Constraint{
				Operator: ">=",
				Version:  semver.MustParse("1.0.0"),
			},
			want: ">= 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constraint.String()
			if got != tt.want {
				t.Errorf("Constraint.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestConstraintMatchesNilVersion tests Matches with nil version input
func TestConstraintMatchesNilVersion(t *testing.T) {
	constraint, _ := ParseConstraint(">= 1.0.0")
	got := constraint.Matches(nil)
	if got != false {
		t.Errorf("Constraint.Matches(nil) = %v, want false", got)
	}
}

// TestConstraintMatchesInvalidOperator tests Matches with unknown operator
func TestConstraintMatchesInvalidOperator(t *testing.T) {
	constraint := &Constraint{
		Operator: "@@", // Invalid operator
		Version:  semver.MustParse("1.0.0"),
	}
	version := semver.MustParse("1.0.0")
	got := constraint.Matches(version)
	if got != false {
		t.Errorf("Constraint.Matches() with invalid operator = %v, want false", got)
	}
}

// TestConstraintsMatchesNilConstraints tests Matches with empty constraints list
func TestConstraintsMatchesNilConstraints(t *testing.T) {
	var constraints Constraints
	version := semver.MustParse("1.0.0")
	got := constraints.Matches(version)
	if got != true {
		t.Errorf("Constraints.Matches() with nil constraints = %v, want true", got)
	}
}

// TestConstraintsMatchesStringInvalidVersion tests MatchesString with invalid version
func TestConstraintsMatchesStringInvalidVersion(t *testing.T) {
	constraints, _ := ParseConstraints(">= 1.0.0")
	got := constraints.MatchesString("invalid-version")
	if got != false {
		t.Errorf("Constraints.MatchesString() with invalid version = %v, want false", got)
	}
}

// TestConstraintsMatchesStringEmptyString tests MatchesString with empty version string
func TestConstraintsMatchesStringEmptyString(t *testing.T) {
	constraints, _ := ParseConstraints(">= 1.0.0")
	got := constraints.MatchesString("")
	if got != false {
		t.Errorf("Constraints.MatchesString() with empty string = %v, want false", got)
	}
}

// TestConstraintsString tests the Constraints.String() method
func TestConstraintsString(t *testing.T) {
	tests := []struct {
		name        string
		constraints Constraints
		want        string
	}{
		{
			name:        "nil constraints",
			constraints: nil,
			want:        "",
		},
		{
			name:        "empty constraints",
			constraints: Constraints{},
			want:        "",
		},
		{
			name: "single constraint",
			constraints: Constraints{
				&Constraint{
					Operator: ">=",
					Version:  semver.MustParse("1.0.0"),
				},
			},
			want: ">= 1.0.0",
		},
		{
			name: "multiple constraints",
			constraints: Constraints{
				&Constraint{
					Operator: ">=",
					Version:  semver.MustParse("1.0.0"),
				},
				&Constraint{
					Operator: "<",
					Version:  semver.MustParse("2.0.0"),
				},
			},
			want: ">= 1.0.0, < 2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constraints.String()
			if got != tt.want {
				t.Errorf("Constraints.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPessimisticConstraintEdgeCases tests pessimistic constraint with various precision levels
func TestPessimisticConstraintEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		// Pessimistic with pre-release versions
		{"pessimistic with prerelease lower", "~> 1.2.0", "1.2.0-rc1", false},
		{"pessimistic with prerelease higher", "~> 1.2.0-rc1", "1.2.0", true},
		{"pessimistic with prerelease in range", "~> 1.2.0", "1.2.5", true},

		// Pessimistic with major version boundaries
		{"pessimistic major boundary low", "~> 1.0.0", "0.9.9", false},
		{"pessimistic major boundary high", "~> 1.0.0", "1.1.0", false},
		{"pessimistic major boundary exact", "~> 1.0.0", "1.0.5", true},

		// Pessimistic with various patch levels
		{"pessimistic patch boundary", "~> 1.2.5", "1.3.0", false},
		{"pessimistic patch within range", "~> 1.2.5", "1.2.99", true},

		// Pessimistic major-only versions
		{"pessimistic major only start", "~> 2", "2.0.0", true},
		{"pessimistic major only high", "~> 2", "2.99.99", true},
		{"pessimistic major only next", "~> 2", "3.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() error = %v", err)
			}

			version := semver.MustParse(tt.version)
			got := constraint.Matches(version)

			if got != tt.want {
				t.Errorf("Constraint.Matches(%q) with %q = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

// TestConstraintWithPrerelease tests constraint handling with pre-release versions
func TestConstraintWithPrerelease(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		// Pre-release versions
		{"prerelease exact match", "= 1.0.0-alpha", "1.0.0-alpha", true},
		{"prerelease not equal", "!= 1.0.0-alpha", "1.0.0-beta", true},
		{"prerelease greater than", "> 1.0.0-alpha", "1.0.0", true},
		{"prerelease greater than false", "> 1.0.0", "1.0.0-alpha", false},
		{"prerelease less than", "< 1.0.0-beta", "1.0.0-alpha", true},
		{"prerelease less than equal", "<= 1.0.0-beta", "1.0.0-beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() error = %v", err)
			}

			version := semver.MustParse(tt.version)
			got := constraint.Matches(version)

			if got != tt.want {
				t.Errorf("Constraint.Matches(%q) with %q = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

// TestConstraintParseEdgeCases tests parsing edge cases and error conditions
func TestConstraintParseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		// Valid edge cases
		{"no spaces", ">=1.0.0", false},
		{"multiple spaces", ">    1.0.0", false},
		{"tabs", ">\t1.0.0", false},

		// Invalid cases
		{"missing operator", "1.0.0", true},
		{"malformed operator", "==> 1.0.0", true},
		{"version with letters", ">= 1.a.0", true},
		{"negative version", ">= -1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConstraint(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}

// TestConstraintsMatchesAllFail tests scenario where all constraints fail
func TestConstraintsMatchesAllFail(t *testing.T) {
	constraints, _ := ParseConstraints(">= 2.0.0, < 3.0.0")
	version := semver.MustParse("1.0.0")
	got := constraints.Matches(version)
	if got != false {
		t.Errorf("Constraints.Matches() when all fail = %v, want false", got)
	}
}

// TestConstraintStringWithNilVersion tests Constraint.String with nil version
func TestConstraintStringWithNilVersion(t *testing.T) {
	constraint := &Constraint{
		Operator: ">=",
		Version:  nil,
	}
	got := constraint.String()
	if got != "<nil>" {
		t.Errorf("Constraint.String() with nil version = %q, want <nil>", got)
	}
}
