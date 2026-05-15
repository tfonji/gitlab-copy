package copy

import (
	"fmt"
	"sort"
	"strings"

	"gitlab-copy/internal"
	"gitlab-copy/internal/gitlab"
)

type GroupCopier struct {
	src             *gitlab.Client
	dst             *gitlab.Client
	domains         []string
	dryRun          bool
	policyProject   string // dest security policy project (enforce_security_policy)
	mrPolicyProject string // dest MR policy project (link_merge_request_policy)
}

func NewGroupCopier(src, dst *gitlab.Client, domains []string, dryRun bool, policyProject, mrPolicyProject string) *GroupCopier {
	return &GroupCopier{
		src:             src,
		dst:             dst,
		domains:         domains,
		dryRun:          dryRun,
		policyProject:   policyProject,
		mrPolicyProject: mrPolicyProject,
	}
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
	case "group_settings":
		return c.copyGroupSettings(groupPath)
	case "hooks":
		return c.copyGroupHooks(groupPath)
	case "push_rules":
		return c.copyPushRules(groupPath)
	case "description":
		return c.copyDescription(groupPath)
	case "default_branch_name":
		return c.copyDefaultBranchName(groupPath)
	case "default_branch_protection":
		return c.copyDefaultBranchProtection(groupPath)
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
	case "badges":
		return c.copyGroupBadges(groupPath)
	case "compliance_frameworks":
		return c.copyComplianceFrameworks(groupPath)
	case "compliance_assignments":
		return c.copyComplianceAssignments(groupPath)
	case "security_policy_project":
		return c.copySecurityPolicyProject(groupPath)
	case "enforce_security_policy":
		return c.enforceSecurityPolicy(groupPath)
	case "link_merge_request_policy":
		return c.linkMergeRequestPolicy(groupPath)
	case "deploy_tokens":
		return c.copyGroupDeployTokens(groupPath)
	case "access_tokens":
		return c.copyGroupAccessTokens(groupPath)
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

	var diffs []internal.DiffLine
	if dstExists {
		action = internal.ActionUpdated
		diffs = pushRuleDiffs(src, dst)
	} else {
		action = internal.ActionCreated
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "push_rules", Action: action, DryRun: true, Diffs: diffs},
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
		{Key: "push_rules", Action: action, Diffs: diffs},
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

	diffs := []internal.DiffLine{{Field: "default_branch_name", Src: dst.DefaultBranchName, Dst: src.DefaultBranchName}}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_name", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
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
		{Key: "default_branch_name", Action: internal.ActionUpdated, Diffs: diffs},
	}
	return result
}

// --- default_branch_protection ---

