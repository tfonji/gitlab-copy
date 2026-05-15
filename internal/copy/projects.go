package copy

import (
	"fmt"
	"sort"
	"strings"

	"gitlab-copy/internal"
	"gitlab-copy/internal/gitlab"
)

type ProjectCopier struct {
	src     *gitlab.Client
	dst     *gitlab.Client
	domains []string
	dryRun  bool
}

func NewProjectCopier(src, dst *gitlab.Client, domains []string, dryRun bool) *ProjectCopier {
	return &ProjectCopier{src: src, dst: dst, domains: domains, dryRun: dryRun}
}

func (c *ProjectCopier) Copy(projectPath string) []internal.DomainCopyResult {
	results := make([]internal.DomainCopyResult, 0, len(c.domains))
	for _, domain := range c.domains {
		results = append(results, c.copyDomain(projectPath, domain))
	}
	return results
}

func (c *ProjectCopier) copyDomain(projectPath, domain string) internal.DomainCopyResult {
	switch domain {
	case "topics":
		return c.copyTopics(projectPath)
	case "project_settings":
		return c.copyProjectSettings(projectPath)
	case "project_hooks":
		return c.copyProjectHooks(projectPath)
	case "environments":
		return c.copyEnvironments(projectPath)
	case "protected_environments":
		return c.copyProtectedEnvironments(projectPath)
	case "jira_integration":
		return c.copyJiraIntegration(projectPath)
	case "pipeline_triggers":
		return c.copyPipelineTriggers(projectPath)
	case "deploy_keys":
		return c.copyDeployKeys(projectPath)
	case "project_push_rules":
		return c.copyProjectPushRules(projectPath)
	case "project_mr_approvals":
		return c.copyProjectMRApprovals(projectPath)
	case "project_approval_rules":
		return c.copyProjectApprovalRules(projectPath)
	case "badges":
		return c.copyBadges(projectPath)
	case "project_protected_branches":
		return c.copyProtectedBranches(projectPath)
	case "project_protected_tags":
		return c.copyProtectedTags(projectPath)
	case "deploy_tokens":
		return c.copyProjectDeployTokens(projectPath)
	case "access_tokens":
		return c.copyProjectAccessTokens(projectPath)
	case "pipeline_schedules":
		return c.copyPipelineSchedules(projectPath)
	case "job_token_scope":
		return c.copyJobTokenScope(projectPath)
	default:
		return internal.DomainCopyResult{
			Domain: domain,
			Error:  fmt.Errorf("unknown domain %q", domain),
		}
	}
}

// --- project_settings ---

func (c *ProjectCopier) copyProjectSettings(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_settings"}

	src, err := c.src.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source project: %w", err)
		return result
	}
	dst, err := c.dst.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest project: %w", err)
		return result
	}

	diffs := []internal.DiffLine{
		{Field: "description", Src: src.Description, Dst: dst.Description, Match: src.Description == dst.Description},
		{Field: "auto_cancel_pending_pipelines", Src: src.AutoCancelPendingPipelines, Dst: dst.AutoCancelPendingPipelines, Match: src.AutoCancelPendingPipelines == dst.AutoCancelPendingPipelines},
		{Field: "ci_forward_deployment_enabled", Src: fmt.Sprintf("%v", src.CIForwardDeploymentEnabled), Dst: fmt.Sprintf("%v", dst.CIForwardDeploymentEnabled), Match: fmt.Sprintf("%v", src.CIForwardDeploymentEnabled) == fmt.Sprintf("%v", dst.CIForwardDeploymentEnabled)},
		{Field: "ci_separated_caches", Src: fmt.Sprintf("%v", src.CISeperateCache), Dst: fmt.Sprintf("%v", dst.CISeperateCache), Match: fmt.Sprintf("%v", src.CISeperateCache) == fmt.Sprintf("%v", dst.CISeperateCache)},
		{Field: "printing_merge_request_link_enabled", Src: fmt.Sprintf("%v", src.PrintingMergeRequestLinkEnabled), Dst: fmt.Sprintf("%v", dst.PrintingMergeRequestLinkEnabled), Match: src.PrintingMergeRequestLinkEnabled == dst.PrintingMergeRequestLinkEnabled},
		{Field: "remove_source_branch_after_merge", Src: fmt.Sprintf("%v", src.RemoveSourceBranchAfterMerge), Dst: fmt.Sprintf("%v", dst.RemoveSourceBranchAfterMerge), Match: src.RemoveSourceBranchAfterMerge == dst.RemoveSourceBranchAfterMerge},
		{Field: "allow_merge_on_skipped_pipeline", Src: fmt.Sprintf("%v", src.AllowMergeOnSkippedPipeline), Dst: fmt.Sprintf("%v", dst.AllowMergeOnSkippedPipeline), Match: fmt.Sprintf("%v", src.AllowMergeOnSkippedPipeline) == fmt.Sprintf("%v", dst.AllowMergeOnSkippedPipeline)},
	}

	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{{Key: "project_settings", Action: internal.ActionSkipped, Diffs: diffs, DryRun: c.dryRun}}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{{Key: "project_settings", Action: internal.ActionUpdated, Diffs: diffs, DryRun: true}}
		return result
	}

	req := gitlab.ProjectUpdateRequest{
		Description:                     gitlab.StrPtr(src.Description),
		AutoCancelPendingPipelines:      gitlab.StrPtr(src.AutoCancelPendingPipelines),
		CIForwardDeploymentEnabled:      src.CIForwardDeploymentEnabled,
		CISeperateCache:                 src.CISeperateCache,
		PrintingMergeRequestLinkEnabled: gitlab.BoolPtr(src.PrintingMergeRequestLinkEnabled),
		RemoveSourceBranchAfterMerge:    gitlab.BoolPtr(src.RemoveSourceBranchAfterMerge),
		AllowMergeOnSkippedPipeline:     src.AllowMergeOnSkippedPipeline,
	}

	if err := c.dst.UpdateProject(projectPath, req); err != nil {
		result.Items = []internal.ItemResult{{Key: "project_settings", Action: internal.ActionFailed, Error: err}}
		return result
	}

	result.Items = []internal.ItemResult{{Key: "project_settings", Action: internal.ActionUpdated, Diffs: diffs}}
	return result
}

