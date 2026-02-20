// Package automation - Semantic version management.
package automation

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Prefix     string // "v" or ""
}

// ParseVersion parses a version string.
func ParseVersion(s string) (*Version, error) {
	v := &Version{}

	if strings.HasPrefix(s, "v") {
		v.Prefix = "v"
		s = strings.TrimPrefix(s, "v")
	}

	// Split by dash for prerelease
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 2 {
		v.Prerelease = parts[1]
		s = parts[0]
	}

	// Parse major.minor.patch
	versionParts := strings.Split(s, ".")
	if len(versionParts) >= 1 {
		v.Major, _ = strconv.Atoi(versionParts[0])
	}
	if len(versionParts) >= 2 {
		v.Minor, _ = strconv.Atoi(versionParts[1])
	}
	if len(versionParts) >= 3 {
		v.Patch, _ = strconv.Atoi(versionParts[2])
	}

	return v, nil
}

// String returns the version as string.
func (v *Version) String() string {
	base := fmt.Sprintf("%s%d.%d.%d", v.Prefix, v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		return base + "-" + v.Prerelease
	}
	return base
}

// BumpMajor increments major version.
func (v *Version) BumpMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.Prerelease = ""
}

// BumpMinor increments minor version.
func (v *Version) BumpMinor() {
	v.Minor++
	v.Patch = 0
	v.Prerelease = ""
}

// BumpPatch increments patch version.
func (v *Version) BumpPatch() {
	v.Patch++
	v.Prerelease = ""
}

// GetCurrentVersion gets current version from git tags.
func GetCurrentVersion() (*Version, error) {
	tag, err := GetLatestTag()
	if err != nil {
		return &Version{Prefix: "v", Major: 0, Minor: 0, Patch: 0}, nil
	}
	return ParseVersion(tag)
}

// CreateVersionTag creates a new git tag for the version.
func CreateVersionTag(version, message string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", version, "-m", message)
	} else {
		args = append(args, version)
	}

	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// DetermineBumpType determines bump type from changelog commits.
func DetermineBumpType(commits []ChangelogCommit) string {
	for _, c := range commits {
		if c.Breaking {
			return "major"
		}
	}

	for _, c := range commits {
		if c.Type == "feat" || c.Type == "feature" {
			return "minor"
		}
	}

	return "patch"
}
