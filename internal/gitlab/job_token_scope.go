package gitlab

import (
	"fmt"
	"net/url"
)

// JobTokenScope holds the job token access settings for a project.
type JobTokenScope struct {
	InboundEnabled                     bool `json:"inbound_enabled"`
	CiPushRepositoryForJobTokenAllowed bool `json:"ci_push_repository_for_job_token_allowed"`
}

// JobTokenAllowlistProject is a project in the job token allowlist.
type JobTokenAllowlistProject struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// JobTokenAllowlistGroup is a group in the job token groups allowlist.
type JobTokenAllowlistGroup struct {
	ID       int    `json:"id"`
	FullPath string `json:"full_path"`
}

// --- Read ---

func (c *Client) GetJobTokenScope(projectID int) (*JobTokenScope, error) {
	// inbound_enabled lives on the job_token_scope endpoint
	var scope JobTokenScope
	if err := c.get(fmt.Sprintf("/projects/%d/job_token_scope", projectID), nil, &scope); err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}

	// ci_push_repository_for_job_token_allowed lives on the project endpoint
	var proj struct {
		CiPushRepositoryForJobTokenAllowed bool `json:"ci_push_repository_for_job_token_allowed"`
	}
	if err := c.get(fmt.Sprintf("/projects/%d", projectID), nil, &proj); err == nil {
		scope.CiPushRepositoryForJobTokenAllowed = proj.CiPushRepositoryForJobTokenAllowed
	}

	return &scope, nil
}

func (c *Client) GetJobTokenAllowlist(projectID int) ([]JobTokenAllowlistProject, error) {
	var all []JobTokenAllowlistProject
	page := 1
	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		var batch []JobTokenAllowlistProject
		if err := c.get(fmt.Sprintf("/projects/%d/job_token_scope/allowlist", projectID), params, &batch); err != nil {
			if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
				return nil, nil
			}
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) GetJobTokenGroupsAllowlist(projectID int) ([]JobTokenAllowlistGroup, error) {
	var all []JobTokenAllowlistGroup
	page := 1
	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		var batch []JobTokenAllowlistGroup
		if err := c.get(fmt.Sprintf("/projects/%d/job_token_scope/groups_allowlist", projectID), params, &batch); err != nil {
			if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
				return nil, nil
			}
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// --- Write ---

// SetJobTokenScope sets the inbound_enabled toggle on a project.
// Note: GET returns inbound_enabled but PATCH takes enabled (GitLab API inconsistency).
func (c *Client) SetJobTokenScope(projectID int, enabled bool) error {
	return c.patch(fmt.Sprintf("/projects/%d/job_token_scope", projectID), map[string]bool{
		"enabled": enabled,
	})
}

// SetJobTokenPushPermission sets ci_push_repository_for_job_token_allowed via the projects API.
func (c *Client) SetJobTokenPushPermission(projectID int, allowed bool) error {
	return c.put(fmt.Sprintf("/projects/%d", projectID), map[string]bool{
		"ci_push_repository_for_job_token_allowed": allowed,
	}, nil)
}

// AddJobTokenAllowlistProject adds a project to the job token allowlist by dest project ID.
func (c *Client) AddJobTokenAllowlistProject(projectID, targetProjectID int) error {
	return c.post(fmt.Sprintf("/projects/%d/job_token_scope/allowlist", projectID), map[string]int{
		"target_project_id": targetProjectID,
	}, nil)
}

// AddJobTokenAllowlistGroup adds a group to the job token groups allowlist by dest group ID.
func (c *Client) AddJobTokenAllowlistGroup(projectID, targetGroupID int) error {
	return c.post(fmt.Sprintf("/projects/%d/job_token_scope/groups_allowlist", projectID), map[string]int{
		"target_group_id": targetGroupID,
	}, nil)
}
