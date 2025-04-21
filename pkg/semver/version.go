package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version represents a semver version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

var semverRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z-.]+))?(?:\+([0-9A-Za-z-.]+))?$`)
var constraintRegex = regexp.MustCompile(`^([~^><=]*)(.+)$`)

// SemverRegexForDebug returns the semver regex for debugging
func SemverRegexForDebug() *regexp.Regexp {
	return semverRegex
}

// ConstraintRegexForDebug returns the constraint regex for debugging
func ConstraintRegexForDebug() *regexp.Regexp {
	return constraintRegex
}

// Parse parses a semver string into a Version
func Parse(version string) (*Version, error) {
	// Now parse the clean version
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("invalid semver: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor := 0
	if len(matches[2]) > 0 {
		minor, _ = strconv.Atoi(matches[2])
	}
	patch := 0
	if len(matches[3]) > 0 {
		patch, _ = strconv.Atoi(matches[3])
	}

	v := &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}

	return v, nil
}

// ParseConstraint parses a version constraint (e.g., ^1.2.3) and returns
// the constraint type and the version it applies to
func ParseConstraint(constraint string) (string, *Version, error) {
	matches := constraintRegex.FindStringSubmatch(constraint)
	if matches == nil || len(matches) < 3 {
		return "", nil, fmt.Errorf("invalid constraint format: %s", constraint)
	}

	constraintType := matches[1]
	versionStr := matches[2]

	// Parse the version part
	version, err := Parse(versionStr)
	if err != nil {
		return "", nil, err
	}

	return constraintType, version, nil
}

// Compare compares two versions
// returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func Compare(v1, v2 *Version) int {
	if v1.Major != v2.Major {
		if v1.Major > v2.Major {
			return 1
		}
		return -1
	}

	if v1.Minor != v2.Minor {
		if v1.Minor > v2.Minor {
			return 1
		}
		return -1
	}

	if v1.Patch != v2.Patch {
		if v1.Patch > v2.Patch {
			return 1
		}
		return -1
	}

	// If we get here, the core version is the same, so we need to compare prerelease

	// No prerelease is greater than any prerelease
	if v1.Prerelease == "" && v2.Prerelease != "" {
		return 1
	}
	if v1.Prerelease != "" && v2.Prerelease == "" {
		return -1
	}
	if v1.Prerelease == "" && v2.Prerelease == "" {
		return 0
	}

	// Both have prerelease versions, compare them
	// TODO: Implement proper prerelease comparison
	if v1.Prerelease < v2.Prerelease {
		return -1
	}
	if v1.Prerelease > v2.Prerelease {
		return 1
	}

	return 0
}

// String converts a Version to a string
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}
	if v.Build != "" {
		result += "+" + v.Build
	}
	return result
}

// IsCompatible checks if the version is compatible with the constraint
func IsCompatible(version, constraint string) (bool, error) {
	// TODO: Implement proper semver constraint checking
	// For now, just check if the versions match
	return version == constraint, nil
}
