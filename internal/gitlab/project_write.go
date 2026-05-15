package gitlab

// ProjectUpdateRequest is the write body for PUT /projects/:id.
// Uses omitempty on pointer fields so each domain only touches its own fields.
type ProjectUpdateRequest struct {
	Topics                             []string `json:"topics,omitempty"`
	Description                        *string  `json:"description,omitempty"`
	DefaultBranch                      *string  `json:"default_branch,omitempty"`
	AutoCancelPendingPipelines         *string  `json:"auto_cancel_pending_pipelines,omitempty"`
	CIForwardDeploymentEnabled         *bool    `json:"ci_forward_deployment_enabled,omitempty"`
	CISeperateCache                    *bool    `json:"ci_separated_caches,omitempty"`
	PrintingMergeRequestLinkEnabled    *bool    `json:"printing_merge_request_link_enabled,omitempty"`
	RemoveSourceBranchAfterMerge       *bool    `json:"remove_source_branch_after_merge,omitempty"`
	AllowMergeOnSkippedPipeline        *bool    `json:"allow_merge_on_skipped_pipeline,omitempty"`
	CiPushRepositoryForJobTokenAllowed *bool    `json:"ci_push_repository_for_job_token_allowed,omitempty"`
}

// UpdateProject issues a PUT /projects/:id with the provided fields.
func (c *Client) UpdateProject(projectPath string, req ProjectUpdateRequest) error {
	return c.put("/projects/"+encodePath(projectPath), req, nil)
}
