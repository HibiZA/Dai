package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
		wantErr  bool
	}{
		{"1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"1.2", &Version{Major: 1, Minor: 2, Patch: 0}, false},
		{"1", &Version{Major: 1, Minor: 0, Patch: 0}, false},
		{"1.2.3-beta", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta"}, false},
		{"1.2.3+build", &Version{Major: 1, Minor: 2, Patch: 3, Build: "build"}, false},
		{"1.2.3-beta+build", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta", Build: "build"}, false},
		{"v1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"invalid", nil, true},
	}

	for _, test := range tests {
		v, err := Parse(test.input)
		if (err != nil) != test.wantErr {
			t.Errorf("Parse(%q) error = %v, wantErr %v", test.input, err, test.wantErr)
			continue
		}

		if test.wantErr {
			continue
		}

		if v.Major != test.expected.Major ||
			v.Minor != test.expected.Minor ||
			v.Patch != test.expected.Patch ||
			v.Prerelease != test.expected.Prerelease ||
			v.Build != test.expected.Build {
			t.Errorf("Parse(%q) = %+v, want %+v", test.input, v, test.expected)
		}
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		input           string
		expectedType    string
		expectedVersion *Version
		wantErr         bool
	}{
		{"^1.2.3", "^", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"~1.2.3", "~", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{">1.2.3", ">", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"<1.2.3", "<", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{">=1.2.3", ">=", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"<=1.2.3", "<=", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"=1.2.3", "=", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"1.2.3", "", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"invalid", "", nil, true},
	}

	for _, test := range tests {
		constraintType, v, err := ParseConstraint(test.input)
		if (err != nil) != test.wantErr {
			t.Errorf("ParseConstraint(%q) error = %v, wantErr %v", test.input, err, test.wantErr)
			continue
		}

		if test.wantErr {
			continue
		}

		if constraintType != test.expectedType {
			t.Errorf("ParseConstraint(%q) constraint type = %q, want %q", test.input, constraintType, test.expectedType)
		}

		if v.Major != test.expectedVersion.Major ||
			v.Minor != test.expectedVersion.Minor ||
			v.Patch != test.expectedVersion.Patch ||
			v.Prerelease != test.expectedVersion.Prerelease ||
			v.Build != test.expectedVersion.Build {
			t.Errorf("ParseConstraint(%q) version = %+v, want %+v", test.input, v, test.expectedVersion)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       *Version
		v2       *Version
		expected int
	}{
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 1, Minor: 2, Patch: 3}, 0},
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 1, Minor: 2, Patch: 4}, -1},
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 1, Minor: 3, Patch: 0}, -1},
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 2, Minor: 0, Patch: 0}, -1},
		{&Version{Major: 1, Minor: 2, Patch: 4}, &Version{Major: 1, Minor: 2, Patch: 3}, 1},
		{&Version{Major: 1, Minor: 3, Patch: 0}, &Version{Major: 1, Minor: 2, Patch: 3}, 1},
		{&Version{Major: 2, Minor: 0, Patch: 0}, &Version{Major: 1, Minor: 2, Patch: 3}, 1},
		{&Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"}, &Version{Major: 1, Minor: 2, Patch: 3}, -1},
		{&Version{Major: 1, Minor: 2, Patch: 3}, &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"}, 1},
	}

	for _, test := range tests {
		result := Compare(test.v1, test.v2)
		if result != test.expected {
			t.Errorf("Compare(%+v, %+v) = %d, want %d", test.v1, test.v2, result, test.expected)
		}
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		expected   bool
		wantErr    bool
	}{
		// Exact match
		{"1.2.3", "1.2.3", true, false},
		{"1.2.3", "1.2.4", false, false},

		// Caret range
		{"1.2.3", "^1.2.0", true, false},
		{"1.3.0", "^1.2.3", true, false},
		{"2.0.0", "^1.2.3", false, false},
		{"0.2.5", "^0.2.3", true, false},
		{"0.3.0", "^0.2.3", false, false},
		{"0.0.4", "^0.0.3", false, false},

		// Tilde range
		{"1.2.3", "~1.2.3", true, false},
		{"1.2.4", "~1.2.3", true, false},
		{"1.3.0", "~1.2.3", false, false},

		// Comparison operators
		{"1.2.3", ">1.2.2", true, false},
		{"1.2.3", ">1.2.3", false, false},
		{"1.2.3", ">=1.2.3", true, false},
		{"1.2.3", "<1.2.4", true, false},
		{"1.2.3", "<1.2.3", false, false},
		{"1.2.3", "<=1.2.3", true, false},

		// Error cases
		{"invalid", "1.2.3", false, true},
		{"1.2.3", "invalid", false, true},
	}

	for _, test := range tests {
		compatible, err := IsCompatible(test.version, test.constraint)
		if (err != nil) != test.wantErr {
			t.Errorf("IsCompatible(%q, %q) error = %v, wantErr %v", test.version, test.constraint, err, test.wantErr)
			continue
		}

		if !test.wantErr && compatible != test.expected {
			t.Errorf("IsCompatible(%q, %q) = %v, want %v", test.version, test.constraint, compatible, test.expected)
		}
	}
}
