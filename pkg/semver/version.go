package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// NormalizeVersion strips range indicators (^, ~, etc.) from version strings
// to get a clean version string for parsing
func NormalizeVersion(version string) string {
	// Trim spaces first
	version = strings.TrimSpace(version)

	// Remove range indicators
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, "<")
	version = strings.TrimPrefix(version, "=")

	return strings.TrimSpace(version)
}

// Parse parses a semver string into a Version
func Parse(version string) (*Version, error) {
	// First normalize the version by removing any range indicators
	version = NormalizeVersion(version)

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
		// Try parsing it as a plain version without constraints
		version, err := Parse(constraint)
		if err != nil {
			return "", nil, fmt.Errorf("invalid constraint format: %s", constraint)
		}
		return "", version, nil
	}

	constraintType := matches[1]
	versionStr := matches[2]

	// Parse the version part (after normalizing it)
	version, err := Parse(versionStr)
	if err != nil {
		// Attempt to normalize and parse again if the first attempt failed
		normalizedVersion := NormalizeVersion(versionStr)
		version, err = Parse(normalizedVersion)
		if err != nil {
			return "", nil, err
		}
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
	v, err := Parse(version)
	if err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}

	constraintType, constraintVersion, err := ParseConstraint(constraint)
	if err != nil {
		return false, fmt.Errorf("invalid constraint: %w", err)
	}

	switch constraintType {
	case "^":
		// Compatible with changes that do not modify the left-most non-zero digit
		// ^1.2.3 => >=1.2.3 <2.0.0
		// ^0.2.3 => >=0.2.3 <0.3.0
		// ^0.0.3 => >=0.0.3 <0.0.4
		if constraintVersion.Major > 0 {
			return v.Major == constraintVersion.Major &&
				Compare(v, constraintVersion) >= 0, nil
		} else if constraintVersion.Minor > 0 {
			return v.Major == 0 &&
				v.Minor == constraintVersion.Minor &&
				Compare(v, constraintVersion) >= 0, nil
		} else {
			return v.Major == 0 &&
				v.Minor == 0 &&
				v.Patch == constraintVersion.Patch, nil
		}
	case "~":
		// Compatible with patch-level changes
		// ~1.2.3 => >=1.2.3 <1.3.0
		return v.Major == constraintVersion.Major &&
			v.Minor == constraintVersion.Minor &&
			v.Patch >= constraintVersion.Patch, nil
	case ">":
		return Compare(v, constraintVersion) > 0, nil
	case ">=":
		return Compare(v, constraintVersion) >= 0, nil
	case "<":
		return Compare(v, constraintVersion) < 0, nil
	case "<=":
		return Compare(v, constraintVersion) <= 0, nil
	case "=", "":
		// Exact match
		return Compare(v, constraintVersion) == 0, nil
	default:
		return false, fmt.Errorf("unsupported constraint type: %s", constraintType)
	}
}
