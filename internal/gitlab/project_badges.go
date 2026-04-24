package gitlab

// Badge represents a project badge.
type Badge struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LinkURL  string `json:"link_url"`
	ImageURL string `json:"image_url"`
	Kind     string `json:"kind"`
}

// BadgeRequest is the write body for POST /projects/:id/badges.
type BadgeRequest struct {
	Name     string `json:"name"`
	LinkURL  string `json:"link_url"`
	ImageURL string `json:"image_url"`
}

// --- Read ---

func (c *Client) GetProjectBadges(projectPath string) ([]Badge, error) {
	var badges []Badge
	err := c.get("/projects/"+encodePath(projectPath)+"/badges", nil, &badges)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	// Only return project-level badges (not group-inherited ones)
	var project []Badge
	for _, b := range badges {
		if b.Kind == "project" {
			project = append(project, b)
		}
	}
	return project, nil
}

// --- Write ---

func (c *Client) CreateProjectBadge(projectPath string, req BadgeRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/badges", req, nil)
}

func (c *Client) DeleteProjectBadge(projectPath string, badgeID int) error {
	return c.delete("/projects/" + encodePath(projectPath) + "/badges/" + itoa(badgeID))
}
