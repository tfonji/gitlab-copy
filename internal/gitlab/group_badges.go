package gitlab

// GetGroupBadges fetches badges defined directly on a group (not inherited).
func (c *Client) GetGroupBadges(groupPath string) ([]Badge, error) {
	var badges []Badge
	err := c.get("/groups/"+encodePath(groupPath)+"/badges", nil, &badges)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	// Filter to group-owned badges only (exclude inherited ones)
	var own []Badge
	for _, b := range badges {
		if b.Kind == "group" {
			own = append(own, b)
		}
	}
	return own, nil
}

func (c *Client) CreateGroupBadge(groupPath string, req BadgeRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/badges", req, nil)
}

func (c *Client) DeleteGroupBadge(groupPath string, badgeID int) error {
	return c.delete("/groups/" + encodePath(groupPath) + "/badges/" + itoa(badgeID))
}