// --- topics ---

func (c *ProjectCopier) copyTopics(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "topics"}

	src, err := c.src.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source project: %w", err)
		return result
	}
	dst, err := c.dst.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest project: %w", err)
		return result
	}

	if len(src.Topics) == 0 {
		return result
	}

	// Build dest topic set for comparison
	dstTopics := make(map[string]bool, len(dst.Topics))
	for _, t := range dst.Topics {
		dstTopics[t] = true
	}

	// Determine per-topic actions
	var toCreate []string
	for _, t := range src.Topics {
		if dstTopics[t] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    t,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
		} else {
			toCreate = append(toCreate, t)
			result.Items = append(result.Items, internal.ItemResult{
				Key:    t,
				Action: internal.ActionCreated,
				DryRun: c.dryRun,
			})
		}
	}

	if len(toCreate) == 0 || c.dryRun {
		return result
	}

	// Apply — replace the full topic list with source topics
	if err := c.dst.UpdateProject(projectPath, gitlab.ProjectUpdateRequest{
		Topics: src.Topics,
	}); err != nil {
		// Mark the would-be-created items as failed
		for i := range result.Items {
			if result.Items[i].Action == internal.ActionCreated {
				result.Items[i].Action = internal.ActionFailed
				result.Items[i].Error = err
			}
		}
	}

	return result
}

// --- environments ---

func (c *ProjectCopier) copyEnvironments(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "environments"}

	srcEnvs, err := c.src.GetProjectEnvironments(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source environments: %w", err)
		return result
	}
	dstEnvs, err := c.dst.GetProjectEnvironments(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest environments: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstEnvs))
	for _, e := range dstEnvs {
		dstByName[e.Name] = true
	}

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

		req := gitlab.EnvironmentRequest{
			Name:        env.Name,
			ExternalURL: env.ExternalURL,
		}
		if err := c.dst.CreateProjectEnvironment(projectPath, req); err != nil {
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

// --- protected_environments ---

func (c *ProjectCopier) copyProtectedEnvironments(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "protected_environments"}

	srcEnvs, err := c.src.GetProjectProtectedEnvironments(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source protected environments: %w", err)
		return result
	}
	dstEnvs, err := c.dst.GetProjectProtectedEnvironments(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest protected environments: %w", err)
		return result
	}

	dstByName := make(map[string]bool, len(dstEnvs))
	for _, e := range dstEnvs {
		dstByName[e.Name] = true
	}

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
		if err := c.dst.CreateProjectProtectedEnvironment(projectPath, req); err != nil {
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

// --- jira_integration ---

// requiredJiraFields are fields the GitLab Jira integration API requires on PUT.
// These are typically masked in GET responses — if missing, the copy cannot proceed.
var requiredJiraFields = []string{"password", "url"}

func (c *ProjectCopier) copyJiraIntegration(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "jira_integration"}

	src, err := c.src.GetProjectJiraIntegration(projectPath)
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

	// Check that required credential fields are present and non-empty.
	// The GitLab API masks these in GET responses — if any are missing,
	// the PUT will fail with 400. Flag as manual rather than attempting a doomed write.
	for _, field := range requiredJiraFields {
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

	dst, err := c.dst.GetProjectJiraIntegration(projectPath)
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

	if err := c.dst.SetProjectJiraIntegration(projectPath, src.Properties); err != nil {
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

// --- pipeline_triggers ---

func (c *ProjectCopier) copyPipelineTriggers(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "pipeline_triggers"}

	srcTriggers, err := c.src.GetProjectTriggers(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source triggers: %w", err)
		return result
	}
	dstTriggers, err := c.dst.GetProjectTriggers(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest triggers: %w", err)
		return result
	}

	// Index dest by description — not enforced unique but best natural key
	dstByDesc := make(map[string]bool, len(dstTriggers))
	for _, t := range dstTriggers {
		dstByDesc[t.Description] = true
	}

	sort.Slice(srcTriggers, func(i, j int) bool {
		return srcTriggers[i].Description < srcTriggers[j].Description
	})

	for _, trigger := range srcTriggers {
		if dstByDesc[trigger.Description] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.PipelineTriggerRequest{Description: trigger.Description}
		resp, err := c.dst.CreateProjectTrigger(projectPath, req)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionCreated,
				Token:  resp.Token,
				Error:  fmt.Errorf("new token generated — update any CI variables or webhook URLs referencing the source token"),
			})
		}
	}

	return result
}

