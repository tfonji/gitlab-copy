package gitlab

import (
	"fmt"
	"net/url"
)

type Group struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Description string  `json:"description"`
	Visibility  string  `json:"visibility"`
	AvatarURL   *string `json:"avatar_url"`
	FullPath    string  `json:"full_path"`
	FullName    string  `json:"full_name"`
	WebURL      string  `json:"web_url"`

	LFSEnabled                            bool    `json:"lfs_enabled"`
	RequestAccessEnabled                  bool    `json:"request_access_enabled"`
	ShareWithGroupLock                    bool    `json:"share_with_group_lock"`
	PreventSharingGroupsOutsideHierarchy  bool    `json:"prevent_sharing_groups_outside_hierarchy"`
	RequireTwoFactorAuthentication        bool    `json:"require_two_factor_authentication"`
	TwoFactorGracePeriod                  int     `json:"two_factor_grace_period"`
	ProjectCreationLevel                  string  `json:"project_creation_level"`
	SubgroupCreationLevel                 string  `json:"subgroup_creation_level"`
	EmailsEnabled                         bool    `json:"emails_enabled"`
	MentionsDisabled                      *bool   `json:"mentions_disabled"`
	MembershipLock                        bool    `json:"membership_lock"`
	PreventForkingOutsideGroup            *bool   `json:"prevent_forking_outside_group"`
	ServiceAccessTokensExpirationEnforced *bool   `json:"service_access_tokens_expiration_enforced"`
	IPRestrictionRanges                   *string `json:"ip_restriction_ranges"`

	UniqueProjectDownloadLimit                  *int     `json:"unique_project_download_limit"`
	UniqueProjectDownloadLimitIntervalInSeconds *int     `json:"unique_project_download_limit_interval_in_seconds"`
	UniqueProjectDownloadLimitAllowlist         []string `json:"unique_project_download_limit_allowlist"`

	EnabledGitAccessProtocol        string                           `json:"enabled_git_access_protocol"`
	DefaultBranchProtection         int                              `json:"default_branch_protection"`
	DefaultBranchProtectionDefaults *DefaultBranchProtectionDefaults `json:"default_branch_protection_defaults"`

	AutoDevopsEnabled    *bool  `json:"auto_devops_enabled"`
	SharedRunnersSetting string `json:"shared_runners_setting"`

	CRMEnabled  bool  `json:"crm_enabled"`
	WikiEnabled *bool `json:"wiki_enabled"`

	DefaultBranchName string `json:"default_branch_name"`

	OnlyAllowMergeIfPipelineSucceeds          bool  `json:"only_allow_merge_if_pipeline_succeeds"`
	OnlyAllowMergeIfAllDiscussionsAreResolved bool  `json:"only_allow_merge_if_all_discussions_are_resolved"`
	PreventMergeWithoutJiraIssue              *bool `json:"prevent_merge_without_jira_issue"`
}

type DefaultBranchProtectionDefaults struct {
	AllowedToPush           []map[string]any `json:"allowed_to_push"`
	AllowForcePush          bool             `json:"allow_force_push"`
	AllowedToMerge          []map[string]any `json:"allowed_to_merge"`
	DeveloperCanInitialPush bool             `json:"developer_can_initial_push"`
}

func (c *Client) GetGroup(groupPath string) (*Group, error) {
	var g Group
	params := url.Values{}
	params.Set("with_projects", "false")
	params.Set("with_statistics", "true")
	err := c.get("/groups/"+encodePath(groupPath), params, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

type GroupListItem struct {
	ID       int    `json:"id"`
	FullPath string `json:"full_path"`
	ParentID *int   `json:"parent_id"`
}

func (c *Client) ListSubgroups(groupPath string) ([]GroupListItem, error) {
	var all []GroupListItem
	page := 1

	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))

		var batch []GroupListItem
		err := c.get("/groups/"+encodePath(groupPath)+"/descendant_groups", params, &batch)
		if err != nil {
			if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
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
