// Package github - Pull Request management.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PullRequest represents a GitHub PR.
type PullRequest struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	Head      PRRef  `json:"head"`
	Base      PRRef  `json:"base"`
	CreatedAt string `json:"created_at"`
	User      PRUser `json:"user"`
}

// PRRef represents a branch reference.
type PRRef struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo PRRepo `json:"repo"`
}

// PRRepo represents a repository in a PR reference.
type PRRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// PRUser represents a GitHub user in a PR.
type PRUser struct {
	Login string `json:"login"`
}

// CreatePRParams contains parameters for creating a PR.
type CreatePRParams struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

// CreatePR creates a new pull request.
func (c *Client) CreatePR(owner, repo string, params CreatePRParams) (*PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	if err := handleResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// ListPRs lists pull requests.
func (c *Client) ListPRs(owner, repo, state string) ([]PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls?state=%s", owner, repo, state)

	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := handleResponse(resp, &prs); err != nil {
		return nil, err
	}

	return prs, nil
}