// --- deploy_keys ---

func (c *ProjectCopier) copyDeployKeys(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "deploy_keys"}

	// Fetch numeric IDs — the deploy keys API requires them
	srcProject, err := c.src.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source project: %w", err)
		return result
	}
	dstProject, err := c.dst.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest project: %w", err)
		return result
	}

	srcKeys, err := c.src.GetProjectDeployKeys(srcProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching source deploy keys: %w", err)
		return result
	}
	dstKeys, err := c.dst.GetProjectDeployKeys(dstProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest deploy keys: %w", err)
		return result
	}

	// Index dest by title — primary match key
	dstByTitle := make(map[string]bool, len(dstKeys))
	for _, k := range dstKeys {
		dstByTitle[k.Title] = true
	}

	sort.Slice(srcKeys, func(i, j int) bool {
		return srcKeys[i].Title < srcKeys[j].Title
	})

	for _, key := range srcKeys {
		if dstByTitle[key.Title] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    key.Title,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    key.Title,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		req := gitlab.DeployKeyRequest{
			Title:   key.Title,
			Key:     key.Key,
			CanPush: key.CanPush,
		}
		if err := c.dst.CreateProjectDeployKey(dstProject.ID, req); err != nil {
			// 422 means the public key already exists globally on the dest instance
			if apiErr, ok := err.(*gitlab.APIError); ok && apiErr.StatusCode == 422 {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    key.Title,
					Action: internal.ActionFailed,
					Error:  fmt.Errorf("key already exists on dest instance — enable it manually via Settings > Repository > Deploy Keys"),
				})
			} else {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    key.Title,
					Action: internal.ActionFailed,
					Error:  err,
				})
			}
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    key.Title,
				Action: internal.ActionCreated,
			})
		}
	}

	return result
}

// --- project_push_rules ---

func (c *ProjectCopier) copyProjectPushRules(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_push_rules"}

	src, err := c.src.GetProjectPushRules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source push rules: %w", err)
		return result
	}
	if src == nil {
		result.Error = fmt.Errorf("source push rules not accessible (403)")
		return result
	}
	if src.IsEmpty() {
		result.Items = []internal.ItemResult{
			{Key: "project_push_rules", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	dst, err := c.dst.GetProjectPushRules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest push rules: %w", err)
		return result
	}

	dstExists := dst != nil && !dst.IsEmpty()

	if dstExists && src.Equal(dst) {
		result.Items = []internal.ItemResult{
			{Key: "project_push_rules", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	action := internal.ActionCreated
	var diffs []internal.DiffLine
	if dstExists {
		action = internal.ActionUpdated
		diffs = pushRuleDiffs(src, dst)
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "project_push_rules", Action: action, DryRun: true, Diffs: diffs},
		}
		return result
	}

	req := gitlab.PushRuleRequestFrom(src)
	var writeErr error
	if dstExists {
		writeErr = c.dst.UpdateProjectPushRules(projectPath, req)
	} else {
		writeErr = c.dst.CreateProjectPushRules(projectPath, req)
	}
	if writeErr != nil {
		result.Items = []internal.ItemResult{
			{Key: "project_push_rules", Action: internal.ActionFailed, Error: writeErr},
		}
		return result
	}
	result.Items = []internal.ItemResult{
		{Key: "project_push_rules", Action: action, Diffs: diffs},
	}
	return result
}

// --- project_mr_approvals ---

func (c *ProjectCopier) copyProjectMRApprovals(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_mr_approvals"}

	src, err := c.src.GetProjectMRApprovals(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source MR approvals: %w", err)
		return result
	}
	if src == nil {
		result.Items = []internal.ItemResult{
			{Key: "project_mr_approvals", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	dst, err := c.dst.GetProjectMRApprovals(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest MR approvals: %w", err)
		return result
	}

	// If dst is nil there are no existing settings — treat as an update from defaults
	if dst == nil {
		dst = &gitlab.ProjectApprovalSettings{}
	}

	diffs := projectMRApprovalsDiffs(src, dst)
	if !hasChanges(diffs) {
		result.Items = []internal.ItemResult{
			{Key: "project_mr_approvals", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "project_mr_approvals", Action: internal.ActionUpdated, DryRun: true, Diffs: diffs},
		}
		return result
	}

	if err := c.dst.SetProjectMRApprovals(projectPath, src); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "project_mr_approvals", Action: internal.ActionFailed, Error: err},
		}
		return result
	}
	result.Items = []internal.ItemResult{
		{Key: "project_mr_approvals", Action: internal.ActionUpdated, Diffs: diffs},
	}
	return result
}

// --- project_approval_rules ---

func (c *ProjectCopier) copyProjectApprovalRules(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_approval_rules"}

	srcRules, err := c.src.GetProjectApprovalRules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source approval rules: %w", err)
		return result
	}
	dstRules, err := c.dst.GetProjectApprovalRules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest approval rules: %w", err)
		return result
	}

	dstByName := make(map[string]gitlab.ProjectApprovalRule, len(dstRules))
	for _, r := range dstRules {
		dstByName[r.Name] = r
	}

	sort.Slice(srcRules, func(i, j int) bool {
		return srcRules[i].Name < srcRules[j].Name
	})

	for _, src := range srcRules {
		if src.RuleType == "code_owner" {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		req := gitlab.ProjectApprovalRuleRequest{
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
			writeErr = c.dst.UpdateProjectApprovalRule(projectPath, dst.ID, req)
		} else {
			writeErr = c.dst.CreateProjectApprovalRule(projectPath, req)
		}

		if writeErr != nil {
			// GitLab only allows one any-approver rule per project.
			// If dest already has one, treat this as a skip rather than a failure.
			if apiErr, ok := writeErr.(*gitlab.APIError); ok && apiErr.StatusCode == 400 {
				if strings.Contains(apiErr.Body, "any-approver") {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    src.Name,
						Action: internal.ActionSkipped,
						DryRun: c.dryRun,
						Error:  fmt.Errorf("any-approver rule already exists on dest — skipped"),
					})
					continue
				}
			}
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  writeErr,
			})
			continue
		}

		item := internal.ItemResult{Key: src.Name, Action: action}
		if src.RuleType == "regular" {
			item.Error = fmt.Errorf("rule created but approvers not copied — user/group IDs are instance-specific, assign manually")
		}
		result.Items = append(result.Items, item)
	}

	return result
}

