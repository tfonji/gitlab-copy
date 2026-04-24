package copy

import (
	"fmt"
	"sort"

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

	// Credentials (password/token) are masked in API responses and won't
	// transfer — flag for manual follow-up
	result.Items = []internal.ItemResult{
		{
			Key:    "jira_integration",
			Action: action,
			Error:  fmt.Errorf("credentials not copied — verify password/token on dest manually"),
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
		if err := c.dst.CreateProjectTrigger(projectPath, req); err != nil {
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionFailed,
				Error:  err,
			})
		} else {
			// Token is auto-generated on dest — cannot be copied from source
			result.Items = append(result.Items, internal.ItemResult{
				Key:    trigger.Description,
				Action: internal.ActionCreated,
				Error:  fmt.Errorf("trigger token is newly generated — update any CI variables referencing the source token"),
			})
		}
	}

	return result
}

// --- deploy_keys ---

func (c *ProjectCopier) copyDeployKeys(projectPath string) internal.DomainCopyResult {
	result := internal.DomainCopyResult{Domain: "deploy_keys"}

	srcKeys, err := c.src.GetProjectDeployKeys(projectPath)
	if err != nil {
		result.Error = fmt.Errorf("fetching source deploy keys: %w", err)
		return result
	}
	dstKeys, err := c.dst.GetProjectDeployKeys(projectPath)
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
		if err := c.dst.CreateProjectDeployKey(projectPath, req); err != nil {
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
