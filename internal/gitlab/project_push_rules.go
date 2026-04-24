package gitlab

// GetProjectPushRules fetches push rules for a project.
// Returns empty struct (not nil) on 404 — meaning no rules configured.
// Returns nil on 403 — meaning not accessible.
func (c *Client) GetProjectPushRules(projectPath string) (*PushRule, error) {
	var pr PushRule
	err := c.get("/projects/"+encodePath(projectPath)+"/push_rule", nil, &pr)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return &PushRule{}, nil
		}
		if apiErr, ok := err.(*APIError); ok && apiErr.IsForbidden() {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

// CreateProjectPushRules creates push rules on the dest project via POST.
func (c *Client) CreateProjectPushRules(projectPath string, req PushRuleRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/push_rule", req, nil)
}

// UpdateProjectPushRules updates existing push rules on the dest project via PUT.
func (c *Client) UpdateProjectPushRules(projectPath string, req PushRuleRequest) error {
	return c.put("/projects/"+encodePath(projectPath)+"/push_rule", req, nil)
}
