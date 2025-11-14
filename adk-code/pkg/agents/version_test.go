package agents

import (
	"testing"
)

// TestParseVersion tests version parsing
func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    *Version
		wantErr bool
	}{
		{
			input: "1.0.0",
			want: &Version{
				Major: 1, Minor: 0, Patch: 0, Prerelease: "",
			},
			wantErr: false,
		},
		{
			input: "2.3.4",
			want: &Version{
				Major: 2, Minor: 3, Patch: 4, Prerelease: "",
			},
			wantErr: false,
		},
		{
			input: "1.0.0-alpha",
			want: &Version{
				Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha",
			},
			wantErr: false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "1.0",
			wantErr: true,
		},
		{
			input:   "1.0.0-alpha-beta",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		v, err := ParseVersion(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("ParseVersion(%q): expected error, got nil", tt.input)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseVersion(%q): got error %v", tt.input, err)
		}
		if !tt.wantErr && tt.want != nil {
			if v.Major != tt.want.Major || v.Minor != tt.want.Minor ||
				v.Patch != tt.want.Patch || v.Prerelease != tt.want.Prerelease {
				t.Errorf("ParseVersion(%q): got %v, want %v", tt.input, v, tt.want)
			}
		}
	}
}

// TestVersionString tests version string representation
func TestVersionString(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: ""}
	if v.String() != "1.2.3" {
		t.Errorf("Expected '1.2.3', got %q", v.String())
	}

	v.Prerelease = "alpha"
	if v.String() != "1.2.3-alpha" {
		t.Errorf("Expected '1.2.3-alpha', got %q", v.String())
	}
}

// TestParseConstraint tests constraint parsing
func TestParseConstraint(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1.0.0", false},
		{"^1.0.0", false},
		{"~1.0.0", false},
		{">=1.0.0", false},
		{">1.0.0", false},
		{"<=1.0.0", false},
		{"<1.0.0", false},
		{"==1.0.0", false},
		{"1.0.0-2.0.0", false},
		{"", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		_, err := ParseConstraint(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("ParseConstraint(%q): expected error, got nil", tt.input)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseConstraint(%q): got error %v", tt.input, err)
		}
	}
}

// TestConstraintExact tests exact version constraint
func TestConstraintExact(t *testing.T) {
	c, _ := ParseConstraint("1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.1")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint 1.0.0")
	}

	if c.Matches(v2) {
		t.Error("Expected 1.0.1 to not match constraint 1.0.0")
	}
}

// TestConstraintGreater tests greater than constraint
func TestConstraintGreater(t *testing.T) {
	c, _ := ParseConstraint(">1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.1")

	if c.Matches(v1) {
		t.Error("Expected 1.0.0 to not match constraint >1.0.0")
	}

	if !c.Matches(v2) {
		t.Error("Expected 1.0.1 to match constraint >1.0.0")
	}
}

// TestConstraintGreaterEq tests greater than or equal constraint
func TestConstraintGreaterEq(t *testing.T) {
	c, _ := ParseConstraint(">=1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.1")
	v3, _ := ParseVersion("0.9.0")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint >=1.0.0")
	}

	if !c.Matches(v2) {
		t.Error("Expected 1.0.1 to match constraint >=1.0.0")
	}

	if c.Matches(v3) {
		t.Error("Expected 0.9.0 to not match constraint >=1.0.0")
	}
}

// TestConstraintLess tests less than constraint
func TestConstraintLess(t *testing.T) {
	c, _ := ParseConstraint("<1.0.0")

	v1, _ := ParseVersion("0.9.0")
	v2, _ := ParseVersion("1.0.0")

	if !c.Matches(v1) {
		t.Error("Expected 0.9.0 to match constraint <1.0.0")
	}

	if c.Matches(v2) {
		t.Error("Expected 1.0.0 to not match constraint <1.0.0")
	}
}

// TestConstraintLessEq tests less than or equal constraint
func TestConstraintLessEq(t *testing.T) {
	c, _ := ParseConstraint("<=1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.1")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint <=1.0.0")
	}

	if c.Matches(v2) {
		t.Error("Expected 1.0.1 to not match constraint <=1.0.0")
	}
}

