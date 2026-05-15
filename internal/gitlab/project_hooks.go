package gitlab

import (
	"fmt"
	"net/url"
)

// ProjectHook represents a project-level webhook.
type ProjectHook struct {
	ID                        int    `json:"id"`
	URL                       string `json:"url"`
	Name                      string `json:"name"`
	Description               string `json:"description"`
	PushEvents                bool   `json:"push_events"`
	TagPushEvents             bool   `json:"tag_push_events"`
	MergeRequestsEvents       bool   `json:"merge_requests_events"`
	IssuesEvents              bool   `json:"issues_events"`
	ConfidentialIssuesEvents  bool   `json:"confidential_issues_events"`
	NoteEvents                bool   `json:"note_events"`
	ConfidentialNoteEvents    bool   `json:"confidential_note_events"`
	PipelineEvents            bool   `json:"pipeline_events"`
	WikiPageEvents            bool   `json:"wiki_page_events"`
	JobEvents                 bool   `json:"job_events"`
	DeploymentEvents          bool   `json:"deployment_events"`
	ReleasesEvents            bool   `json:"releases_events"`
	MemberEvents              bool   `json:"member_events"`
	FeatureFlagEvents         bool   `json:"feature_flag_events"`
	EnableSSLVerification     bool   `json:"enable_ssl_verification"`
	ResourceAccessTokenEvents bool   `json:"resource_access_token_events"`
	EmojiEvents               bool   `json:"emoji_events"`
	MilestoneEvents           bool   `json:"milestone_events"`
	VulnerabilityEvents       bool   `json:"vulnerability_events"`
	PushEventsBranchFilter    string `json:"push_events_branch_filter"`
	BranchFilterStrategy      string `json:"branch_filter_strategy"`
}

// ProjectHookRequest is the write body for POST/PUT /projects/:id/hooks.
type ProjectHookRequest struct {
	URL                       string `json:"url"`
	Name                      string `json:"name,omitempty"`
	Description               string `json:"description,omitempty"`
	PushEvents                bool   `json:"push_events"`
	PushEventsBranchFilter    string `json:"push_events_branch_filter,omitempty"`
	BranchFilterStrategy      string `json:"branch_filter_strategy,omitempty"`
	TagPushEvents             bool   `json:"tag_push_events"`
	MergeRequestsEvents       bool   `json:"merge_requests_events"`
	IssuesEvents              bool   `json:"issues_events"`
	ConfidentialIssuesEvents  bool   `json:"confidential_issues_events"`
	NoteEvents                bool   `json:"note_events"`
	ConfidentialNoteEvents    bool   `json:"confidential_note_events"`
	PipelineEvents            bool   `json:"pipeline_events"`
	WikiPageEvents            bool   `json:"wiki_page_events"`
	JobEvents                 bool   `json:"job_events"`
	DeploymentEvents          bool   `json:"deployment_events"`
	ReleasesEvents            bool   `json:"releases_events"`
	MemberEvents              bool   `json:"member_events"`
	FeatureFlagEvents         bool   `json:"feature_flag_events"`
	EnableSSLVerification     bool   `json:"enable_ssl_verification"`
	ResourceAccessTokenEvents bool   `json:"resource_access_token_events"`
	EmojiEvents               bool   `json:"emoji_events"`
	MilestoneEvents           bool   `json:"milestone_events"`
	VulnerabilityEvents       bool   `json:"vulnerability_events"`
}

func (c *Client) GetProjectHooks(projectPath string) ([]ProjectHook, error) {
	var all []ProjectHook
	page := 1
	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		var batch []ProjectHook
		if err := c.get("/projects/"+encodePath(projectPath)+"/hooks", params, &batch); err != nil {
			if apiErr, ok := err.(*APIError); ok && (apiErr.IsNotFound() || apiErr.IsForbidden()) {
				return nil, nil
			}
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) CreateProjectHook(projectPath string, req ProjectHookRequest) error {
	return c.post("/projects/"+encodePath(projectPath)+"/hooks", req, nil)
}

func (c *Client) UpdateProjectHook(projectPath string, hookID int, req ProjectHookRequest) error {
	return c.put(fmt.Sprintf("/projects/%s/hooks/%d", encodePath(projectPath), hookID), req, nil)
}
