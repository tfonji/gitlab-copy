package gitlab

import "fmt"

// ApprovalRule represents a group-level MR approval rule.
type ApprovalRule struct {
	ID                   int              `json:"id"`
	Name                 string           `json:"name"`
	RuleType             string           `json:"rule_type"`
	ApprovalsRequired    int              `json:"approvals_required"`
	EligibleApprovers    []map[string]any `json:"eligible_approvers"`
	ContainsHiddenGroups bool             `json:"contains_hidden_groups"`
}

// ApprovalRuleRequest is the write body for POST/PUT.
// UserIDs and GroupIDs intentionally omitted from the default copy path —
// those IDs are instance-specific and won't transfer. See RuleType handling
// in the copier for how each rule type is treated.
type ApprovalRuleRequest struct {
	Name              string `json:"name"`
	ApprovalsRequired int    `json:"approvals_required"`
}

// --- Read ---

func (c *Client) GetGroupApprovalRules(groupPath string) ([]ApprovalRule, error) {
	var rules []ApprovalRule
	err := c.get("/groups/"+encodePath(groupPath)+"/approval_rules", nil, &rules)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return rules, nil
}

// --- Write ---

// CreateGroupApprovalRule creates a new approval rule on the dest group.
func (c *Client) CreateGroupApprovalRule(groupPath string, req ApprovalRuleRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/approval_rules", req, nil)
}

// UpdateGroupApprovalRule updates an existing approval rule by ID on the dest group.
func (c *Client) UpdateGroupApprovalRule(groupPath string, ruleID int, req ApprovalRuleRequest) error {
	return c.put("/groups/"+encodePath(groupPath)+"/approval_rules/"+itoa(ruleID), req, nil)
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
