package filter

import (
	"testing"
)

func TestNewMatcher_ExactMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		wantErr  bool
		wantMode MatchMode
	}{
		{
			name:     "simple exact pattern",
			pattern:  "vault-starter",
			wantErr:  false,
			wantMode: MatchModeExact,
		},
		{
			name:     "full module path",
			pattern:  "hashicorp/vault-starter/aws",
			wantErr:  false,
			wantMode: MatchModeExact,
		},
		{
			name:     "pattern with dots - is regex",
			pattern:  "registry.example.com/org/module/provider",
			wantErr:  false,
			wantMode: MatchModeRegex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && matcher.Mode != tt.wantMode {
				t.Errorf("expected Mode %v, got %v", tt.wantMode, matcher.Mode)
			}
		})
	}
}

func TestNewMatcher_RegexMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		wantErr  bool
		wantMode MatchMode
	}{
		{
			name:     "simple regex with dot",
			pattern:  ".*vpc.*",
			wantErr:  false,
			wantMode: MatchModeRegex,
		},
		{
			name:     "regex with quantifier",
			pattern:  "vault.*",
			wantErr:  false,
			wantMode: MatchModeRegex,
		},
		{
			name:     "regex with character class",
			pattern:  "[a-z]+/[a-z]+/[a-z]+",
			wantErr:  false,
			wantMode: MatchModeRegex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && matcher.Mode != tt.wantMode {
				t.Errorf("expected Mode %v, got %v", tt.wantMode, matcher.Mode)
			}
		})
	}
}

func TestMatcher_Matches_Exact(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		source  string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "vault-starter",
			source:  "vault-starter",
			want:    true,
		},
		{
			name:    "full path exact match",
			pattern: "hashicorp/vault-starter/aws",
			source:  "hashicorp/vault-starter/aws",
			want:    true,
		},
		{
			name:    "no match - different string",
			pattern: "vault-starter",
			source:  "vault",
			want:    false,
		},
		{
			name:    "no match - partial",
			pattern: "vault-starter",
			source:  "hashicorp/vault-starter/aws",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, _ := NewMatcher(tt.pattern)
			if got := matcher.Matches(tt.source); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_Matches_Regex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		source  string
		want    bool
	}{
		{
			name:    "regex matches wildcard",
			pattern: ".*vpc.*",
			source:  "terraform-aws-vpc",
			want:    true,
		},
		{
			name:    "regex matches prefix",
			pattern: "vault.*",
			source:  "vault-starter",
			want:    true,
		},
		{
			name:    "regex matches suffix",
			pattern: ".*vpc",
			source:  "terraform-aws-vpc",
			want:    true,
		},
		{
			name:    "regex no match",
			pattern: "vault.*",
			source:  "hashicorp/aws/ec2",
			want:    false,
		},
		{
			name:    "regex with full path",
			pattern: "hashicorp/.*/aws",
			source:  "hashicorp/vault-starter/aws",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, _ := NewMatcher(tt.pattern)
			if got := matcher.Matches(tt.source); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchModule(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		pattern string
		want    bool
		wantErr bool
	}{
		{
			name:    "exact match",
			source:  "vault-starter",
			pattern: "vault-starter",
			want:    true,
			wantErr: false,
		},
		{
			name:    "regex match",
			source:  "terraform-aws-vpc",
			pattern: ".*vpc.*",
			want:    true,
			wantErr: false,
		},
		{
			name:    "no match",
			source:  "vault",
			pattern: ".*vpc.*",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchModule(tt.source, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchModule() = %v, want %v", got, tt.want)
			}
		})
	}
}