// --- badges ---

// Badges have no natural unique key — we match by link_url+image_url.
// Dest badges that don't exist on source are deleted (idempotent cleanup).
// Source badges missing on dest are created.
func (c *ProjectCopier) copyBadges(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "badges"}

	srcBadges, err := c.src.GetProjectBadges(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source badges: %w", err)
		return result
	}
	dstBadges, err := c.dst.GetProjectBadges(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest badges: %w", err)
		return result
	}

	// Build lookup by link_url+image_url as composite key
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
					Action: internal.ActionUpdated, // represents "would delete"
					DryRun: true,
					Error:  fmt.Errorf("extra badge on dest would be deleted"),
				})
				continue
			}
			if err := c.dst.DeleteProjectBadge(projectPath, dstBadge.ID); err != nil {
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
		key := badgeKey(srcBadge)
		if _, exists := dstByKey[key]; exists {
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
		if err := c.dst.CreateProjectBadge(projectPath, req); err != nil {
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

// --- project_protected_branches ---

// Protected branches are matched by name.
// Existing branches that differ are deleted and recreated (GitLab has no PUT).
// Only role-based access levels are copied — user/group specific ones are skipped
// as those IDs are instance-specific.
func (c *ProjectCopier) copyProtectedBranches(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_protected_branches"}

	srcBranches, err := c.src.GetProjectProtectedBranches(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source protected branches: %w", err)
		return result
	}
	dstBranches, err := c.dst.GetProjectProtectedBranches(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest protected branches: %w", err)
		return result
	}

	dstByName := make(map[string]gitlab.ProtectedBranch, len(dstBranches))
	for _, b := range dstBranches {
		dstByName[b.Name] = b
	}

	// Deduplicate source branches by name — some GitLab versions return duplicates
	seen := make(map[string]bool)
	var dedupedSrc []gitlab.ProtectedBranch
	for _, b := range srcBranches {
		if !seen[b.Name] {
			seen[b.Name] = true
			dedupedSrc = append(dedupedSrc, b)
		}
	}
	srcBranches = dedupedSrc

	sort.Slice(srcBranches, func(i, j int) bool {
		return srcBranches[i].Name < srcBranches[j].Name
	})

	for _, src := range srcBranches {
		req := gitlab.ProtectedBranchRequestFrom(src)
		dst, exists := dstByName[src.Name]

		if exists && protectedBranchMatches(src, dst) {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		action := internal.ActionCreated
		var diffs []internal.DiffLine
		if exists {
			action = internal.ActionUpdated
			diffs = protectedBranchDiffs(src, dst)
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: action,
				DryRun: true,
				Diffs:  diffs,
			})
			continue
		}

		// For existing protections, try PATCH first (preserves policy-enforced protections).
		// Fall back to delete+recreate only if PATCH fails.
		// If both fail with 403, the protection is policy-enforced and cannot be modified via API.
		if exists {
			if err := c.dst.UpdateProjectProtectedBranch(projectPath, src.Name, req); err != nil {
				if apiErr, ok := err.(*gitlab.APIError); ok && apiErr.IsForbidden() {
					// Try delete+recreate
					if delErr := c.dst.DeleteProjectProtectedBranch(projectPath, src.Name); delErr != nil {
						if apiErr2, ok := delErr.(*gitlab.APIError); ok && apiErr2.IsForbidden() {
							// Both PATCH and DELETE are 403 — protection is policy-enforced, skip gracefully
							result.Items = append(result.Items, internal.ItemResult{
								Key:    src.Name,
								Action: internal.ActionSkipped,
								Error:  fmt.Errorf("protection is enforced by a security policy — cannot be modified via API"),
							})
							continue
						}
						result.Items = append(result.Items, internal.ItemResult{
							Key:    src.Name,
							Action: internal.ActionFailed,
							Error:  fmt.Errorf("update failed (%v), delete also failed: %w", err, delErr),
						})
						continue
					}
					if err := c.dst.CreateProjectProtectedBranch(projectPath, req); err != nil {
						result.Items = append(result.Items, internal.ItemResult{
							Key:    src.Name,
							Action: internal.ActionFailed,
							Error:  err,
						})
						continue
					}
				} else {
					result.Items = append(result.Items, internal.ItemResult{
						Key:    src.Name,
						Action: internal.ActionFailed,
						Error:  err,
					})
					continue
				}
			}
			item := internal.ItemResult{Key: src.Name, Action: action, Diffs: diffs}
			if hasUserGroupAccessLevels(src) {
				item.Error = fmt.Errorf("user/group-specific access levels not copied — role-based levels only")
			}
			result.Items = append(result.Items, item)
			continue
		}

		if err := c.dst.CreateProjectProtectedBranch(projectPath, req); err != nil {
			if apiErr, ok := err.(*gitlab.APIError); ok && apiErr.IsForbidden() {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.Name,
					Action: internal.ActionSkipped,
					Error:  fmt.Errorf("protection is enforced by a security policy — cannot be modified via API"),
				})
			} else {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.Name,
					Action: internal.ActionFailed,
					Error:  err,
				})
			}
		} else {
			item := internal.ItemResult{Key: src.Name, Action: action, Diffs: diffs}
			if hasUserGroupAccessLevels(src) {
				item.Error = fmt.Errorf("user/group-specific access levels not copied — role-based levels only")
			}
			result.Items = append(result.Items, item)
		}
	}

	return result
}

