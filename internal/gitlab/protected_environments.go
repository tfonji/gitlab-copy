package gitlab

// ProtectedEnvironment represents a protected environment on a group or project.
type ProtectedEnvironment struct {
	Name                  string                    `json:"name"`
	DeployAccessLevels    []EnvironmentAccessLevel  `json:"deploy_access_levels"`
	RequiredApprovalCount int                       `json:"required_approval_count"`
	ApprovalRules         []EnvironmentApprovalRule `json:"approval_rules"`
}

type EnvironmentAccessLevel struct {
	AccessLevel            int    `json:"access_level"`
	AccessLevelDescription string `json:"access_level_description"`
	UserID                 *int   `json:"user_id"`
	GroupID                *int   `json:"group_id"`
}

type EnvironmentApprovalRule struct {
	UserID                 *int   `json:"user_id"`
	GroupID                *int   `json:"group_id"`
	AccessLevel            int    `json:"access_level"`
	AccessLevelDescription string `json:"access_level_description"`
	RequiredApprovals      int    `json:"required_approvals"`
}

// ProtectedEnvironmentRequest is the write body for POST.
// Note: UserID and GroupID in access levels reference instance-specific IDs
// and will NOT transfer correctly across instances. Only role-based
// access levels (UserID == nil && GroupID == nil) copy cleanly.
type ProtectedEnvironmentRequest struct {
	Name                  string                `json:"name"`
	DeployAccessLevels    []AccessLevelRequest  `json:"deploy_access_levels"`
	RequiredApprovalCount int                   `json:"required_approval_count"`
	ApprovalRules         []ApprovalRuleRequest `json:"approval_rules,omitempty"`
}

type AccessLevelRequest struct {
	AccessLevel int  `json:"access_level,omitempty"`
	UserID      *int `json:"user_id,omitempty"`
	GroupID     *int `json:"group_id,omitempty"`
}

type ApprovalRuleRequest struct {
	AccessLevel       int  `json:"access_level,omitempty"`
	UserID            *int `json:"user_id,omitempty"`
	GroupID           *int `json:"group_id,omitempty"`
	RequiredApprovals int  `json:"required_approvals"`
}

// ProtectedEnvironmentRequestFrom converts a ProtectedEnvironment to a create request.
func ProtectedEnvironmentRequestFrom(pe ProtectedEnvironment) ProtectedEnvironmentRequest {
	req := ProtectedEnvironmentRequest{
		Name:                  pe.Name,
		RequiredApprovalCount: pe.RequiredApprovalCount,
	}
	for _, al := range pe.DeployAccessLevels {
		req.DeployAccessLevels = append(req.DeployAccessLevels, AccessLevelRequest{
			AccessLevel: al.AccessLevel,
			UserID:      al.UserID,
			GroupID:     al.GroupID,
		})
	}
	for _, ar := range pe.ApprovalRules {
		req.ApprovalRules = append(req.ApprovalRules, ApprovalRuleRequest{
			AccessLevel:       ar.AccessLevel,
			UserID:            ar.UserID,
			GroupID:           ar.GroupID,
			RequiredApprovals: ar.RequiredApprovals,
		})
	}
	return req
}

// --- Read ---

func (c *Client) GetGroupProtectedEnvironments(groupPath string) ([]ProtectedEnvironment, error) {
	var envs []ProtectedEnvironment
	err := c.get("/groups/"+encodePath(groupPath)+"/protected_environments", nil, &envs)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return envs, nil
}

func (c *Client) GetProjectProtectedEnvironments(projectPath string) ([]ProtectedEnvironment, error) {
	var envs []ProtectedEnvironment
	err := c.get("/projects/"+encodePath(projectPath)+"/protected_environments", nil, &envs)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return envs, nil
}

// --- Write ---

// CreateGroupProtectedEnvironment creates a protected environment on a group.
func (c *Client) CreateGroupProtectedEnvironment(groupPath string, req ProtectedEnvironmentRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/protected_environments", req, nil)
}

// CreateProjectProtectedEnvironment creates a protected environment on a project.
func (c *Client) CreateProjectProtectedEnvironment(projectPath string, req ProtectedEnvironmentRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/protected_environments", req, nil)
}
