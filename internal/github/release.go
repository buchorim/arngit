// Package github - Release management.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Release represents a GitHub release.
type Release struct {
	ID          int64   `json:"id"`
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	CreatedAt   string  `json:"created_at"`
	PublishedAt string  `json:"published_at"`
	HTMLURL     string  `json:"html_url"`
	UploadURL   string  `json:"upload_url"`
	Assets      []Asset `json:"assets"`
	Author      User    `json:"author"`
}

// Asset represents a release asset.
type Asset struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Label         string `json:"label"`
	ContentType   string `json:"content_type"`
	Size          int64  `json:"size"`
	DownloadCount int    `json:"download_count"`
	DownloadURL   string `json:"browser_download_url"`
}

// ListReleases returns all releases for a repository.
func (c *Client) ListReleases(owner, repo string) ([]Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases", owner, repo)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := handleResponse(resp, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// GetLatestRelease returns the latest release.
func (c *Client) GetLatestRelease(owner, repo string) (*Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := handleResponse(resp, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// CreateReleaseParams contains parameters for creating a release.
type CreateReleaseParams struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name,omitempty"`
	Body            string `json:"body,omitempty"`
	Draft           bool   `json:"draft,omitempty"`
	Prerelease      bool   `json:"prerelease,omitempty"`
}

// CreateRelease creates a new release.
func (c *Client) CreateRelease(owner, repo string, params CreateReleaseParams) (*Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases", owner, repo)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var release Release
	if err := handleResponse(resp, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// UploadReleaseAsset uploads a file to a release.
func (c *Client) UploadReleaseAsset(uploadURL, filename, contentType string, data []byte) (*Asset, error) {
	// Remove the template part and add filename
	uploadURL = strings.Split(uploadURL, "{")[0]
	uploadURL = uploadURL + "?name=" + filename

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)
	req.ContentLength = int64(len(data))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var asset Asset
	if err := handleResponse(resp, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}
