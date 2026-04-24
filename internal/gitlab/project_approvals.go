package gitlab

import "fmt"

// ProjectApprovalSettings represents project-level MR approval settings.
type ProjectApprovalSettings struct {
	ApprovalsBeforeMerge                      int  `json:"approvals_before_merge"`
	ResetApprovalsOnPush                      bool `json:"reset_approvals_on_push"`
	SelectiveCodeOwnerRemovals                bool `json:"selective_code_owner_removals"`
	DisableOverridingApproversPerMergeRequest bool `json:"disable_overriding_approvers_per_merge_request"`
	MergeRequestsAuthorApproval               bool `json:"merge_requests_author_approval"`
	MergeRequestsDisableCommittersApproval    bool `json:"merge_requests_disable_committers_approval"`
	RequirePasswordToApprove                  bool `json:"require_password_to_approve"`
}

// ProjectApprovalRule represents a project-level MR approval rule.
type ProjectApprovalRule struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	RuleType          string `json:"rule_type"`
	ApprovalsRequired int    `json:"approvals_required"`
}

// ProjectApprovalRuleRequest is the write body for POST/PUT.
type ProjectApprovalRuleRequest struct {
	Name              string `json:"name"`
	ApprovalsRequired int    `json:"approvals_required"`
}

// --- Read ---

func (c *Client) GetProjectMRApprovals(projectPath string) (*ProjectApprovalSettings, error) {
	var s ProjectApprovalSettings
	err := c.get("/projects/"+encodePath(projectPath)+"/approvals", nil, &s)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (c *Client) GetProjectApprovalRules(projectPath string) ([]ProjectApprovalRule, error) {
	var rules []ProjectApprovalRule
	err := c.get("/projects/"+encodePath(projectPath)+"/approval_rules", nil, &rules)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return rules, nil
}

// --- Write ---

// SetProjectMRApprovals sets project-level MR approval settings via POST.
func (c *Client) SetProjectMRApprovals(projectPath string, s *ProjectApprovalSettings) error {
	return c.post("/projects/"+encodePath(projectPath)+"/approvals", s, nil)
}

// CreateProjectApprovalRule creates a new approval rule on the dest project.
func (c *Client) CreateProjectApprovalRule(projectPath string, req ProjectApprovalRuleRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/approval_rules", req, nil)
}

// UpdateProjectApprovalRule updates an existing approval rule by ID.
func (c *Client) UpdateProjectApprovalRule(projectPath string, ruleID int, req ProjectApprovalRuleRequest) error {
	return c.put("/projects/"+encodePath(projectPath)+"/approval_rules/"+fmt.Sprintf("%d", ruleID), req, nil)
}
