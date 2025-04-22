package semver

import (
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"^1.2.3", "1.2.3"},
		{"~1.2.3", "1.2.3"},
		{">1.2.3", "1.2.3"},
		{">=1.2.3", "1.2.3"},
		{"<1.2.3", "1.2.3"},
		{"<=1.2.3", "1.2.3"},
		{"=1.2.3", "1.2.3"},
		{"^0.2.3", "0.2.3"},
		{"  ^1.2.3  ", "1.2.3"},
	}

	for _, test := range tests {
		result := NormalizeVersion(test.input)
		if result != test.expected {
			t.Errorf("NormalizeVersion(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		version  string
		expected *Version
		valid    bool
	}{
		{"1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"v1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"1.2", &Version{Major: 1, Minor: 2, Patch: 0}, true},
		{"1", &Version{Major: 1, Minor: 0, Patch: 0}, true},
		{"1.2.3-beta", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta"}, true},
		{"1.2.3+build", &Version{Major: 1, Minor: 2, Patch: 3, Build: "build"}, true},
		{"1.2.3-beta+build", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta", Build: "build"}, true},
		{"^1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"~1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"invalid", nil, false},
	}

	for _, test := range tests {
		result, err := Parse(test.version)
		if test.valid && err != nil {
			t.Errorf("Parse(%q) returned error: %v", test.version, err)
		} else if !test.valid && err == nil {
			t.Errorf("Parse(%q) should have returned an error", test.version)
		}

		if test.valid {
			if result.Major != test.expected.Major ||
				result.Minor != test.expected.Minor ||
				result.Patch != test.expected.Patch ||
				result.Prerelease != test.expected.Prerelease ||
				result.Build != test.expected.Build {
				t.Errorf("Parse(%q) = %+v, expected %+v", test.version, result, test.expected)
			}
		}
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		constraint      string
		expectedType    string
		expectedVersion *Version
		valid           bool
	}{
		{"^1.2.3", "^", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"~1.2.3", "~", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{">1.2.3", ">", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{">=1.2.3", ">=", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"<1.2.3", "<", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"<=1.2.3", "<=", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"=1.2.3", "=", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"1.2.3", "", &Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"invalid", "", nil, false},
	}

	for _, test := range tests {
		constraintType, version, err := ParseConstraint(test.constraint)
		if test.valid && err != nil {
			t.Errorf("ParseConstraint(%q) returned error: %v", test.constraint, err)
		} else if !test.valid && err == nil {
			t.Errorf("ParseConstraint(%q) should have returned an error", test.constraint)
		}

		if test.valid {
			if constraintType != test.expectedType {
				t.Errorf("ParseConstraint(%q) returned constraint type %q, expected %q",
					test.constraint, constraintType, test.expectedType)
			}

			if version.Major != test.expectedVersion.Major ||
				version.Minor != test.expectedVersion.Minor ||
				version.Patch != test.expectedVersion.Patch {
				t.Errorf("ParseConstraint(%q) returned version %+v, expected %+v",
					test.constraint, version, test.expectedVersion)
			}
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
