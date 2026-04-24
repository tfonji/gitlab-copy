package copy

import (
	"fmt"
	"sort"

	"gitlab-copy/internal"
	"gitlab-copy/internal/gitlab"
)

type GroupCopier struct {
	src     *gitlab.Client
	dst     *gitlab.Client
	domains []string
	dryRun  bool
}

func NewGroupCopier(src, dst *gitlab.Client, domains []string, dryRun bool) *GroupCopier {
	return &GroupCopier{src: src, dst: dst, domains: domains, dryRun: dryRun}
}

func (c *GroupCopier) Copy(groupPath string) []internal.DomainCopyResult {
	results := make([]internal.DomainCopyResult, 0, len(c.domains))
	for _, domain := range c.domains {
		results = append(results, c.copyDomain(groupPath, domain))
	}
	return results
}

func (c *GroupCopier) copyDomain(groupPath, domain string) internal.DomainCopyResult {
	switch domain {
	case "push_rules":
		return c.copyPushRules(groupPath)
	case "default_branch_name":
		return c.copyDefaultBranchName(groupPath)
	case "mr_settings":
		return c.copyMRSettings(groupPath)
	case "protected_environments":
		return c.copyProtectedEnvironments(groupPath)
	default:
		return internal.DomainCopyResult{
			Domain: domain,
			Error:  fmt.Errorf("unknown domain %q", domain),
		}
	}
}

// --- push_rules ---

func (c *GroupCopier) copyPushRules(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "push_rules"}

	src, err := c.src.GetGroupPushRules(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source push rules: %w", err)
		return result
	}
	// nil means 403 — not accessible on source
	if src == nil {
		result.Error = fmt.Errorf("source push rules not accessible (403)")
		return result
	}
	// Empty struct means 404 — no rules configured on source
	if src.IsEmpty() {
		result.Items = []internal.ItemResult{
			{Key: "push_rules", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	dst, err := c.dst.GetGroupPushRules(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest push rules: %w", err)
		return result
	}

	// Determine what action is needed
	var action internal.CopyAction
	dstExists := dst != nil && !dst.IsEmpty()

	if dstExists && src.Equal(dst) {
		result.Items = []internal.ItemResult{
			{Key: "push_rules", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if dstExists {
		action = internal.ActionUpdated
	} else {
		action = internal.ActionCreated
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "push_rules", Action: action, DryRun: true},
		}
		return result
	}

	req := gitlab.PushRuleRequestFrom(src)
	if dstExists {
		err = c.dst.UpdateGroupPushRules(groupPath, req)
	} else {
		err = c.dst.CreateGroupPushRules(groupPath, req)
	}
	if err != nil {
		result.Items = []internal.ItemResult{
			{Key: "push_rules", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "push_rules", Action: action},
	}
	return result
}

// --- default_branch_name ---

func (c *GroupCopier) copyDefaultBranchName(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "default_branch_name"}

	src, err := c.src.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source group: %w", err)
		return result
	}
	dst, err := c.dst.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest group: %w", err)
		return result
	}

	if src.DefaultBranchName == "" {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_name", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if src.DefaultBranchName == dst.DefaultBranchName {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_name", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_name", Action: internal.ActionUpdated, DryRun: true},
		}
		return result
	}

	if err := c.dst.UpdateGroup(groupPath, gitlab.GroupUpdateRequest{
		DefaultBranchName: gitlab.StrPtr(src.DefaultBranchName),
	}); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_name", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "default_branch_name", Action: internal.ActionUpdated},
	}
	return result
}

// --- mr_settings ---

func (c *GroupCopier) copyMRSettings(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "mr_settings"}

	src, err := c.src.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source group: %w", err)
		return result
	}
	dst, err := c.dst.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest group: %w", err)
		return result
	}

	// Compare all MR setting fields
	matches := src.OnlyAllowMergeIfPipelineSucceeds == dst.OnlyAllowMergeIfPipelineSucceeds &&
		src.OnlyAllowMergeIfAllDiscussionsAreResolved == dst.OnlyAllowMergeIfAllDiscussionsAreResolved &&
		ptrBoolEqual(src.PreventMergeWithoutJiraIssue, dst.PreventMergeWithoutJiraIssue)

	if matches {
		result.Items = []internal.ItemResult{
			{Key: "mr_settings", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "mr_settings", Action: internal.ActionUpdated, DryRun: true},
		}
		return result
	}

	if err := c.dst.UpdateGroup(groupPath, gitlab.GroupUpdateRequest{
		OnlyAllowMergeIfPipelineSucceeds:          gitlab.BoolPtr(src.OnlyAllowMergeIfPipelineSucceeds),
		OnlyAllowMergeIfAllDiscussionsAreResolved: gitlab.BoolPtr(src.OnlyAllowMergeIfAllDiscussionsAreResolved),
		PreventMergeWithoutJiraIssue:              src.PreventMergeWithoutJiraIssue,
	}); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "mr_settings", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "mr_settings", Action: internal.ActionUpdated},
	}
	return result
}

// --- protected_environments ---

func (c *GroupCopier) copyProtectedEnvironments(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "protected_environments"}

	srcEnvs, err := c.src.GetGroupProtectedEnvironments(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source protected environments: %w", err)
		return result
	}
	dstEnvs, err := c.dst.GetGroupProtectedEnvironments(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest protected environments: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstEnvs))
	for _, e := range dstEnvs {
		dstByName[e.Name] = true
	}

	// Sort source for deterministic output
	sort.Slice(srcEnvs, func(i, j int) bool {
		return srcEnvs[i].Name < srcEnvs[j].Name
	})

	for _, env := range srcEnvs {
		if dstByName[env.Name] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    env.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    env.Name,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.ProtectedEnvironmentRequestFrom(env)
		if err := c.dst.CreateGroupProtectedEnvironment(groupPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    env.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    env.Name,
				Action: internal.ActionCreated,
			})
		}
	}

	return result
}

// --- helpers ---

func ptrBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
