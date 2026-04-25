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
	case "description":
		return c.copyDescription(groupPath)
	case "default_branch_name":
		return c.copyDefaultBranchName(groupPath)
	case "mr_settings":
		return c.copyMRSettings(groupPath)
	case "mr_approval_settings":
		return c.copyMRApprovalSettings(groupPath)
	case "protected_environments":
		return c.copyProtectedEnvironments(groupPath)
	case "approval_rules":
		return c.copyApprovalRules(groupPath)
	case "jira_integration":
		return c.copyJiraIntegration(groupPath)
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

// --- mr_approval_settings ---

func (c *GroupCopier) copyMRApprovalSettings(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "mr_approval_settings"}

	src, err := c.src.GetMRApprovalSettings(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source MR approval settings: %w", err)
		return result
	}
	if src == nil {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}
	dst, err := c.dst.GetMRApprovalSettings(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest MR approval settings: %w", err)
		return result
	}

	matches := dst != nil &&
		src.AllowAuthorApproval.Value == dst.AllowAuthorApproval.Value &&
		src.AllowCommitterApproval.Value == dst.AllowCommitterApproval.Value &&
		src.AllowOverridesToApproverList.Value == dst.AllowOverridesToApproverList.Value &&
		src.RequirePasswordToApprove.Value == dst.RequirePasswordToApprove.Value &&
		src.RetainApprovalsOnPush.Value == dst.RetainApprovalsOnPush.Value &&
		src.SelectiveCodeOwnerRemovals.Value == dst.SelectiveCodeOwnerRemovals.Value

	if matches {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionUpdated, DryRun: true},
		}
		return result
	}

	req := gitlab.MergeRequestApprovalSettingsRequest{
		AllowAuthorApproval:          src.AllowAuthorApproval.Value,
		AllowCommitterApproval:       src.AllowCommitterApproval.Value,
		AllowOverridesToApproverList: src.AllowOverridesToApproverList.Value,
		RequirePasswordToApprove:     src.RequirePasswordToApprove.Value,
		RetainApprovalsOnPush:        src.RetainApprovalsOnPush.Value,
		SelectiveCodeOwnerRemovals:   src.SelectiveCodeOwnerRemovals.Value,
	}
	if err := c.dst.SetMRApprovalSettings(groupPath, req); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "mr_approval_settings", Action: internal.ActionUpdated},
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

// --- description ---

func (c *GroupCopier) copyDescription(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "description"}

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

	if src.Description == dst.Description {
		result.Items = []internal.ItemResult{
			{Key: "description", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "description", Action: internal.ActionUpdated, DryRun: true},
		}
		return result
	}

	if err := c.dst.UpdateGroup(groupPath, gitlab.GroupUpdateRequest{
		Description: gitlab.StrPtr(src.Description),
	}); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "description", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "description", Action: internal.ActionUpdated},
	}
	return result
}

// --- variables ---

func (c *GroupCopier) copyVariables(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "variables"}

	srcVars, err := c.src.GetGroupVariables(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source variables: %w", err)
		return result
	}
	dstVars, err := c.dst.GetGroupVariables(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest variables: %w", err)
		return result
	}

	// Index dest by key::scope
	dstByKey := make(map[string]gitlab.Variable, len(dstVars))
	for _, v := range dstVars {
		dstByKey[v.Key+"::"+v.EnvironmentScope] = v
	}

	sort.Slice(srcVars, func(i, j int) bool {
		return srcVars[i].Key < srcVars[j].Key
	})

	for _, src := range srcVars {
		itemKey := src.Key + "::" + src.EnvironmentScope

		// Skip masked/hidden — value cannot be read from API
		if src.IsSensitive() {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    itemKey,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
				Error:  fmt.Errorf("masked/hidden variable — value not accessible, create manually on dest"),
			})
			continue
		}

		dst, exists := dstByKey[itemKey]

		// Check if all copyable fields match
		if exists &&
			src.Value == dst.Value &&
			src.VariableType == dst.VariableType &&
			src.Protected == dst.Protected &&
			src.Raw == dst.Raw &&
			src.Description == dst.Description {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    itemKey,
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
				Key:    itemKey,
				Action: action,
				DryRun: true,
			})
			continue
		}

		req := gitlab.VariableRequest{
			Key:              src.Key,
			Value:            src.Value,
			VariableType:     src.VariableType,
			Protected:        src.Protected,
			Masked:           src.Masked,
			Raw:              src.Raw,
			EnvironmentScope: src.EnvironmentScope,
			Description:      src.Description,
		}

		var writeErr error
		if exists {
			writeErr = c.dst.UpdateGroupVariable(groupPath, src.Key, src.EnvironmentScope, req)
		} else {
			writeErr = c.dst.CreateGroupVariable(groupPath, req)
		}

		if writeErr != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    itemKey,
				Action: internal.ActionFailed,
				Error:  writeErr,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    itemKey,
				Action: action,
			})
		}
	}

	return result
}

// --- jira_integration ---

var requiredGroupJiraFields = []string{"password", "url"}

func (c *GroupCopier) copyJiraIntegration(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "jira_integration"}

	src, err := c.src.GetGroupJiraIntegration(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source Jira integration: %w", err)
		return result
	}
	if src == nil {
		result.Items = []internal.ItemResult{
			{Key: "jira_integration", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	// Credentials are masked in GET responses — if missing, flag as manual
	for _, field := range requiredGroupJiraFields {
		val, ok := src.Properties[field]
		if !ok || val == nil || val == "" {
			result.Items = []internal.ItemResult{
				{
					Key:    "jira_integration",
					Action: internal.ActionSkipped,
					DryRun: c.dryRun,
					Error:  fmt.Errorf("credentials masked in source API response — configure Jira integration manually on dest"),
				},
			}
			return result
		}
	}

	dst, err := c.dst.GetGroupJiraIntegration(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest Jira integration: %w", err)
		return result
	}

	action := internal.ActionCreated
	if dst != nil {
		action = internal.ActionUpdated
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "jira_integration", Action: action, DryRun: true},
		}
		return result
	}

	if err := c.dst.SetGroupJiraIntegration(groupPath, src.Properties); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "jira_integration", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{
			Key:    "jira_integration",
			Action: action,
			Error:  fmt.Errorf("verify credentials on dest — password/token values may not have transferred correctly"),
		},
	}
	return result
}

// --- compliance_frameworks ---

func (c *GroupCopier) copyComplianceFrameworks(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "compliance_frameworks"}

	srcFrameworks, err := c.src.GetGroupComplianceFrameworks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source compliance frameworks: %w", err)
		return result
	}
	dstFrameworks, err := c.dst.GetGroupComplianceFrameworks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest compliance frameworks: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstFrameworks))
	for _, f := range dstFrameworks {
		dstByName[f.Name] = true
	}

	sort.Slice(srcFrameworks, func(i, j int) bool {
		return srcFrameworks[i].Name < srcFrameworks[j].Name
	})

	for _, fw := range srcFrameworks {
		if dstByName[fw.Name] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fw.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fw.Name,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		_, err := c.dst.CreateComplianceFramework(groupPath, fw)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fw.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			item := internal.ItemResult{Key: fw.Name, Action: internal.ActionCreated}
			if fw.PipelineConfigurationFullPath != "" {
				item.Error = fmt.Errorf("pipeline config path references source instance — verify %q is valid on dest", fw.PipelineConfigurationFullPath)
			}
			result.Items = append(result.Items, item)
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
