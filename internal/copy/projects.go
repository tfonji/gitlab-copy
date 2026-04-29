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
	default:
		return internal.DomainCopyResult{
			Domain: domain,
			Error:  fmt.Errorf("unknown domain %q", domain),
		}
	}
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
		result.Items = []internal.ItemResult{
			{Key: "topics", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	if topicsEqual(src.Topics, dst.Topics) {
		result.Items = []internal.ItemResult{
			{Key: "topics", Action: internal.ActionSkipped, DryRun: c.dryRun},
		}
		return result
	}

	action := internal.ActionUpdated
	if len(dst.Topics) == 0 {
		action = internal.ActionCreated
	}

	if c.dryRun {
		result.Items = []internal.ItemResult{
			{Key: "topics", Action: action, DryRun: true},
		}
		return result
	}

	if err := c.dst.UpdateProject(projectPath, gitlab.ProjectUpdateRequest{
		Topics: src.Topics,
	}); err != nil {
		result.Items = []internal.ItemResult{
			{Key: "topics", Action: internal.ActionFailed, Error: err},
		}
		return result
	}

	result.Items = []internal.ItemResult{
		{Key: "topics", Action: action},
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

		// DELETE existing branch protection before recreating
		if exists {
			if err := c.dst.DeleteProjectProtectedBranch(projectPath, src.Name); err != nil {
				result.Items = append(result.Items, internal.ItemResult{
					Key:    src.Name,
					Action: internal.ActionFailed,
					Error:  fmt.Errorf("deleting existing protection before recreate: %w", err),
				})
				continue
			}
		}

		if err := c.dst.CreateProjectProtectedBranch(projectPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    src.Name,
				Action: internal.ActionFailed,
				Error:  err,
			})
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
	dstByKey := make(map[scheduleKey]bool)
	for _, s := range dstSchedules {
		dstByKey[scheduleKey{s.Description, s.Cron, s.Ref}] = true
	}

	sort.Slice(srcSchedules, func(i, j int) bool {
		return srcSchedules[i].Description < srcSchedules[j].Description
	})

	for _, sched := range srcSchedules {
		key := scheduleKey{sched.Description, sched.Cron, sched.Ref}

		if dstByKey[key] {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    sched.Description,
				Action: internal.ActionSkipped,
				DryRun: c.dryRun,
			})
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