// --- project_protected_tags ---

func (c *ProjectCopier) copyProtectedTags(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_protected_tags"}

	srcTags, err := c.src.GetProjectProtectedTags(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source protected tags: %w", err)
		return result
	}
	dstTags, err := c.dst.GetProjectProtectedTags(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest protected tags: %w", err)
		return result
	}

	dstByName := make(map[string]gitlab.ProtectedTag, len(dstTags))
	for _, t := range dstTags {
		dstByName[t.Name] = t
	}

	// Deduplicate source tags by name
	seenTags := make(map[string]bool)
	var dedupedSrcTags []gitlab.ProtectedTag
	for _, t := range srcTags {
		if !seenTags[t.Name] {
			seenTags[t.Name] = true
			dedupedSrcTags = append(dedupedSrcTags, t)
		}
	}
	srcTags = dedupedSrcTags

	sort.Slice(srcTags, func(i, j int) bool {
		return srcTags[i].Name < srcTags[j].Name
	})

	for _, src := range srcTags {
		req := gitlab.ProtectedTagRequestFrom(src)
		dst, exists := dstByName[src.Name]

		if exists && protectedTagMatches(src, dst) {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		action := internal.ActionCreated
		var diffs []internal.DiffLine
		if exists {
			action = internal.ActionUpdated
			diffs = protectedTagDiffs(src, dst)
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: action,
				DryRun: true,
				Diffs:  diffs,
			})
			continue
		}

		if exists {
			if err := c.dst.DeleteProjectProtectedTag(projectPath, src.Name); err != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.Name,
					Action: internal.ActionFailed,
					Error:  fmt.Errorf("deleting existing tag protection before recreate: %w", err),
				})
				continue
			}
		}

		if err := c.dst.CreateProjectProtectedTag(projectPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			item := internal.ItemResult{Key: src.Name, Action: action, Diffs: diffs}
			if hasUserGroupTagAccessLevels(src) {
				item.Error = fmt.Errorf("user/group-specific access levels not copied — role-based levels only")
			}
			result.Items = append(result.Items, item)
		}
	}

	return result
}

// --- deploy_tokens ---

func (c *ProjectCopier) copyProjectDeployTokens(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "deploy_tokens"}

	srcTokens, err := c.src.GetProjectDeployTokens(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source deploy tokens: %w", err)
		return result
	}
	dstTokens, err := c.dst.GetProjectDeployTokens(projectPath)
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
		resp, err := c.dst.CreateProjectDeployToken(projectPath, req)
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

func (c *ProjectCopier) copyProjectAccessTokens(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "access_tokens"}

	srcTokens, err := c.src.GetProjectAccessTokens(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source access tokens: %w", err)
		return result
	}
	dstTokens, err := c.dst.GetProjectAccessTokens(projectPath)
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
		resp, err := c.dst.CreateProjectAccessToken(projectPath, req)
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

// --- pipeline_schedules ---

