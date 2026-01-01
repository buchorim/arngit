// Package smart - Pull Request management
package smart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
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

// PRRepo represents a repository.
type PRRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// PRUser represents a GitHub user.
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

// PRClient handles PR operations.
type PRClient struct {
	token    string
	username string
	client   *http.Client
}

// NewPRClient creates a new PR client.
func NewPRClient(username, token string) *PRClient {
	return &PRClient{
		token:    token,
		username: username,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// CreatePR creates a new pull request.
func (c *PRClient) CreatePR(owner, repo string, params CreatePRParams) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// ListPRs lists pull requests.
func (c *PRClient) ListPRs(owner, repo, state string) ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=%s", owner, repo, state)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return prs, nil
}

// GetCurrentBranch returns the current git branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDefaultBranch returns the default branch (main or master).
func GetDefaultBranch() string {
	// Try to get from remote
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: check if main exists
	cmd = exec.Command("git", "rev-parse", "--verify", "main")
	if cmd.Run() == nil {
		return "main"
	}

	return "master"
}

// GeneratePRTitle generates a PR title from branch name.
func GeneratePRTitle(branch string) string {
	// Remove common prefixes
	prefixes := []string{"feature/", "feat/", "fix/", "bugfix/", "hotfix/", "release/"}
	title := branch

	for _, prefix := range prefixes {
		if strings.HasPrefix(title, prefix) {
			title = strings.TrimPrefix(title, prefix)
			break
		}
	}

	// Replace separators with spaces
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize first letter
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	return title
}

// GeneratePRBody generates a PR body template.
func GeneratePRBody(commits []Commit) string {
	var sb strings.Builder

	sb.WriteString("## Changes\n\n")

	if len(commits) > 0 {
		for _, c := range commits {
			sb.WriteString(fmt.Sprintf("- %s\n", c.Message))
		}
	} else {
		sb.WriteString("- Describe your changes here\n")
	}

	sb.WriteString("\n## Type of Change\n\n")
	sb.WriteString("- [ ] Bug fix\n")
	sb.WriteString("- [ ] New feature\n")
	sb.WriteString("- [ ] Breaking change\n")
	sb.WriteString("- [ ] Documentation update\n")

	sb.WriteString("\n## Checklist\n\n")
	sb.WriteString("- [ ] I have tested these changes\n")
	sb.WriteString("- [ ] I have updated the documentation\n")
	sb.WriteString("- [ ] My code follows the project style\n")

	return sb.String()
}
