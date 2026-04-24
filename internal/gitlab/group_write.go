package gitlab

// GroupUpdateRequest is the write body for PUT /groups/:id.
// Uses pointer fields with omitempty so only explicitly-set fields are sent —
// this lets default_branch_name and mr_settings each issue a targeted PUT
// without clobbering fields they don't own.
type GroupUpdateRequest struct {
	DefaultBranchName                         *string `json:"default_branch_name,omitempty"`
	OnlyAllowMergeIfPipelineSucceeds          *bool   `json:"only_allow_merge_if_pipeline_succeeds,omitempty"`
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool   `json:"only_allow_merge_if_all_discussions_are_resolved,omitempty"`
	PreventMergeWithoutJiraIssue              *bool   `json:"prevent_merge_without_jira_issue,omitempty"`
}

// UpdateGroup issues a PUT /groups/:id with the provided fields.
func (c *Client) UpdateGroup(groupPath string, req GroupUpdateRequest) error {
	return c.put("/groups/"+encodePath(groupPath), req, nil)
}

// Helper constructors so callers don't need to take addresses of literals.
func StrPtr(s string) *string { return &s }
func BoolPtr(b bool) *bool    { return &b }