func (c *ProjectCopier) copyPipelineSchedules(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "pipeline_schedules"}

	srcSchedules, err := c.src.GetProjectPipelineSchedules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source pipeline schedules: %w", err)
		return result
	}
	if len(srcSchedules) == 0 {
		return result
	}

	dstSchedules, err := c.dst.GetProjectPipelineSchedules(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest pipeline schedules: %w", err)
		return result
	}

	// Match by description + cron + ref — natural key for schedules
	type scheduleKey struct {
		description string
		cron        string
		ref         string
	}
	dstByKey := make(map[scheduleKey]gitlab.PipelineSchedule)
	for _, s := range dstSchedules {
		dstByKey[scheduleKey{s.Description, s.Cron, s.Ref}] = s
	}

	sort.Slice(srcSchedules, func(i, j int) bool {
		return srcSchedules[i].Description < srcSchedules[j].Description
	})

	for _, sched := range srcSchedules {
		key := scheduleKey{sched.Description, sched.Cron, sched.Ref}

		if dst, exists := dstByKey[key]; exists {
			// Key matches — check if mutable fields differ
			diffs := []internal.DiffLine{
				{Field: "active", Src: fmt.Sprintf("%v", sched.Active), Dst: fmt.Sprintf("%v", dst.Active), Match: sched.Active == dst.Active},
				{Field: "cron_timezone", Src: sched.CronTimezone, Dst: dst.CronTimezone, Match: sched.CronTimezone == dst.CronTimezone},
			}

			needsUpdate := false
			for _, d := range diffs {
				if !d.Match {
					needsUpdate = true
					break
				}
			}

			if !needsUpdate {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    sched.Description,
					Action: internal.ActionSkipped,
					Diffs:  diffs,
					DryRun: c.dryRun,
				})
				continue
			}

			if c.dryRun {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    sched.Description,
					Action: internal.ActionUpdated,
					Diffs:  diffs,
					DryRun: true,
				})
				continue
			}

			if err := c.dst.UpdateProjectPipelineSchedule(projectPath, dst.ID, gitlab.PipelineScheduleRequest{
				Description:  sched.Description,
				Ref:          sched.Ref,
				Cron:         sched.Cron,
				CronTimezone: sched.CronTimezone,
				Active:       sched.Active,
			}); err != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    sched.Description,
					Action: internal.ActionFailed,
					Error:  err,
					Diffs:  diffs,
				})
			} else {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    sched.Description,
					Action: internal.ActionUpdated,
					Diffs:  diffs,
				})
			}
			continue
		}

		// Fetch variables for this schedule from source
		srcVars, err := c.src.GetPipelineScheduleVariables(projectPath, sched.ID)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    sched.Description,
				Action: internal.ActionFailed,
				Error:  fmt.Errorf("fetching source schedule variables: %w", err),
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    sched.Description,
				Action: internal.ActionCreated,
				DryRun: true,
				Error:  fmt.Errorf("owner will be the dest token user — transfer ownership manually if needed"),
			})
			continue
		}

		// Create the schedule on dest
		newID, err := c.dst.CreateProjectPipelineSchedule(projectPath, gitlab.PipelineScheduleRequest{
			Description:  sched.Description,
			Ref:          sched.Ref,
			Cron:         sched.Cron,
			CronTimezone: sched.CronTimezone,
			Active:       sched.Active,
		})
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    sched.Description,
				Action: internal.ActionFailed,
				Error:  err,
			})
			continue
		}

		// Copy schedule variables
		var varErrors []string
		for _, v := range srcVars {
			if err := c.dst.CreatePipelineScheduleVariable(projectPath, newID, gitlab.PipelineScheduleVariableRequest{
				Key:          v.Key,
				Value:        v.Value,
				VariableType: v.VariableType,
			}); err != nil {
				varErrors = append(varErrors, fmt.Sprintf("%s: %v", v.Key, err))
			}
		}

		item := internal.ItemResult{
			Key:    sched.Description,
			Action: internal.ActionCreated,
			Error:  fmt.Errorf("owner is dest token user — transfer ownership manually if needed"),
		}
		if len(varErrors) > 0 {
			item.Error = fmt.Errorf("owner is dest token user — transfer manually; variable errors: %s", strings.Join(varErrors, ", "))
		}
		result.Items = append(result.Items, item)
	}

	return result
}

// --- helpers ---

// topicsEqual returns true if two topic slices contain the same elements
// regardless of order.
func topicsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]bool, len(a))
	for _, t := range a {
		set[t] = true
	}
	for _, t := range b {
		if !set[t] {
			return false
		}
	}
	return true
}

func protectedBranchMatches(src, dst gitlab.ProtectedBranch) bool {
	if src.AllowForcePush != dst.AllowForcePush ||
		src.CodeOwnerApprovalRequired != dst.CodeOwnerApprovalRequired {
		return false
	}
	return accessLevelsMatch(src.PushAccessLevels, dst.PushAccessLevels) &&
		accessLevelsMatch(src.MergeAccessLevels, dst.MergeAccessLevels) &&
		accessLevelsMatch(src.UnprotectAccessLevels, dst.UnprotectAccessLevels)
}