func (c *GroupCopier) copyDefaultBranchProtection(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "default_branch_protection"}

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

	diffs := defaultBranchProtectionDiffs(src, dst)
	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_protection", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_protection", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
		}
		return result
	}

	req := gitlab.GroupUpdateRequest{
		DefaultBranchProtection:         gitlab.IntPtr(src.DefaultBranchProtection),
		DefaultBranchProtectionDefaults: src.DefaultBranchProtectionDefaults,
	}
	if err := c.dst.UpdateGroup(groupPath, req); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "default_branch_protection", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "default_branch_protection", Action: internal.ActionUpdated, Diffs: diffs},
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

	diffs := mrSettingsDiffs(src, dst)
	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{
			{Key: "mr_settings", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "mr_settings", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
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
		{Key: "mr_settings", Action: internal.ActionUpdated, Diffs: diffs},
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

	if dst == nil {
		dst = &gitlab.MergeRequestApprovalSettings{}
	}

	diffs := mrApprovalSettingsDiffs(src, dst)
	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "mr_approval_settings", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
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
		{Key: "mr_approval_settings", Action: internal.ActionUpdated, Diffs: diffs},
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

	diffs := []internal.DiffLine{{Field: "description", Src: src.Description, Dst: dst.Description}}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "description", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
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
		{Key: "description", Action: internal.ActionUpdated, Diffs: diffs},
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

// --- badges ---

func (c *GroupCopier) copyGroupBadges(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "badges"}

	srcBadges, err := c.src.GetGroupBadges(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source badges: %w", err)
		return result
	}
	dstBadges, err := c.dst.GetGroupBadges(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest badges: %w", err)
		return result
	}

	badgeKey := func(b gitlab.Badge) string { return b.LinkURL + "|" + b.ImageURL }

	srcByKey := make(map[string]gitlab.Badge, len(srcBadges))
	for _, b := range srcBadges {
		srcByKey[badgeKey(b)] = b
	}
	dstByKey := make(map[string]gitlab.Badge, len(dstBadges))
	for _, b := range dstBadges {
		dstByKey[badgeKey(b)] = b
	}

	// Delete dest badges not present on source
	for key, dstBadge := range dstByKey {
		if _, exists := srcByKey[key]; !exists {
			if c.dryRun {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    dstBadge.Name,
					Action: internal.ActionUpdated,
					DryRun: true,
					Error:  fmt.Errorf("extra badge on dest would be deleted"),
				})
				continue
			}
			if err := c.dst.DeleteGroupBadge(groupPath, dstBadge.ID); err != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    dstBadge.Name,
					Action: internal.ActionFailed,
					Error:  fmt.Errorf("deleting extra badge: %w", err),
				})
			}
		}
	}

	// Create source badges missing on dest
	sort.Slice(srcBadges, func(i, j int) bool {
		return srcBadges[i].Name < srcBadges[j].Name
	})
	for _, srcBadge := range srcBadges {
		if _, exists := dstByKey[badgeKey(srcBadge)]; exists {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    srcBadge.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    srcBadge.Name,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.BadgeRequest{
			Name:     srcBadge.Name,
			LinkURL:  srcBadge.LinkURL,
			ImageURL: srcBadge.ImageURL,
		}
		if err := c.dst.CreateGroupBadge(groupPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    srcBadge.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    srcBadge.Name,
				Action: internal.ActionCreated,
			})
		}
	}

	return result
}

// --- compliance_assignments ---

// copyComplianceAssignments assigns compliance frameworks to projects on dest
// matching the source assignments. It resolves dest framework IDs by name —
// so compliance_frameworks should be run before this domain to ensure frameworks
// exist on dest.
func (c *GroupCopier) copyComplianceAssignments(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "compliance_assignments"}

	srcAssignments, err := c.src.GetGroupComplianceAssignments(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source compliance assignments: %w", err)
		return result
	}
	if len(srcAssignments) == 0 {
		return result
	}

	// Build dest framework name→ID map
	dstFrameworks, err := c.dst.GetGroupComplianceFrameworks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest compliance frameworks: %w", err)
		return result
	}
	dstIDByName := make(map[string]string, len(dstFrameworks))
	for _, fw := range dstFrameworks {
		dstIDByName[fw.Name] = fw.ID
	}

	sort.Slice(srcAssignments, func(i, j int) bool {
		return srcAssignments[i].ProjectPath < srcAssignments[j].ProjectPath
	})

	for _, assignment := range srcAssignments {
		// Collect all framework IDs for this project in one batch
		var frameworkIDs []string
		var missingFrameworks []string

		for _, fwName := range assignment.FrameworkNames {
			dstID, ok := dstIDByName[fwName]
			if !ok {
				missingFrameworks = append(missingFrameworks, fwName)
				continue
			}
			frameworkIDs = append(frameworkIDs, dstID)
		}

		// Report missing frameworks
		for _, fwName := range missingFrameworks {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    assignment.ProjectPath + " → " + fwName,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
				Error:  fmt.Errorf("framework %q not found on dest — run compliance_frameworks first", fwName),
			})
		}

		if len(frameworkIDs) == 0 {
			continue
		}

		if c.dryRun {
			for _, fwName := range assignment.FrameworkNames {
				if _, ok := dstIDByName[fwName]; ok {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    assignment.ProjectPath + " → " + fwName,
						Action: internal.ActionCreated,
						DryRun: true,
					})
				}
			}
			continue
		}

		// Single call with all framework IDs for this project
		if err := c.dst.AssignComplianceFrameworks(assignment.ProjectPath, frameworkIDs); err != nil {
			if strings.Contains(err.Error(), "doesn't exist on type 'Mutation'") {
				for _, fwName := range assignment.FrameworkNames {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    assignment.ProjectPath + " → " + fwName,
						Action: internal.ActionSkipped,
						Error:  fmt.Errorf("projectUpdateComplianceFrameworks mutation not supported on this GitLab version — assign manually via UI"),
					})
				}
			} else {
				for _, fwName := range assignment.FrameworkNames {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    assignment.ProjectPath + " → " + fwName,
						Action: internal.ActionFailed,
						Error:  err,
					})
				}
			}
		} else {
			for _, fwName := range assignment.FrameworkNames {
				if _, ok := dstIDByName[fwName]; ok {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    assignment.ProjectPath + " → " + fwName,
						Action: internal.ActionCreated,
					})
				}
			}
		}
	}

	return result
}

