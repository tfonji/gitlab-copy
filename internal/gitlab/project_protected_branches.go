package gitlab

// ProtectedBranch represents a protected branch on a project.
type ProtectedBranch struct {
	ID                        int                 `json:"id"`
	Name                      string              `json:"name"`
	PushAccessLevels          []BranchAccessLevel `json:"push_access_levels"`
	MergeAccessLevels         []BranchAccessLevel `json:"merge_access_levels"`
	UnprotectAccessLevels     []BranchAccessLevel `json:"unprotect_access_levels"`
	AllowForcePush            bool                `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool                `json:"code_owner_approval_required"`
}

// BranchAccessLevel represents an access level entry on a protected branch.
type BranchAccessLevel struct {
	ID                     int    `json:"id"`
	AccessLevel            int    `json:"access_level"`
	AccessLevelDescription string `json:"access_level_description"`
	UserID                 *int   `json:"user_id"`
	GroupID                *int   `json:"group_id"`
}

// IsRoleBased returns true when the access level is role-based (not user/group specific).
// Only role-based entries transfer cleanly across instances.
func (a BranchAccessLevel) IsRoleBased() bool {
	return a.UserID == nil && a.GroupID == nil
}

// ProtectedBranchRequest is the write body for POST /projects/:id/protected_branches.
type ProtectedBranchRequest struct {
	Name                      string                     `json:"name"`
	PushAccessLevel           int                        `json:"push_access_level,omitempty"`
	MergeAccessLevel          int                        `json:"merge_access_level,omitempty"`
	UnprotectAccessLevel      int                        `json:"unprotect_access_level,omitempty"`
	AllowForcePush            bool                       `json:"allow_force_push"`
	CodeOwnerApprovalRequired bool                       `json:"code_owner_approval_required"`
	AllowedToPush             []BranchAccessLevelRequest `json:"allowed_to_push,omitempty"`
	AllowedToMerge            []BranchAccessLevelRequest `json:"allowed_to_merge,omitempty"`
	AllowedToUnprotect        []BranchAccessLevelRequest `json:"allowed_to_unprotect,omitempty"`
}

type BranchAccessLevelRequest struct {
	AccessLevel int `json:"access_level"`
}

// ProtectedBranchRequestFrom converts a ProtectedBranch to a create request.
// Only role-based access levels are included — user/group specific ones are skipped.
func ProtectedBranchRequestFrom(b ProtectedBranch) ProtectedBranchRequest {
	req := ProtectedBranchRequest{
		Name:                      b.Name,
		AllowForcePush:            b.AllowForcePush,
		CodeOwnerApprovalRequired: b.CodeOwnerApprovalRequired,
	}
	for _, al := range b.PushAccessLevels {
		if al.IsRoleBased() {
			req.AllowedToPush = append(req.AllowedToPush, BranchAccessLevelRequest{AccessLevel: al.AccessLevel})
		}
	}
	for _, al := range b.MergeAccessLevels {
		if al.IsRoleBased() {
			req.AllowedToMerge = append(req.AllowedToMerge, BranchAccessLevelRequest{AccessLevel: al.AccessLevel})
		}
	}
	for _, al := range b.UnprotectAccessLevels {
		if al.IsRoleBased() {
			req.AllowedToUnprotect = append(req.AllowedToUnprotect, BranchAccessLevelRequest{AccessLevel: al.AccessLevel})
		}
	}
	return req
}

// --- Read ---

func (c *Client) GetProjectProtectedBranches(projectPath string) ([]ProtectedBranch, error) {
	var branches []ProtectedBranch
	err := c.get("/projects/"+encodePath(projectPath)+"/protected_branches", nil, &branches)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return branches, nil
}

// --- Write ---

func (c *Client) CreateProjectProtectedBranch(projectPath string, req ProtectedBranchRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/protected_branches", req, nil)
}

func (c *Client) DeleteProjectProtectedBranch(projectPath string, branchName string) error {
	return c.delete("/projects/" + encodePath(projectPath) + "/protected_branches/" + encodePath(branchName))
}
