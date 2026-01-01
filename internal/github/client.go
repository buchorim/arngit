// Package github provides GitHub API client.
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL   = "https://api.github.com"
	userAgent = "arngit/1.0"
)

// Client is a GitHub API client.
type Client struct {
	httpClient *http.Client
	token      string
	username   string
}

// RateLimit contains rate limit information.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// NewClient creates a new GitHub client.
func NewClient(username, token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		token:      token,
		username:   username,
	}
}

// request makes an authenticated request to the GitHub API.
func (c *Client) request(method, path string, body io.Reader) (*http.Response, error) {
	url := baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// get makes a GET request.
func (c *Client) get(path string) (*http.Response, error) {
	return c.request("GET", path, nil)
}

// post makes a POST request.
func (c *Client) post(path string, body io.Reader) (*http.Response, error) {
	return c.request("POST", path, body)
}

// delete makes a DELETE request.
func (c *Client) delete(path string) (*http.Response, error) {
	return c.request("DELETE", path, nil)
}

// parseRateLimit extracts rate limit info from response headers.
func parseRateLimit(resp *http.Response) *RateLimit {
	limit, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Limit"))
	remaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	resetUnix, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)

	return &RateLimit{
		Limit:     limit,
		Remaining: remaining,
		Reset:     time.Unix(resetUnix, 0),
	}
}

// APIError represents a GitHub API error.
type APIError struct {
	StatusCode int
	Message    string
	RateLimit  *RateLimit
}

func (e *APIError) Error() string {
	if e.StatusCode == 403 && e.RateLimit != nil && e.RateLimit.Remaining == 0 {
		return fmt.Sprintf("GitHub API rate limit exceeded. Resets at %s", e.RateLimit.Reset.Format("15:04:05"))
	}
	return fmt.Sprintf("GitHub API error (%d): %s", e.StatusCode, e.Message)
}

// handleResponse processes the API response and decodes JSON.
func handleResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Message string `json:"message"`
		}
		json.Unmarshal(body, &errResp)

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Message,
			RateLimit:  parseRateLimit(resp),
		}
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// User represents a GitHub user.
type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// GetUser returns the authenticated user.
func (c *Client) GetUser() (*User, error) {
	resp, err := c.get("/user")
	if err != nil {
		return nil, err
	}

	var user User
	if err := handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// PATInfo contains information about the PAT.
type PATInfo struct {
	Valid     bool
	User      *User
	Scopes    []string
	RateLimit *RateLimit
}

// ValidatePAT checks if the PAT is valid and returns scope info.
func (c *Client) ValidatePAT() (*PATInfo, error) {
	resp, err := c.get("/user")
	if err != nil {
		return &PATInfo{Valid: false}, err
	}
	defer resp.Body.Close()

	rateLimit := parseRateLimit(resp)

	if resp.StatusCode == 401 {
		return &PATInfo{Valid: false, RateLimit: rateLimit}, nil
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Message string `json:"message"`
		}
		json.Unmarshal(body, &errResp)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Message,
			RateLimit:  rateLimit,
		}
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// Parse scopes from header
	scopeHeader := resp.Header.Get("X-OAuth-Scopes")
	var scopes []string
	if scopeHeader != "" {
		for _, s := range strings.Split(scopeHeader, ",") {
			scopes = append(scopes, strings.TrimSpace(s))
		}
	}

	return &PATInfo{
		Valid:     true,
		User:      &user,
		Scopes:    scopes,
		RateLimit: rateLimit,
	}, nil
}

// HasScope checks if the PAT has a specific scope.
func (p *PATInfo) HasScope(scope string) bool {
	for _, s := range p.Scopes {
		if s == scope {
			return true
		}
		// Check parent scopes (e.g., "repo" includes "repo:status")
		if strings.HasPrefix(scope, s+":") {
			return true
		}
	}
	return false
}

// ScopeDescription returns a human-readable description of a scope.
func ScopeDescription(scope string) string {
	descriptions := map[string]string{
		"repo":                      "Full control of private repositories",
		"repo:status":               "Access commit status",
		"repo_deployment":           "Access deployment status",
		"public_repo":               "Access public repositories",
		"repo:invite":               "Access repository invitations",
		"workflow":                  "Update GitHub Action workflows",
		"write:packages":            "Upload packages to GitHub Package Registry",
		"read:packages":             "Download packages from GitHub Package Registry",
		"delete:packages":           "Delete packages from GitHub Package Registry",
		"admin:org":                 "Full control of orgs and teams",
		"write:org":                 "Read and write org membership",
		"read:org":                  "Read org membership",
		"admin:public_key":          "Full control of user public keys",
		"write:public_key":          "Write user public keys",
		"read:public_key":           "Read user public keys",
		"admin:repo_hook":           "Full control of repository hooks",
		"write:repo_hook":           "Write repository hooks",
		"read:repo_hook":            "Read repository hooks",
		"admin:org_hook":            "Full control of organization hooks",
		"gist":                      "Create gists",
		"notifications":             "Access notifications",
		"user":                      "Update all user data",
		"read:user":                 "Read user profile data",
		"user:email":                "Access user email addresses",
		"user:follow":               "Follow and unfollow users",
		"project":                   "Full control of projects",
		"read:project":              "Read access to projects",
		"delete_repo":               "Delete repositories",
		"write:discussion":          "Read and write discussions",
		"read:discussion":           "Read discussions",
		"admin:enterprise":          "Full control of enterprise",
		"manage_billing:enterprise": "Read and write enterprise billing data",
		"read:enterprise":           "Read enterprise profile data",
	}

	if desc, ok := descriptions[scope]; ok {
		return desc
	}
	return "Unknown scope"
}
