package gitlab

// MergeRequestApprovalSettings represents the group-level MR approval settings.
// Each field is wrapped in a value struct as that is how the API returns them.
type MergeRequestApprovalSettings struct {
	AllowAuthorApproval          ApprovalSettingValue `json:"allow_author_approval"`
	AllowCommitterApproval       ApprovalSettingValue `json:"allow_committer_approval"`
	AllowOverridesToApproverList ApprovalSettingValue `json:"allow_overrides_to_approver_list_per_merge_request"`
	RequirePasswordToApprove     ApprovalSettingValue `json:"require_password_to_approve"`
	RetainApprovalsOnPush        ApprovalSettingValue `json:"retain_approvals_on_push"`
	SelectiveCodeOwnerRemovals   ApprovalSettingValue `json:"selective_code_owner_removals"`
}

type ApprovalSettingValue struct {
	Value bool `json:"value"`
}

// MergeRequestApprovalSettingsRequest is the write body for PUT.
// The API takes the boolean values directly (not wrapped in value structs).
type MergeRequestApprovalSettingsRequest struct {
	AllowAuthorApproval          bool `json:"allow_author_approval"`
	AllowCommitterApproval       bool `json:"allow_committer_approval"`
	AllowOverridesToApproverList bool `json:"allow_overrides_to_approver_list_per_merge_request"`
	RequirePasswordToApprove     bool `json:"require_password_to_approve"`
	RetainApprovalsOnPush        bool `json:"retain_approvals_on_push"`
	SelectiveCodeOwnerRemovals   bool `json:"selective_code_owner_removals"`
}

// --- Read ---

func (c *Client) GetMRApprovalSettings(groupPath string) (*MergeRequestApprovalSettings, error) {
	var s MergeRequestApprovalSettings
	err := c.get("/groups/"+encodePath(groupPath)+"/merge_request_approval_setting", nil, &s)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// --- Write ---

func (c *Client) SetMRApprovalSettings(groupPath string, req MergeRequestApprovalSettingsRequest) error {
	return c.put("/groups/"+encodePath(groupPath)+"/merge_request_approval_setting", req, nil)
}
