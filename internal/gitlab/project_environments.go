package gitlab

import "net/url"

// Environment represents a deployment environment on a project.
type Environment struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	ExternalURL string `json:"external_url"`
	State       string `json:"state"`
}

// EnvironmentRequest is the write body for POST /projects/:id/environments.
type EnvironmentRequest struct {
	Name        string `json:"name"`
	ExternalURL string `json:"external_url,omitempty"`
}

// --- Read ---

// GetProjectEnvironments fetches all environments for a project.
// Only returns environments with state "available" (not stopped/deleted).
func (c *Client) GetProjectEnvironments(projectPath string) ([]Environment, error) {
	var all []Environment
	params := url.Values{}
	params.Set("per_page", "100")
	params.Set("states", "available")
	err := c.get("/projects/"+encodePath(projectPath)+"/environments", params, &all)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return all, nil
}

// --- Write ---

// CreateProjectEnvironment creates a deployment environment on a project.
func (c *Client) CreateProjectEnvironment(projectPath string, req EnvironmentRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/environments", req, nil)
}
