package gitlab

import (
	"fmt"
	"net/url"
)

// GroupHook represents a group-level webhook.
type GroupHook struct {
	ID                       int    `json:"id"`
	URL                      string `json:"url"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	PushEvents               bool   `json:"push_events"`
	TagPushEvents            bool   `json:"tag_push_events"`
	MergeRequestsEvents      bool   `json:"merge_requests_events"`
	IssuesEvents             bool   `json:"issues_events"`
	ConfidentialIssuesEvents bool   `json:"confidential_issues_events"`
	NoteEvents               bool   `json:"note_events"`
	ConfidentialNoteEvents   bool   `json:"confidential_note_events"`
	PipelineEvents           bool   `json:"pipeline_events"`
	WikiPageEvents           bool   `json:"wiki_page_events"`
	JobEvents                bool   `json:"job_events"`
	DeploymentEvents         bool   `json:"deployment_events"`
	ReleasesEvents           bool   `json:"releases_events"`
	SubGroupEvents           bool   `json:"subgroup_events"`
	MemberEvents             bool   `json:"member_events"`
	FeatureFlagEvents        bool   `json:"feature_flag_events"`
	EnableSSLVerification    bool   `json:"enable_ssl_verification"`
}

// GroupHookRequest is the write body for POST /groups/:id/hooks.
// Token is intentionally omitted — it is write-only and cannot be read from source.
type GroupHookRequest struct {
	URL                      string `json:"url"`
	Name                     string `json:"name,omitempty"`
	Description              string `json:"description,omitempty"`
	PushEvents               bool   `json:"push_events"`
	TagPushEvents            bool   `json:"tag_push_events"`
	MergeRequestsEvents      bool   `json:"merge_requests_events"`
	IssuesEvents             bool   `json:"issues_events"`
	ConfidentialIssuesEvents bool   `json:"confidential_issues_events"`
	NoteEvents               bool   `json:"note_events"`
	ConfidentialNoteEvents   bool   `json:"confidential_note_events"`
	PipelineEvents           bool   `json:"pipeline_events"`
	WikiPageEvents           bool   `json:"wiki_page_events"`
	JobEvents                bool   `json:"job_events"`
	DeploymentEvents         bool   `json:"deployment_events"`
	ReleasesEvents           bool   `json:"releases_events"`
	SubGroupEvents           bool   `json:"subgroup_events"`
	MemberEvents             bool   `json:"member_events"`
	FeatureFlagEvents        bool   `json:"feature_flag_events"`
	EnableSSLVerification    bool   `json:"enable_ssl_verification"`
}

func (c *Client) GetGroupHooks(groupPath string) ([]GroupHook, error) {
	var all []GroupHook
	page := 1
	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		var batch []GroupHook
		if err := c.get("/groups/"+encodePath(groupPath)+"/hooks", params, &batch); err != nil {
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

func (c *Client) CreateGroupHook(groupPath string, req GroupHookRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/hooks", req, nil)
}