func protectedTagMatches(src, dst gitlab.ProtectedTag) bool {
	return accessLevelsMatch(src.CreateAccessLevels, dst.CreateAccessLevels)
}

// accessLevelsMatch compares role-based access levels only.
func accessLevelsMatch(src, dst []gitlab.BranchAccessLevel) bool {
	srcLevels := roleBasedLevels(src)
	dstLevels := roleBasedLevels(dst)
	if len(srcLevels) != len(dstLevels) {
		return false
	}
	srcSet := make(map[int]bool, len(srcLevels))
	for _, l := range srcLevels {
		srcSet[l] = true
	}
	for _, l := range dstLevels {
		if !srcSet[l] {
			return false
		}
	}
	return true
}

func roleBasedLevels(levels []gitlab.BranchAccessLevel) []int {
	var result []int
	for _, l := range levels {
		if l.IsRoleBased() {
			result = append(result, l.AccessLevel)
		}
	}
	return result
}

// --- project_hooks ---

func projectHookRequest(h gitlab.ProjectHook) gitlab.ProjectHookRequest {
	return gitlab.ProjectHookRequest{
		URL:                       h.URL,
		Name:                      h.Name,
		Description:               h.Description,
		PushEvents:                h.PushEvents,
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
		MemberEvents:              h.MemberEvents,
		FeatureFlagEvents:         h.FeatureFlagEvents,
		EnableSSLVerification:     h.EnableSSLVerification,
		ResourceAccessTokenEvents: h.ResourceAccessTokenEvents,
		EmojiEvents:               h.EmojiEvents,
	}
}

func projectHooksMatch(src, dst gitlab.ProjectHook) bool {
	return src.Name == dst.Name &&
		src.Description == dst.Description &&
		src.PushEvents == dst.PushEvents &&
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
		src.MemberEvents == dst.MemberEvents &&
		src.FeatureFlagEvents == dst.FeatureFlagEvents &&
		src.EnableSSLVerification == dst.EnableSSLVerification &&
		src.ResourceAccessTokenEvents == dst.ResourceAccessTokenEvents &&
		src.EmojiEvents == dst.EmojiEvents
}

