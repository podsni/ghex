package update

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // Pre-release identifier (e.g., "beta.1", "rc.1")
}

// versionRegex matches semantic version strings
var versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)

// ParseVersion parses a version string into Version struct
// Accepts formats: "1.0.0", "v1.0.0", "1.0.0-beta.1", "v1.0.0-rc.1"
func ParseVersion(v string) (*Version, error) {
	v = strings.TrimSpace(v)
	matches := versionRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidVersion, v)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	pre := ""
	if len(matches) > 4 {
		pre = matches[4]
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
	}, nil
}

// String returns the version as a string (without 'v' prefix)
func (v *Version) String() string {
	if v.Pre != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Pre)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}


// Compare compares two versions
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *Version) Compare(other *Version) int {
	// Compare major
	if v.Major < other.Major {
		return -1
	}
	if v.Major > other.Major {
		return 1
	}

	// Compare minor
	if v.Minor < other.Minor {
		return -1
	}
	if v.Minor > other.Minor {
		return 1
	}

	// Compare patch
	if v.Patch < other.Patch {
		return -1
	}
	if v.Patch > other.Patch {
		return 1
	}

	// Compare pre-release
	// A version without pre-release is greater than one with pre-release
	// e.g., 1.0.0 > 1.0.0-beta.1
	if v.Pre == "" && other.Pre != "" {
		return 1
	}
	if v.Pre != "" && other.Pre == "" {
		return -1
	}
	if v.Pre == other.Pre {
		return 0
	}

	// Compare pre-release strings lexicographically
	if v.Pre < other.Pre {
		return -1
	}
	return 1
}

// IsNewerThan returns true if v is newer than other
func (v *Version) IsNewerThan(other *Version) bool {
	return v.Compare(other) > 0
}

// Equals returns true if v equals other
func (v *Version) Equals(other *Version) bool {
	return v.Compare(other) == 0
}

// CompareVersionStrings compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2, error if invalid
func CompareVersionStrings(v1, v2 string) (int, error) {
	ver1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}
	ver2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}
	return ver1.Compare(ver2), nil
}