// TestConstraintCaret tests caret range constraint
func TestConstraintCaret(t *testing.T) {
	c, _ := ParseConstraint("^1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.5.0")
	v3, _ := ParseVersion("2.0.0")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint ^1.0.0")
	}

	if !c.Matches(v2) {
		t.Error("Expected 1.5.0 to match constraint ^1.0.0")
	}

	if c.Matches(v3) {
		t.Error("Expected 2.0.0 to not match constraint ^1.0.0")
	}
}

// TestConstraintTilde tests tilde range constraint
func TestConstraintTilde(t *testing.T) {
	c, _ := ParseConstraint("~1.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.5")
	v3, _ := ParseVersion("1.1.0")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint ~1.0.0")
	}

	if !c.Matches(v2) {
		t.Error("Expected 1.0.5 to match constraint ~1.0.0")
	}

	if c.Matches(v3) {
		t.Error("Expected 1.1.0 to not match constraint ~1.0.0")
	}
}

// TestConstraintRange tests range constraint
func TestConstraintRange(t *testing.T) {
	c, _ := ParseConstraint("1.0.0 - 2.0.0")

	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.5.0")
	v3, _ := ParseVersion("2.0.0")
	v4, _ := ParseVersion("2.0.1")

	if !c.Matches(v1) {
		t.Error("Expected 1.0.0 to match constraint 1.0.0 - 2.0.0")
	}

	if !c.Matches(v2) {
		t.Error("Expected 1.5.0 to match constraint 1.0.0 - 2.0.0")
	}

	if !c.Matches(v3) {
		t.Error("Expected 2.0.0 to match constraint 1.0.0 - 2.0.0")
	}

	if c.Matches(v4) {
		t.Error("Expected 2.0.1 to not match constraint 1.0.0 - 2.0.0")
	}
}

// TestConstraintString tests constraint string representation
func TestConstraintString(t *testing.T) {
	c, _ := ParseConstraint("^1.0.0")
	if c.String() != "^1.0.0" {
		t.Errorf("Expected '^1.0.0', got %q", c.String())
	}

	c, _ = ParseConstraint("1.0.0 - 2.0.0")
	if c.String() != "1.0.0-2.0.0" {
		t.Errorf("Expected '1.0.0-2.0.0', got %q", c.String())
	}
}

// TestVersionComparison tests version comparison
func TestVersionComparison(t *testing.T) {
	v1, _ := ParseVersion("1.0.0")
	v2, _ := ParseVersion("1.0.1")
	v3, _ := ParseVersion("2.0.0")

	c := &Constraint{}

	// v1 < v2 < v3
	if !c.versionLess(v1, v2) {
		t.Error("Expected 1.0.0 < 1.0.1")
	}

	if !c.versionLess(v2, v3) {
		t.Error("Expected 1.0.1 < 2.0.0")
	}

	if c.versionGreater(v1, v2) {
		t.Error("Expected 1.0.0 not > 1.0.1")
	}

	// Equal test
	v4, _ := ParseVersion("1.0.0")
	if !c.versionEqual(v1, v4) {
		t.Error("Expected 1.0.0 == 1.0.0")
	}
}

// TestPrereleaseVersionComparison tests prerelease version comparison
func TestPrereleaseVersionComparison(t *testing.T) {
	v1, _ := ParseVersion("1.0.0-alpha")
	v2, _ := ParseVersion("1.0.0")

	c := &Constraint{}

	// Prerelease should be less than release
	if !c.versionLess(v1, v2) {
		t.Error("Expected 1.0.0-alpha < 1.0.0")
	}
}

// TestVersionNegativeNumbers tests version parsing with negative numbers
func TestVersionNegativeNumbers(t *testing.T) {
	_, err := ParseVersion("-1.0.0")
	if err == nil {
		t.Error("Expected error for negative major version")
	}

	_, err = ParseVersion("1.-1.0")
	if err == nil {
		t.Error("Expected error for negative minor version")
	}

	_, err = ParseVersion("1.0.-1")
	if err == nil {
		t.Error("Expected error for negative patch version")
	}
}