func (c *ProjectCopier) copyProjectHooks(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "project_hooks"}

	srcHooks, err := c.src.GetProjectHooks(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source hooks: %w", err)
		return result
	}
	if len(srcHooks) == 0 {
		return result
	}

	dstHooks, err := c.dst.GetProjectHooks(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest hooks: %w", err)
		return result
	}

	dstByURL := make(map[string]gitlab.ProjectHook, len(dstHooks))
	for _, h := range dstHooks {
		dstByURL[h.URL] = h
	}

	tokenWarning := fmt.Errorf("webhook token cannot be copied — set token manually on dest if required")

	for _, src := range srcHooks {
		req := projectHookRequest(src)

		if dst, exists := dstByURL[src.URL]; exists {
			if projectHooksMatch(src, dst) {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionSkipped,
					DryRun: c.dryRun,
				})
				continue
			}

			if c.dryRun {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.URL,
					Action: internal.ActionUpdated,
					DryRun: true,
					Error:  tokenWarning,
				})
				continue
			}

			if err := c.dst.UpdateProjectHook(projectPath, dst.ID, req); err != nil {
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

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.URL,
				Action: internal.ActionCreated,
				DryRun: true,
				Error:  tokenWarning,
			})
			continue
		}

		if err := c.dst.CreateProjectHook(projectPath, req); err != nil {
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

func hasUserGroupAccessLevels(b gitlab.ProtectedBranch) bool {
	for _, al := range append(append(b.PushAccessLevels, b.MergeAccessLevels...), b.UnprotectAccessLevels...) {
		if !al.IsRoleBased() {
			return true
		}
	}
	return false
}

func hasUserGroupTagAccessLevels(t gitlab.ProtectedTag) bool {
	for _, al := range t.CreateAccessLevels {
		if !al.IsRoleBased() {
			return true
		}
	}
	return false
}

// --- job_token_scope ---

func (c *ProjectCopier) copyJobTokenScope(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "job_token_scope"}

	// Fetch source and dest project IDs
	srcProject, err := c.src.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source project: %w", err)
		return result
	}
	dstProject, err := c.dst.GetProject(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest project: %w", err)
		return result
	}

	// --- Layer 1: inbound_enabled toggle ---
	srcScope, err := c.src.GetJobTokenScope(srcProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching source job token scope: %w", err)
		return result
	}
	if srcScope == nil {
		return result
	}

	dstScope, err := c.dst.GetJobTokenScope(dstProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest job token scope: %w", err)
		return result
	}

	diffs := []internal.DiffLine{
		{
			Field: "inbound_enabled",
			Src:   fmt.Sprintf("%v", srcScope.InboundEnabled),
			Dst: fmt.Sprintf("%v", func() bool {
				if dstScope != nil {
					return dstScope.InboundEnabled
				}
				return false
			}()),
			Match: dstScope != nil && srcScope.InboundEnabled == dstScope.InboundEnabled,
		},
		{
			Field: "ci_push_repository_for_job_token_allowed",
			Src:   fmt.Sprintf("%v", srcScope.CiPushRepositoryForJobTokenAllowed),
			Dst: fmt.Sprintf("%v", func() bool {
				if dstScope != nil {
					return dstScope.CiPushRepositoryForJobTokenAllowed
				}
				return false
			}()),
			Match: dstScope != nil && srcScope.CiPushRepositoryForJobTokenAllowed == dstScope.CiPushRepositoryForJobTokenAllowed,
		},
	}

	scopeChanged := dstScope == nil || srcScope.InboundEnabled != dstScope.InboundEnabled
	pushChanged := dstScope == nil || srcScope.CiPushRepositoryForJobTokenAllowed != dstScope.CiPushRepositoryForJobTokenAllowed

	if scopeChanged || pushChanged {
		if !c.dryRun {
			var applyErr error
			if scopeChanged {
				applyErr = c.dst.SetJobTokenScope(dstProject.ID, srcScope.InboundEnabled)
			}
			if applyErr == nil && pushChanged {
				applyErr = c.dst.SetJobTokenPushPermission(dstProject.ID, srcScope.CiPushRepositoryForJobTokenAllowed)
			}
			if applyErr != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    "job_token_permissions",
					Action: internal.ActionFailed,
					Error:  applyErr,
				})
			} else {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    "job_token_permissions",
					Action: internal.ActionUpdated,
					Diffs:  diffs,
				})
			}
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "job_token_permissions",
				Action: internal.ActionUpdated,
				Diffs:  diffs,
				DryRun: true,
			})
		}
	} else {
		result.Items = append(result.Items, internal.ItemResult{
			Key:    "job_token_permissions",
			Action: internal.ActionSkipped,
			Diffs:  diffs,
			DryRun: c.dryRun,
		})
	}

	// --- Layer 2: project allowlist ---
	srcProjects, err := c.src.GetJobTokenAllowlist(srcProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching source project allowlist: %w", err)
		return result
	}

	dstProjects, err := c.dst.GetJobTokenAllowlist(dstProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest project allowlist: %w", err)
		return result
	}

	// Build dest allowlist set by path
	dstProjectPaths := make(map[string]bool)
	for _, p := range dstProjects {
		dstProjectPaths[p.PathWithNamespace] = true
	}

	for _, sp := range srcProjects {
		// Skip the project itself — always in its own allowlist by default
		if sp.PathWithNamespace == projectPath {
			continue
		}

		if dstProjectPaths[sp.PathWithNamespace] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "project: " + sp.PathWithNamespace,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "project: " + sp.PathWithNamespace,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		// Look up the project on dest to get its numeric ID
		dstTargetProject, err := c.dst.GetProject(sp.PathWithNamespace)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "project: " + sp.PathWithNamespace,
				Action: internal.ActionFailed,
				Error:  fmt.Errorf("project not found on dest: %w", err),
			})
			continue
		}

		if err := c.dst.AddJobTokenAllowlistProject(dstProject.ID, dstTargetProject.ID); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "project: " + sp.PathWithNamespace,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "project: " + sp.PathWithNamespace,
				Action: internal.ActionCreated,
			})
		}
	}

	// --- Layer 3: group allowlist ---
	srcGroups, err := c.src.GetJobTokenGroupsAllowlist(srcProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching source groups allowlist: %w", err)
		return result
	}

	dstGroups, err := c.dst.GetJobTokenGroupsAllowlist(dstProject.ID)
	if err != nil {
		result.Error = fmt.Errorf("fetching dest groups allowlist: %w", err)
		return result
	}

	dstGroupPaths := make(map[string]bool)
	for _, g := range dstGroups {
		dstGroupDetails, err := c.dst.GetGroup(fmt.Sprintf("%d", g.ID))
		if err == nil && dstGroupDetails != nil {
			dstGroupPaths[dstGroupDetails.FullPath] = true
		}
	}

	for _, sg := range srcGroups {
		// The allowlist API only returns id/name — look up full_path via source
		srcGroupDetails, err := c.src.GetGroup(fmt.Sprintf("%d", sg.ID))
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    fmt.Sprintf("group: (id %d)", sg.ID),
				Action: internal.ActionFailed,
				Error:  fmt.Errorf("fetching source group details: %w", err),
			})
			continue
		}
		groupPath := srcGroupDetails.FullPath

		if dstGroupPaths[groupPath] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "group: " + groupPath,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
			continue
		}

		if c.dryRun {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "group: " + groupPath,
				Action: internal.ActionCreated,
				DryRun: true,
			})
			continue
		}

		dstTargetGroup, err := c.dst.GetGroup(groupPath)
		if err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "group: " + groupPath,
				Action: internal.ActionFailed,
				Error:  fmt.Errorf("group not found on dest: %w", err),
			})
			continue
		}

		if err := c.dst.AddJobTokenAllowlistGroup(dstProject.ID, dstTargetGroup.ID); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "group: " + groupPath,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    "group: " + groupPath,
				Action: internal.ActionCreated,
			})
		}
	}

	return result
}
