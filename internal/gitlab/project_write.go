package gitlab

// ProjectUpdateRequest is the write body for PUT /projects/:id.
// Uses omitempty on pointer fields so each domain only touches its own fields.
type ProjectUpdateRequest struct {
	Topics []string `json:"topics,omitempty"`
}

// UpdateProject issues a PUT /projects/:id with the provided fields.
func (c *Client) UpdateProject(projectPath string, req ProjectUpdateRequest) error {
	return c.put("/projects/"+encodePath(projectPath), req, nil)
}
