package gitlab

// AccessToken represents a group or project access token.
type AccessToken struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   string   `json:"expires_at"`
	AccessLevel int      `json:"access_level"`
	Active      bool     `json:"active"`
	Revoked     bool     `json:"revoked"`
}

// AccessTokenRequest is the write body for POST.
// AccessLevel uses GitLab role integers:
//
//	10 = Guest, 20 = Reporter, 30 = Developer, 40 = Maintainer, 50 = Owner
type AccessTokenRequest struct {
	Name        string   `json:"name"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
	AccessLevel int      `json:"access_level"`
}

// AccessTokenResponse is returned by POST — includes the secret token value
// which is only available at creation time.
type AccessTokenResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   string   `json:"expires_at"`
	AccessLevel int      `json:"access_level"`
	Token       string   `json:"token"`
}

// --- Group access tokens ---

func (c *Client) GetGroupAccessTokens(groupPath string) ([]AccessToken, error) {
	var tokens []AccessToken
	err := c.get("/groups/"+encodePath(groupPath)+"/access_tokens", nil, &tokens)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	// Only return active, non-revoked tokens
	var active []AccessToken
	for _, t := range tokens {
		if t.Active && !t.Revoked {
			active = append(active, t)
		}
	}
	return active, nil
}

func (c *Client) CreateGroupAccessToken(groupPath string, req AccessTokenRequest) (*AccessTokenResponse, error) {
	var resp AccessTokenResponse
	if err := c.post("/groups/"+encodePath(groupPath)+"/access_tokens", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Project access tokens ---

func (c *Client) GetProjectAccessTokens(projectPath string) ([]AccessToken, error) {
	var tokens []AccessToken
	err := c.get("/projects/"+encodePath(projectPath)+"/access_tokens", nil, &tokens)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	var active []AccessToken
	for _, t := range tokens {
		if t.Active && !t.Revoked {
			active = append(active, t)
		}
	}
	return active, nil
}

func (c *Client) CreateProjectAccessToken(projectPath string, req AccessTokenRequest) (*AccessTokenResponse, error) {
	var resp AccessTokenResponse
	if err := c.post("/projects/"+encodePath(projectPath)+"/access_tokens", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