// --- security_policy_project ---

// copySecurityPolicyProject links the same security policy project on dest
// that is linked on source. The security policy project itself must already
// exist on dest (migrated by Congregate) with the same full path.
func (c *GroupCopier) copySecurityPolicyProject(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "security_policy_project"}

	src, err := c.src.GetGroupSecurityPolicyProject(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source security policy project: %w", err)
		return result
	}
	if src == nil {
		result.Items = []internal.ItemResult{
			{Key: "security_policy_project", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	dst, err := c.dst.GetGroupSecurityPolicyProject(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest security policy project: %w", err)
		return result
	}

	// Already linked to the same project
	if dst != nil && dst.FullPath == src.FullPath {
		result.Items = []internal.ItemResult{
			{Key: src.FullPath, Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		action := internal.ActionCreated
		if dst != nil {
			action = internal.ActionUpdated
		}
		result.Items = []internal.ItemResult{
			{Key: src.FullPath, Action: action, DryRun: true},
		}
		return result
	}

	if err := c.dst.LinkSecurityPolicyProject(groupPath, src.FullPath); err != nil {
		errMsg := err
		if apiErr, ok := err.(*gitlab.APIError); ok && apiErr.IsNotFound() {
			errMsg = fmt.Errorf("security policy project %q not found on dest — ensure Congregate has migrated it first", src.FullPath)
		}
		result.Items = []internal.ItemResult{
			{Key: src.FullPath, Action: internal.ActionFailed, Error: errMsg},
		}
		return result
	}

	action := internal.ActionCreated
	if dst != nil {
		action = internal.ActionUpdated
	}
	result.Items = []internal.ItemResult{
		{
			Key:    src.FullPath,
			Action: action,
			Error:  fmt.Errorf("verify the security policy project exists on dest before running"),
		},
	}
	return result
}

// --- deploy_tokens ---

func (c *GroupCopier) copyGroupDeployTokens(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "deploy_tokens"}

	srcTokens, err := c.src.GetGroupDeployTokens(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source deploy tokens: %w", err)
		return result
	}
	dstTokens, err := c.dst.GetGroupDeployTokens(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest deploy tokens: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstTokens))
	for _, t := range dstTokens {
		dstByName[t.Name] = true
	}

	sort.Slice(srcTokens, func(i, j int) bool {
		return srcTokens[i].Name < srcTokens[j].Name
	})

	for _, src := range srcTokens {
		if dstByName[src.Name] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.DeployTokenRequest{
			Name:      src.Name,
			Username:  src.Username,
			ExpiresAt: src.ExpiresAt,
			Scopes:    src.Scopes,
		}
		resp, err := c.dst.CreateGroupDeployToken(groupPath, req)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionCreated,
				Token:  resp.Token,
				Error:  fmt.Errorf("new token generated — update any services referencing the source token"),
			})
		}
	}

	return result
}

// --- access_tokens ---

