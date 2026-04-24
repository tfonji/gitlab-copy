package gitlab

// ProtectedTag represents a protected tag on a project.
type ProtectedTag struct {
	ID                 int                 `json:"id"`
	Name               string              `json:"name"`
	CreateAccessLevels []BranchAccessLevel `json:"create_access_levels"`
}

// ProtectedTagRequest is the write body for POST /projects/:id/protected_tags.
type ProtectedTagRequest struct {
	Name              string                     `json:"name"`
	CreateAccessLevel int                        `json:"create_access_level,omitempty"`
	AllowedToCreate   []BranchAccessLevelRequest `json:"allowed_to_create,omitempty"`
}

// ProtectedTagRequestFrom converts a ProtectedTag to a create request.
// Only role-based access levels are included.
func ProtectedTagRequestFrom(t ProtectedTag) ProtectedTagRequest {
	req := ProtectedTagRequest{Name: t.Name}
	for _, al := range t.CreateAccessLevels {
		if al.IsRoleBased() {
			req.AllowedToCreate = append(req.AllowedToCreate, BranchAccessLevelRequest{AccessLevel: al.AccessLevel})
		}
	}
	return req
}

// --- Read ---

func (c *Client) GetProjectProtectedTags(projectPath string) ([]ProtectedTag, error) {
	var tags []ProtectedTag
	err := c.get("/projects/"+encodePath(projectPath)+"/protected_tags", nil, &tags)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return tags, nil
}

// --- Write ---

func (c *Client) CreateProjectProtectedTag(projectPath string, req ProtectedTagRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/protected_tags", req, nil)
}

func (c *Client) DeleteProjectProtectedTag(projectPath string, tagName string) error {
	return c.delete("/projects/" + encodePath(projectPath) + "/protected_tags/" + encodePath(tagName))
}
