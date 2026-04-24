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
	case "approval_rules":
		return c.copyApprovalRules(groupPath)
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

// --- approval_rules ---

// Rule types:
//
//	any_approver — anyone can approve, no specific users/groups → copies cleanly
//	regular      — specific users/groups required → copies name+count only, approvers need manual assignment
//	code_owner   — auto-managed by CODEOWNERS → skipped
func (c *GroupCopier) copyApprovalRules(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "approval_rules"}

	srcRules, err := c.src.GetGroupApprovalRules(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source approval rules: %w", err)
		return result
	}
	dstRules, err := c.dst.GetGroupApprovalRules(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest approval rules: %w", err)
		return result
	}

	dstByName := make(map[string]gitlab.ApprovalRule, len(dstRules))
	for _, r := range dstRules {
		dstByName[r.Name] = r
	}

	sort.Slice(srcRules, func(i, j int) bool {
		return srcRules[i].Name < srcRules[j].Name
	})

	for _, src := range srcRules {
		// code_owner rules are auto-managed by CODEOWNERS — never copy
		if src.RuleType == "code_owner" {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		req := gitlab.ApprovalRuleRequest{
			Name:              src.Name,
			ApprovalsRequired: src.ApprovalsRequired,
		}

		dst, exists := dstByName[src.Name]

		if exists && dst.ApprovalsRequired == src.ApprovalsRequired {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		action := internal.ActionCreated
		if exists {
			action = internal.ActionUpdated
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: action,
				DryRun: true,
			})
			continue
		}

		var writeErr error
		if exists {
			writeErr = c.dst.UpdateGroupApprovalRule(groupPath, dst.ID, req)
		} else {
			writeErr = c.dst.CreateGroupApprovalRule(groupPath, req)
		}

		if writeErr != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  writeErr,
			})
			continue
		}

		item := internal.ItemResult{Key: src.Name, Action: action}
		// For regular rules the rule is created but approvers (user/group IDs)
		// are instance-specific and cannot be copied — flag for manual follow-up
		if src.RuleType == "regular" {
			item.Error = fmt.Errorf("rule created but approvers not copied — user/group IDs are instance-specific, assign manually")
		}
		result.Items = append(result.Items, item)
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
