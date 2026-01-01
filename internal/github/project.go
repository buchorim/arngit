// Package github - Project management
package github

import (
	"fmt"
)

// Project represents a GitHub project (Projects V2).
type Project struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Creator   User   `json:"creator"`
}

// ProjectColumn represents a column in a project.
type ProjectColumn struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CardsURL  string `json:"cards_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ProjectCard represents a card in a project column.
type ProjectCard struct {
	ID         int64  `json:"id"`
	Note       string `json:"note"`
	ContentURL string `json:"content_url"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ListRepoProjects returns all projects for a repository.
func (c *Client) ListRepoProjects(owner, repo string) ([]Project, error) {
	path := fmt.Sprintf("/repos/%s/%s/projects", owner, repo)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListOrgProjects returns all projects for an organization.
func (c *Client) ListOrgProjects(org string) ([]Project, error) {
	path := fmt.Sprintf("/orgs/%s/projects", org)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListUserProjects returns all projects for a user.
func (c *Client) ListUserProjects(username string) ([]Project, error) {
	path := fmt.Sprintf("/users/%s/projects", username)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := handleResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject returns a specific project.
func (c *Client) GetProject(projectID int64) (*Project, error) {
	path := fmt.Sprintf("/projects/%d", projectID)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := handleResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// ListProjectColumns returns all columns for a project.
func (c *Client) ListProjectColumns(projectID int64) ([]ProjectColumn, error) {
	path := fmt.Sprintf("/projects/%d/columns", projectID)
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}

	var columns []ProjectColumn
	if err := handleResponse(resp, &columns); err != nil {
		return nil, err
	}

	return columns, nil
}
