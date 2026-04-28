package gitlab

// DeployToken represents a deploy token on a group or project.
type DeployToken struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Username  string   `json:"username"`
	ExpiresAt string   `json:"expires_at"`
	Scopes    []string `json:"scopes"`
	Revoked   bool     `json:"revoked"`
	Expired   bool     `json:"expired"`
}

// DeployTokenRequest is the write body for POST.
type DeployTokenRequest struct {
	Name      string   `json:"name"`
	Username  string   `json:"username,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
	Scopes    []string `json:"scopes"`
}

// DeployTokenResponse is returned by POST — includes the secret value
// which is only available at creation time.
type DeployTokenResponse struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Token    string   `json:"token"`
	Scopes   []string `json:"scopes"`
}

// --- Group deploy tokens ---

func (c *Client) GetGroupDeployTokens(groupPath string) ([]DeployToken, error) {
	var tokens []DeployToken
	err := c.get("/groups/"+encodePath(groupPath)+"/deploy_tokens", nil, &tokens)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	// Only return active, non-expired tokens
	var active []DeployToken
	for _, t := range tokens {
		if !t.Revoked && !t.Expired {
			active = append(active, t)
		}
	}
	return active, nil
}

func (c *Client) CreateGroupDeployToken(groupPath string, req DeployTokenRequest) (*DeployTokenResponse, error) {
	var resp DeployTokenResponse
	if err := c.post("/groups/"+encodePath(groupPath)+"/deploy_tokens", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Project deploy tokens ---

func (c *Client) GetProjectDeployTokens(projectPath string) ([]DeployToken, error) {
	var tokens []DeployToken
	err := c.get("/projects/"+encodePath(projectPath)+"/deploy_tokens", nil, &tokens)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	var active []DeployToken
	for _, t := range tokens {
		if !t.Revoked && !t.Expired {
			active = append(active, t)
		}
	}
	return active, nil
}

func (c *Client) CreateProjectDeployToken(projectPath string, req DeployTokenRequest) (*DeployTokenResponse, error) {
	var resp DeployTokenResponse
	if err := c.post("/projects/"+encodePath(projectPath)+"/deploy_tokens", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
