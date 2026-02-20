// Package github - Repository management.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Repository represents a GitHub repository.
type Repository struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	Private         bool     `json:"private"`
	Fork            bool     `json:"fork"`
	Archived        bool     `json:"archived"`
	Disabled        bool     `json:"disabled"`
	HTMLURL         string   `json:"html_url"`
	CloneURL        string   `json:"clone_url"`
	SSHURL          string   `json:"ssh_url"`
	DefaultBranch   string   `json:"default_branch"`
	Language        string   `json:"language"`
	ForksCount      int      `json:"forks_count"`
	StargazersCount int      `json:"stargazers_count"`
	WatchersCount   int      `json:"watchers_count"`
	OpenIssuesCount int      `json:"open_issues_count"`
	Topics          []string `json:"topics"`
	Visibility      string   `json:"visibility"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	PushedAt        string   `json:"pushed_at"`
}

// CreateRepoParams contains parameters for creating a repository.
type CreateRepoParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	AutoInit    bool   `json:"auto_init,omitempty"`
}

// CreateRepo creates a new repository.
func (c *Client) CreateRepo(params CreateRepoParams) (*Repository, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.post("/user/repos", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var repo Repository
	if err := handleResponse(resp, &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepo returns a repository.
func (c *Client) GetRepo(owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var repository Repository
	if err := handleResponse(resp, &repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// ListUserRepos returns repositories for the authenticated user.
func (c *Client) ListUserRepos() ([]Repository, error) {
	resp, err := c.get("/user/repos?sort=updated&per_page=30")
	if err != nil {
		return nil, err
	}

	var repos []Repository
	if err := handleResponse(resp, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// DeleteRepo deletes a repository.
func (c *Client) DeleteRepo(owner, repo string) error {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	resp, err := c.delete(path)
	if err != nil {
		return err
	}

	return handleResponse(resp, nil)
}

// RepoExists checks if a repository exists.
func (c *Client) RepoExists(owner, repo string) bool {
	_, err := c.GetRepo(owner, repo)
	return err == nil
}
