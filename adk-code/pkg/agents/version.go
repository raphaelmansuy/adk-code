package agents

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version (Major.Minor.Patch-Prerelease).
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

// ConstraintType represents the type of version constraint.
type ConstraintType string

const (
	ConstraintExact      ConstraintType = "=="
	ConstraintGreater    ConstraintType = ">"
	ConstraintGreaterEq  ConstraintType = ">="
	ConstraintLess       ConstraintType = "<"
	ConstraintLessEq     ConstraintType = "<="
	ConstraintCaretRange ConstraintType = "^"
	ConstraintTildeRange ConstraintType = "~"
	ConstraintRange      ConstraintType = "-"
)

// Constraint represents a version constraint (e.g., ^1.0.0, ~1.0.0, >=1.0.0).
type Constraint struct {
	Type       ConstraintType
	Version    *Version
	UpperBound *Version // For range constraints
}

// ParseVersion parses a semantic version string.
// Supported formats:
//   - 1.0.0
//   - 1.0.0-alpha
//   - 1.0.0-alpha.1
func ParseVersion(s string) (*Version, error) {
	if s == "" {
		return nil, fmt.Errorf("version string is empty")
	}

	// Split prerelease
	var prerelease string
	parts := strings.Split(s, "-")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid version format: %s", s)
	}

	if len(parts) == 2 {
		prerelease = parts[1]
	}

	// Parse major.minor.patch
	versionParts := strings.Split(parts[0], ".")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("version must have major.minor.patch: %s", s)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", versionParts[0])
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", versionParts[2])
	}

	if major < 0 || minor < 0 || patch < 0 {
		return nil, fmt.Errorf("version numbers must be non-negative")
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
	}, nil
}

// ParseConstraint parses a version constraint string.
// Supported formats:
//   - 1.0.0              (exact)
//   - ^1.0.0             (compatible, >=1.0.0 <2.0.0)
//   - ~1.0.0             (patch, >=1.0.0 <1.1.0)
//   - >=1.0.0            (greater than or equal)
//   - >1.0.0             (greater than)
//   - <=1.0.0            (less than or equal)
//   - <1.0.0             (less than)
//   - 1.0.0-2.0.0        (range)
func ParseConstraint(s string) (*Constraint, error) {
	if s == "" {
		return nil, fmt.Errorf("constraint string is empty")
	}

	s = strings.TrimSpace(s)

	// Check for range constraint (X.Y.Z - A.B.C)
	if strings.Contains(s, "-") && !strings.HasPrefix(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 {
			lowerStr := strings.TrimSpace(parts[0])
			upperStr := strings.TrimSpace(parts[1])

			lower, err := ParseVersion(lowerStr)
			if err != nil {
				return nil, err
			}

			upper, err := ParseVersion(upperStr)
			if err != nil {
				return nil, err
			}

			return &Constraint{
				Type:       ConstraintRange,
				Version:    lower,
				UpperBound: upper,
			}, nil
		}
	}

	// Check for operator constraints
	operators := []string{">=", "<=", "==", "^", "~", ">", "<"}
	for _, op := range operators {
		if strings.HasPrefix(s, op) {
			versionStr := strings.TrimPrefix(s, op)
			version, err := ParseVersion(versionStr)
			if err != nil {
				return nil, err
			}

			return &Constraint{
				Type:    ConstraintType(op),
				Version: version,
			}, nil
		}
	}

	// No operator - treat as exact version
	version, err := ParseVersion(s)
	if err != nil {
		return nil, err
	}

	return &Constraint{
		Type:    ConstraintExact,
		Version: version,
	}, nil
}

// Matches checks if a version satisfies this constraint.
func (c *Constraint) Matches(v *Version) bool {
	if v == nil || c.Version == nil {
		return false
	}

	switch c.Type {
	case ConstraintExact:
		return c.versionEqual(v, c.Version)

	case ConstraintGreater:
		return c.versionGreater(v, c.Version)

	case ConstraintGreaterEq:
		return c.versionGreater(v, c.Version) || c.versionEqual(v, c.Version)

	case ConstraintLess:
		return c.versionLess(v, c.Version)

	case ConstraintLessEq:
		return c.versionLess(v, c.Version) || c.versionEqual(v, c.Version)

	case ConstraintCaretRange:
		// ^X.Y.Z := >=X.Y.Z <(X+1).0.0
		if !c.versionGreaterEq(v, c.Version) {
			return false
		}

		upper := &Version{
			Major: c.Version.Major + 1,
			Minor: 0,
			Patch: 0,
		}

		return c.versionLess(v, upper)

	case ConstraintTildeRange:
		// ~X.Y.Z := >=X.Y.Z <X.(Y+1).0
		if !c.versionGreaterEq(v, c.Version) {
			return false
		}

		upper := &Version{
			Major: c.Version.Major,
			Minor: c.Version.Minor + 1,
			Patch: 0,
		}

		return c.versionLess(v, upper)

	case ConstraintRange:
		// X.Y.Z - A.B.C := >=X.Y.Z <=A.B.C
		if !c.versionGreaterEq(v, c.Version) {
			return false
		}

		if c.UpperBound == nil {
			return false
		}

		return c.versionLessEq(v, c.UpperBound)

	default:
		return false
	}
}

// String returns the string representation of a version.
func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	return s
}

// String returns the string representation of a constraint.
func (c *Constraint) String() string {
	if c.Type == ConstraintRange {
		return fmt.Sprintf("%s-%s", c.Version.String(), c.UpperBound.String())
	}
	return fmt.Sprintf("%s%s", c.Type, c.Version.String())
}

// Helper comparison functions
func (c *Constraint) versionEqual(v1, v2 *Version) bool {
	return v1.Major == v2.Major &&
		v1.Minor == v2.Minor &&
		v1.Patch == v2.Patch &&
		v1.Prerelease == v2.Prerelease
}

func (c *Constraint) versionGreater(v1, v2 *Version) bool {
	if v1.Major != v2.Major {
		return v1.Major > v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor > v2.Minor
	}
	if v1.Patch != v2.Patch {
		return v1.Patch > v2.Patch
	}

	// Prerelease versions have lower precedence
	if v1.Prerelease != "" && v2.Prerelease == "" {
		return false
	}
	if v1.Prerelease == "" && v2.Prerelease != "" {
		return true
	}

	return v1.Prerelease > v2.Prerelease
}

func (c *Constraint) versionGreaterEq(v1, v2 *Version) bool {
	return c.versionGreater(v1, v2) || c.versionEqual(v1, v2)
}

func (c *Constraint) versionLess(v1, v2 *Version) bool {
	return !c.versionGreaterEq(v1, v2)
}

func (c *Constraint) versionLessEq(v1, v2 *Version) bool {
	return c.versionLess(v1, v2) || c.versionEqual(v1, v2)
}
