package gitlab

import (
	"fmt"
	"net/url"
)

type Project struct {
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Description       string `json:"description"`
	Visibility        string `json:"visibility"`

	DefaultBranch               string `json:"default_branch"`
	EmptyRepo                   bool   `json:"empty_repo"`
	LFSEnabled                  bool   `json:"lfs_enabled"`
	RequestAccessEnabled        bool   `json:"request_access_enabled"`
	AllowMergeOnSkippedPipeline *bool  `json:"allow_merge_on_skipped_pipeline"`

	MergeMethod                               string `json:"merge_method"`
	SquashOption                              string `json:"squash_option"`
	OnlyAllowMergeIfPipelineSucceeds          bool   `json:"only_allow_merge_if_pipeline_succeeds"`
	OnlyAllowMergeIfAllDiscussionsAreResolved bool   `json:"only_allow_merge_if_all_discussions_are_resolved"`
	RemoveSourceBranchAfterMerge              bool   `json:"remove_source_branch_after_merge"`
	PrintingMergeRequestLinkEnabled           bool   `json:"printing_merge_request_link_enabled"`
	MergeRequestsEnabled                      bool   `json:"merge_requests_enabled"`

	IssuesEnabled                    bool   `json:"issues_enabled"`
	WikiEnabled                      bool   `json:"wiki_enabled"`
	SnippetsEnabled                  bool   `json:"snippets_enabled"`
	JobsEnabled                      bool   `json:"jobs_enabled"`
	PackagesEnabled                  *bool  `json:"packages_enabled"`
	PagesAccessLevel                 string `json:"pages_access_level"`
	OperationsAccessLevel            string `json:"operations_access_level"`
	SecurityAndComplianceAccessLevel string `json:"security_and_compliance_access_level"`

	AutoCancelPendingPipelines string `json:"auto_cancel_pending_pipelines"`
	CIConfigPath               string `json:"ci_config_path"`
	CIDefaultGitDepth          int    `json:"ci_default_git_depth"`
	CIForwardDeploymentEnabled *bool  `json:"ci_forward_deployment_enabled"`
	CISeperateCache            *bool  `json:"ci_separated_caches"`
	SharedRunnersEnabled       bool   `json:"shared_runners_enabled"`
	GroupRunnersEnabled        bool   `json:"group_runners_enabled"`
	AutoDevopsEnabled          bool   `json:"auto_devops_enabled"`
	AutoDevopsDeployStrategy   string `json:"auto_devops_deploy_strategy"`
	BuildTimeout               int    `json:"build_timeout"`
	PublicBuilds               bool   `json:"public_builds"`

	ContainerRegistryEnabled     bool                       `json:"container_registry_enabled"`
	ContainerRegistryAccessLevel string                     `json:"container_registry_access_level"`
	ContainerExpirationPolicy    *ContainerExpirationPolicy `json:"container_expiration_policy"`

	Topics                      []string `json:"topics"`
	OnlyMirrorProtectedBranches *bool    `json:"only_mirror_protected_branches"`
	ApprovalsBeforeMerge        int      `json:"approvals_before_merge"`
}

type ContainerExpirationPolicy struct {
	Cadence         string `json:"cadence"`
	Enabled         bool   `json:"enabled"`
	KeepN           int    `json:"keep_n"`
	OlderThan       string `json:"older_than"`
	NameRegex       string `json:"name_regex"`
	NameRegexDelete string `json:"name_regex_delete"`
	NameRegexKeep   string `json:"name_regex_keep"`
}

type ProjectListItem struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	Archived          bool   `json:"archived"`
}

func (c *Client) GetProject(projectPath string) (*Project, error) {
	var p Project
	params := url.Values{}
	err := c.get("/projects/"+encodePath(projectPath), params, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) ListGroupProjects(groupPath string, includeSubgroups, includeArchived bool) ([]ProjectListItem, error) {
	var all []ProjectListItem
	page := 1

	for {
		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		if includeSubgroups {
			params.Set("include_subgroups", "true")
		} else {
			params.Set("include_subgroups", "false")
		}
		if !includeArchived {
			params.Set("archived", "false")
		}
		params.Set("with_shared", "false")

		var batch []ProjectListItem
		err := c.get("/groups/"+encodePath(groupPath)+"/projects", params, &batch)
		if err != nil {
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