func (c *GroupCopier) copyGroupAccessTokens(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "access_tokens"}

	srcTokens, err := c.src.GetGroupAccessTokens(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source access tokens: %w", err)
		return result
	}
	dstTokens, err := c.dst.GetGroupAccessTokens(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest access tokens: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstTokens))
	for _, t := range dstTokens {
		dstByName[t.Name] = true
	}

	sort.Slice(srcTokens, func(i, j int) bool {
		return srcTokens[i].Name < srcTokens[j].Name
	})

	for _, src := range srcTokens {
		if dstByName[src.Name] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.AccessTokenRequest{
			Name:        src.Name,
			Scopes:      src.Scopes,
			ExpiresAt:   src.ExpiresAt,
			AccessLevel: src.AccessLevel,
		}
		resp, err := c.dst.CreateGroupAccessToken(groupPath, req)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionCreated,
				Token:  resp.Token,
				Error:  fmt.Errorf("new token generated — update any services referencing the source token"),
			})
		}
	}

	return result
}

// --- enforce_security_policy ---

// enforceSecurityPolicy runs only on root groups (parent_id == nil).
// It deletes any compliance frameworks on dest, then links the configured
// policy project to the group. The policy project path comes from
// config.security.policy_project.
func (c *GroupCopier) enforceSecurityPolicy(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "enforce_security_policy"}

	if c.policyProject == "" {
		result.Items = []internal.ItemResult{{
			Key:    groupPath,
			Action: internal.ActionSkipped,
			Error:  fmt.Errorf("security.policy_project not set in config — skipping"),
		}}
		return result
	}

	// Only run on root groups
	dstGroup, err := c.dst.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest group: %w", err)
		return result
	}
	if dstGroup.ParentID != nil {
		result.Items = []internal.ItemResult{{
			Key:    groupPath,
			Action: internal.ActionSkipped,
			DryRun: c.dryRun,
		}}
		return result
	}

	// Step 1 — delete any compliance frameworks on dest
	frameworks, err := c.dst.GetGroupComplianceFrameworks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest compliance frameworks: %w", err)
		return result
	}

	for _, fw := range frameworks {
		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fmt.Sprintf("delete framework: %s", fw.Name),
				Action: internal.ActionCreated, // represents "would delete"
				DryRun: true,
			})
			continue
		}
		if err := c.dst.DeleteComplianceFramework(fw.ID, fw.Default); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fmt.Sprintf("delete framework: %s", fw.Name),
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fmt.Sprintf("deleted framework: %s", fw.Name),
				Action: internal.ActionCreated,
			})
		}
	}

	// Step 2 — link the policy project
	existing, err := c.dst.GetGroupSecurityPolicyProject(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest security policy project: %w", err)
		return result
	}

	if existing != nil && existing.FullPath == c.policyProject {
		result.Items = append(result.Items, internal.ItemResult{
			Key:    c.policyProject,
			Action: internal.ActionSkipped,
			DryRun: c.dryRun,
		})
		return result
	}

	if c.dryRun {
		result.Items = append(result.Items, internal.ItemResult{
			Key:    c.policyProject,
			Action: internal.ActionCreated,
			DryRun: true,
		})
		return result
	}

	if err := c.dst.LinkSecurityPolicyProject(groupPath, c.policyProject); err != nil {
		result.Items = append(result.Items, internal.ItemResult{
			Key:    c.policyProject,
			Action: internal.ActionFailed,
			Error:  err,
		})
	} else {
		result.Items = append(result.Items, internal.ItemResult{
			Key:    c.policyProject,
			Action: internal.ActionCreated,
		})
	}

	return result
}

// --- link_merge_request_policy ---

