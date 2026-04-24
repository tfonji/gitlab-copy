package gitlab

// GroupJiraIntegration represents the Jira integration config for a group.
type GroupJiraIntegration struct {
	Active     bool           `json:"active"`
	Slug       string         `json:"slug"`
	Properties map[string]any `json:"properties"`
}

// --- Read ---

// GetGroupJiraIntegration fetches the Jira integration for a group.
// Returns nil if not configured or not accessible.
func (c *Client) GetGroupJiraIntegration(groupPath string) (*GroupJiraIntegration, error) {
	var j GroupJiraIntegration
	err := c.get("/groups/"+encodePath(groupPath)+"/integrations/jira", nil, &j)
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

// SetGroupJiraIntegration creates or updates the Jira integration on a group.
func (c *Client) SetGroupJiraIntegration(groupPath string, properties map[string]any) error {
	return c.put("/groups/"+encodePath(groupPath)+"/integrations/jira", properties, nil)
}
