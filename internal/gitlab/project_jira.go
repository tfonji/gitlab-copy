package gitlab

// JiraIntegration represents the Jira integration config for a project.
// Properties are returned as a flat map by the API — credentials (password,
// api_url tokens) are typically masked or empty in GET responses, so the
// copy will carry whatever the API returns. Credentials need manual
// verification on dest after copying.
type JiraIntegration struct {
	Active     bool           `json:"active"`
	Slug       string         `json:"slug"`
	Properties map[string]any `json:"properties"`
}

// --- Read ---

// GetProjectJiraIntegration fetches the Jira integration for a project.
// Returns nil if not configured or not accessible.
func (c *Client) GetProjectJiraIntegration(projectPath string) (*JiraIntegration, error) {
	var j JiraIntegration
	err := c.get("/projects/"+encodePath(projectPath)+"/integrations/jira", nil, &j)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	if !j.Active {
		return nil, nil
	}
	return &j, nil
}

// --- Write ---

// SetProjectJiraIntegration creates or updates the Jira integration on dest
// via PUT. The properties map is sent as top-level fields in the request body,
// which is what the GitLab integrations API expects.
func (c *Client) SetProjectJiraIntegration(projectPath string, properties map[string]any) error {
	return c.put("/projects/"+encodePath(projectPath)+"/integrations/jira", properties, nil)
}