// linkMergeRequestPolicy links the configured MR policy project to root groups.
// Unlike enforce_security_policy, it does NOT delete compliance frameworks.
func (c *GroupCopier) linkMergeRequestPolicy(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "link_merge_request_policy"}

	if c.mrPolicyProject == "" {
		result.Items = []internal.ItemResult{{
			Key:    groupPath,
			Action: internal.ActionSkipped,
			Error:  fmt.Errorf("security.merge_request_policy_project not set in config — skipping"),
		}}
		return result
	}

	// Only run on root groups
	dstGroup, err := c.dst.GetGroup(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest group: %w", err)
		return result
	}
	if dstGroup.ParentID != nil {
		result.Items = []internal.ItemResult{{
			Key:    groupPath,
			Action: internal.ActionSkipped,
			DryRun: c.dryRun,
		}}
		return result
	}

	// Check if already linked to the same project
	existing, err := c.dst.GetGroupSecurityPolicyProject(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest security policy project: %w", err)
		return result
	}

	if existing != nil && existing.FullPath == c.mrPolicyProject {
		result.Items = []internal.ItemResult{{
			Key:    c.mrPolicyProject,
			Action: internal.ActionSkipped,
			DryRun: c.dryRun,
		}}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{{
			Key:    c.mrPolicyProject,
			Action: internal.ActionCreated,
			DryRun: true,
		}}
		return result
	}

	if err := c.dst.LinkSecurityPolicyProject(groupPath, c.mrPolicyProject); err != nil {
		result.Items = []internal.ItemResult{{
			Key:    c.mrPolicyProject,
			Action: internal.ActionFailed,
			Error:  err,
		}}
	} else {
		result.Items = []internal.ItemResult{{
			Key:    c.mrPolicyProject,
			Action: internal.ActionCreated,
		}}
	}

	return result
}

// --- group_settings ---

func (c *GroupCopier) copyGroupSettings(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "group_settings"}

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

	diffs := []internal.DiffLine{
		{Field: "mentions_disabled", Src: fmt.Sprintf("%v", src.MentionsDisabled), Dst: fmt.Sprintf("%v", dst.MentionsDisabled), Match: fmt.Sprintf("%v", src.MentionsDisabled) == fmt.Sprintf("%v", dst.MentionsDisabled)},
		{Field: "prevent_sharing_groups_outside_hierarchy", Src: fmt.Sprintf("%v", src.PreventSharingGroupsOutsideHierarchy), Dst: fmt.Sprintf("%v", dst.PreventSharingGroupsOutsideHierarchy), Match: src.PreventSharingGroupsOutsideHierarchy == dst.PreventSharingGroupsOutsideHierarchy},
	}

	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{{Key: "group_settings", Action: internal.ActionSkipped, Diffs: diffs, DryRun: c.dryRun}}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{{Key: "group_settings", Action: internal.ActionUpdated, Diffs: diffs, DryRun: true}}
		return result
	}

	if err := c.dst.UpdateGroup(groupPath, gitlab.GroupUpdateRequest{
		MentionsDisabled:                     src.MentionsDisabled,
		PreventSharingGroupsOutsideHierarchy: gitlab.BoolPtr(src.PreventSharingGroupsOutsideHierarchy),
	}); err != nil {
		result.Items = []internal.ItemResult{{Key: "group_settings", Action: internal.ActionFailed, Error: err}}
		return result
	}

	result.Items = []internal.ItemResult{{Key: "group_settings", Action: internal.ActionUpdated, Diffs: diffs}}
	return result
}

// --- hooks ---

// --- hooks ---

func hookRequest(h gitlab.GroupHook) gitlab.GroupHookRequest {
	return gitlab.GroupHookRequest{
		URL:                       h.URL,
		Name:                      h.Name,
		Description:               h.Description,
		PushEvents:                h.PushEvents,
		PushEventsBranchFilter:    h.PushEventsBranchFilter,
		BranchFilterStrategy:      h.BranchFilterStrategy,
		TagPushEvents:             h.TagPushEvents,
		MergeRequestsEvents:       h.MergeRequestsEvents,
		IssuesEvents:              h.IssuesEvents,
		ConfidentialIssuesEvents:  h.ConfidentialIssuesEvents,
		NoteEvents:                h.NoteEvents,
		ConfidentialNoteEvents:    h.ConfidentialNoteEvents,
		PipelineEvents:            h.PipelineEvents,
		WikiPageEvents:            h.WikiPageEvents,
		JobEvents:                 h.JobEvents,
		DeploymentEvents:          h.DeploymentEvents,
		ReleasesEvents:            h.ReleasesEvents,
		SubGroupEvents:            h.SubGroupEvents,
		MemberEvents:              h.MemberEvents,
		FeatureFlagEvents:         h.FeatureFlagEvents,
		EnableSSLVerification:     h.EnableSSLVerification,
		MilestoneEvents:           h.MilestoneEvents,
		VulnerabilityEvents:       h.VulnerabilityEvents,
		EmojiEvents:               h.EmojiEvents,
		ResourceAccessTokenEvents: h.ResourceAccessTokenEvents,
		ProjectEvents:             h.ProjectEvents,
	}
}

