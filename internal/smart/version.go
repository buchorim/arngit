// Package smart - Version management
package smart

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	// Check for v prefix
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

// SetPrerelease sets prerelease suffix.
func (v *Version) SetPrerelease(pre string) {
	v.Prerelease = pre
}

// GetCurrentVersion gets current version from git tags.
func GetCurrentVersion() (*Version, error) {
	tag, err := GetLatestTag()
	if err != nil {
		return &Version{Prefix: "v", Major: 0, Minor: 0, Patch: 0}, nil
	}
	return ParseVersion(tag)
}

// CreateTag creates a new git tag.
func CreateTag(version string, message string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", version, "-m", message)
	} else {
		args = append(args, version)
	}

	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// PushTag pushes a tag to remote.
func PushTag(remote, tag string) error {
	cmd := exec.Command("git", "push", remote, tag)
	return cmd.Run()
}

// UpdateVersionInFile updates version string in a file.
func UpdateVersionInFile(filePath, oldVersion, newVersion string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Replace version string
	updated := strings.ReplaceAll(string(content), oldVersion, newVersion)

	return os.WriteFile(filePath, []byte(updated), 0644)
}

// FindVersionInFile finds version patterns in a file.
func FindVersionInFile(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Find version-like patterns
	re := regexp.MustCompile(`v?\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?`)
	matches := re.FindAllString(string(content), -1)

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}

	return unique, nil
}

// BumpType represents the type of version bump.
type BumpType string

const (
	BumpTypeMajor      BumpType = "major"
	BumpTypeMinor      BumpType = "minor"
	BumpTypePatch      BumpType = "patch"
	BumpTypePrerelease BumpType = "prerelease"
)

// DetermineBumpType determines bump type from commit types.
func DetermineBumpType(commits []Commit) BumpType {
	for _, c := range commits {
		if c.Breaking {
			return BumpTypeMajor
		}
	}

	for _, c := range commits {
		if c.Type == "feat" || c.Type == "feature" {
			return BumpTypeMinor
		}
	}

	return BumpTypePatch
}
