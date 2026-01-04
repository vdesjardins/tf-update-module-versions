package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Constraint represents a single version constraint with an operator and version.
// The Original field stores the original expression to help determine precision.
type Constraint struct {
	Operator string          // =, !=, >, >=, <, <=, ~>
	Version  *semver.Version // Parsed semantic version
	Original string          // Original constraint expression for precision detection
}

// Constraints represents multiple constraints that must all be satisfied (AND semantics).
type Constraints []*Constraint

// ParseConstraint parses a single constraint expression and returns a Constraint.
//
// Supported operators:
//   - = (exact match)
//   - != (not equal)
//   - > (greater than)
//   - >= (greater than or equal)
//   - < (less than)
//   - <= (less than or equal)
//   - ~> (pessimistic constraint)
//
// Examples:
//   - ">= 1.2.0"
//   - "~> 2.0" (equivalent to ">=2.0.0, <3.0.0")
//   - "!= 1.0.0"
func ParseConstraint(expr string) (*Constraint, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty constraint expression")
	}

	// Pattern to match operator and version
	// Matches operators: =, !=, >, >=, <, <=, ~>
	pattern := regexp.MustCompile(`^(=|!=|>=|>|<=|<|~>)\s*(.+)$`)
	matches := pattern.FindStringSubmatch(expr)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid constraint expression: %q", expr)
	}

	operator := matches[1]
	versionStr := strings.TrimSpace(matches[2])

	// Parse the version using semver
	// For operators like ~>, we may need to handle it specially
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q in constraint: %w", versionStr, err)
	}

	return &Constraint{
		Operator: operator,
		Version:  version,
		Original: expr,
	}, nil
}

// ParseConstraints parses a comma-separated list of constraint expressions.
// All constraints must be satisfied (AND semantics).
//
// Example:
//   - ">= 1.0, < 2.0"
//   - "~> 2.0, != 2.0.1"
func ParseConstraints(expr string) (Constraints, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty constraints expression")
	}

	var constraints Constraints

	// Split by comma for multiple constraints
	parts := strings.Split(expr, ",")
	for _, part := range parts {
		constraint, err := ParseConstraint(part)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

// Matches checks if a version satisfies a single constraint.
//
// Examples:
//   - ">= 1.2.0" matches "1.2.0", "1.3.0", "2.0.0" but not "1.1.0"
//   - "~> 2.0" matches "2.0.0", "2.1.0" but not "3.0.0" or "1.9.0"
//   - "= 1.0.0" matches only "1.0.0"
func (c *Constraint) Matches(version *semver.Version) bool {
	if version == nil {
		return false
	}

	cmp := version.Compare(c.Version)

	switch c.Operator {
	case "=":
		return cmp == 0
	case "!=":
		return cmp != 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case "~>":
		return c.matchesPessimistic(version)
	default:
		return false
	}
}

// matchesPessimistic checks if a version matches a pessimistic constraint.
// Pessimistic constraint ~> allows changes at the precision level below the most specific component.
//
// Examples:
//   - ~> 1.2.3 means >= 1.2.3 and < 1.3.0 (allows patch changes)
//   - ~> 1.2 means >= 1.2.0 and < 1.3.0 (allows minor changes within the major version)
//   - ~> 1 means >= 1.0.0 and < 2.0.0 (allows major changes)
func (c *Constraint) matchesPessimistic(version *semver.Version) bool {
	// Must be >= the constraint version
	if version.Compare(c.Version) < 0 {
		return false
	}

	major := c.Version.Major()
	minor := c.Version.Minor()

	var upperBound *semver.Version

	// Parse the original constraint to determine precision
	// Extract just the version part (after the operator)
	parts := strings.Fields(c.Original)
	if len(parts) < 2 {
		// Fallback: assume patch-level precision
		upperBound, _ = semver.NewVersion(fmt.Sprintf("%d.%d.0", major, minor+1))
	} else {
		versionPart := parts[len(parts)-1]
		versionDots := strings.Count(versionPart, ".")

		switch versionDots {
		case 0:
			// Only major version specified (e.g., ~> 1)
			// Allow changes up to next major: >= 1.0.0, < 2.0.0
			upperBound, _ = semver.NewVersion(fmt.Sprintf("%d.0.0", major+1))
		case 1:
			// Major.minor specified (e.g., ~> 1.2)
			// Allow changes up to next minor: >= 1.2.0, < 1.3.0
			upperBound, _ = semver.NewVersion(fmt.Sprintf("%d.%d.0", major, minor+1))
		default:
			// Major.minor.patch or more specified (e.g., ~> 1.2.3)
			// Allow changes up to next minor: >= 1.2.3, < 1.3.0
			upperBound, _ = semver.NewVersion(fmt.Sprintf("%d.%d.0", major, minor+1))
		}
	}

	return version.Compare(upperBound) < 0
}

// Matches checks if a version satisfies all constraints in the list (AND semantics).
func (c Constraints) Matches(version *semver.Version) bool {
	if len(c) == 0 {
		return true
	}

	for _, constraint := range c {
		if !constraint.Matches(version) {
			return false
		}
	}
	return true
}

// MatchesString checks if a version string satisfies all constraints.
func (c Constraints) MatchesString(versionStr string) bool {
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return false
	}
	return c.Matches(version)
}

// String returns a human-readable representation of the constraint.
func (c *Constraint) String() string {
	if c == nil || c.Version == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s %s", c.Operator, c.Version)
}

// String returns a human-readable representation of all constraints.
func (c Constraints) String() string {
	if len(c) == 0 {
		return ""
	}

	parts := make([]string, len(c))
	for i, constraint := range c {
		parts[i] = constraint.String()
	}
	return strings.Join(parts, ", ")
}