func hooksMatch(src, dst gitlab.GroupHook) bool {
	return src.Name == dst.Name &&
		src.Description == dst.Description &&
		src.PushEvents == dst.PushEvents &&
		src.PushEventsBranchFilter == dst.PushEventsBranchFilter &&
		src.BranchFilterStrategy == dst.BranchFilterStrategy &&
		src.TagPushEvents == dst.TagPushEvents &&
		src.MergeRequestsEvents == dst.MergeRequestsEvents &&
		src.IssuesEvents == dst.IssuesEvents &&
		src.ConfidentialIssuesEvents == dst.ConfidentialIssuesEvents &&
		src.NoteEvents == dst.NoteEvents &&
		src.ConfidentialNoteEvents == dst.ConfidentialNoteEvents &&
		src.PipelineEvents == dst.PipelineEvents &&
		src.WikiPageEvents == dst.WikiPageEvents &&
		src.JobEvents == dst.JobEvents &&
		src.DeploymentEvents == dst.DeploymentEvents &&
		src.ReleasesEvents == dst.ReleasesEvents &&
		src.SubGroupEvents == dst.SubGroupEvents &&
		src.MemberEvents == dst.MemberEvents &&
		src.FeatureFlagEvents == dst.FeatureFlagEvents &&
		src.EnableSSLVerification == dst.EnableSSLVerification &&
		src.MilestoneEvents == dst.MilestoneEvents &&
		src.VulnerabilityEvents == dst.VulnerabilityEvents &&
		src.EmojiEvents == dst.EmojiEvents &&
		src.ResourceAccessTokenEvents == dst.ResourceAccessTokenEvents &&
		src.ProjectEvents == dst.ProjectEvents
}

func (c *GroupCopier) copyGroupHooks(groupPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "hooks"}

	srcHooks, err := c.src.GetGroupHooks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source hooks: %w", err)
		return result
	}
	if len(srcHooks) == 0 {
		return result
	}

	dstHooks, err := c.dst.GetGroupHooks(groupPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest hooks: %w", err)
		return result
	}

	// Match by URL
	dstByURL := make(map[string]gitlab.GroupHook, len(dstHooks))
	for _, h := range dstHooks {
		dstByURL[h.URL] = h
	}

	tokenWarning := fmt.Errorf("webhook token cannot be copied — set token manually on dest if required")

	for _, src := range srcHooks {
		req := hookRequest(src)

		if dst, exists := dstByURL[src.URL]; exists {
			if hooksMatch(src, dst) {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionSkipped,
					DryRun: c.dryRun,
				})
				continue
			}

			// Fields differ — update
			if c.dryRun {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionUpdated,
					DryRun: true,
					Error:  tokenWarning,
				})
				continue
			}

			if err := c.dst.UpdateGroupHook(groupPath, dst.ID, req); err != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionFailed,
					Error:  err,
				})
			} else {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionUpdated,
					Error:  tokenWarning,
				})
			}
			continue
		}

		// Not on dest — create
		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.URL,
				Action: internal.ActionCreated,
				DryRun: true,
				Error:  tokenWarning,
			})
			continue
		}

		if err := c.dst.CreateGroupHook(groupPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.URL,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.URL,
				Action: internal.ActionCreated,
				Error:  tokenWarning,
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
