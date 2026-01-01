// Package github - Package management
package github

import (
	"fmt"
)

// Package represents a GitHub package.
type Package struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	PackageType  string `json:"package_type"`
	Visibility   string `json:"visibility"`
	HTMLURL      string `json:"html_url"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Owner        User   `json:"owner"`
	Repository   *Repo  `json:"repository"`
	VersionCount int    `json:"version_count"`
}

// Repo represents minimal repository info.
type Repo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
}

// PackageVersion represents a version of a package.
type PackageVersion struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	PackageHTMLURL string   `json:"package_html_url"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	HTMLURL        string   `json:"html_url"`
	Metadata       Metadata `json:"metadata"`
}

// Metadata contains package version metadata.
type Metadata struct {
	PackageType string    `json:"package_type"`
	Container   Container `json:"container"`
}

// Container contains container-specific metadata.
type Container struct {
	Tags []string `json:"tags"`
}

// ListUserPackages returns all packages for a user.
func (c *Client) ListUserPackages(username, packageType string) ([]Package, error) {
	path := fmt.Sprintf("/users/%s/packages?package_type=%s", username, packageType)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var packages []Package
	if err := handleResponse(resp, &packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// ListOrgPackages returns all packages for an organization.
func (c *Client) ListOrgPackages(org, packageType string) ([]Package, error) {
	path := fmt.Sprintf("/orgs/%s/packages?package_type=%s", org, packageType)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var packages []Package
	if err := handleResponse(resp, &packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// GetPackage returns a specific package.
func (c *Client) GetPackage(packageType, owner, packageName string) (*Package, error) {
	path := fmt.Sprintf("/users/%s/packages/%s/%s", owner, packageType, packageName)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var pkg Package
	if err := handleResponse(resp, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// ListPackageVersions returns all versions of a package.
func (c *Client) ListPackageVersions(packageType, owner, packageName string) ([]PackageVersion, error) {
	path := fmt.Sprintf("/users/%s/packages/%s/%s/versions", owner, packageType, packageName)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var versions []PackageVersion
	if err := handleResponse(resp, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// DeletePackage deletes a package.
func (c *Client) DeletePackage(packageType, owner, packageName string) error {
	path := fmt.Sprintf("/users/%s/packages/%s/%s", owner, packageType, packageName)
	resp, err := c.delete(path)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// DeletePackageVersion deletes a specific version of a package.
func (c *Client) DeletePackageVersion(packageType, owner, packageName string, versionID int64) error {
	path := fmt.Sprintf("/users/%s/packages/%s/%s/versions/%d", owner, packageType, packageName, versionID)
	resp, err := c.delete(path)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// AllPackageTypes returns all supported package types.
func AllPackageTypes() []string {
	return []string{"npm", "maven", "rubygems", "docker", "nuget", "container"}
}
