package resolve

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"gitlab-copy/internal/gitlab"

	"gopkg.in/yaml.v3"
)

// Result holds the output of a resolve run.
type Result struct {
	AppIDs        []string
	Projects      []string // full paths of matched projects
	GroupsAdded   []string // groups not on dest — added to config
	GroupsSkipped []string // groups already on dest — skipped
}

// Run resolves APP_IDS to a concrete project and group list, updates
// config.yaml in place, and commits+pushes if running in CI.
func Run(configPath string, srcClient, dstClient *gitlab.Client, w io.Writer) error {
	// --- Always update domain flags based on pipeline variables ---
	enforceSecurityPolicy := strings.EqualFold(os.Getenv("ENFORCE_SECURITY_POLICY"), "yes")
	copyCompliance := strings.EqualFold(os.Getenv("COPY_COMPLIANCE_FRAMEWORKS"), "yes")
	linkMRPolicy := !strings.EqualFold(os.Getenv("LINK_MERGE_REQUEST_POLICY"), "no")

	groupDomains := buildGroupDomains(enforceSecurityPolicy, copyCompliance, linkMRPolicy)

	fmt.Fprintf(w, "Domain flags:\n")
	fmt.Fprintf(w, "  ENFORCE_SECURITY_POLICY:     %v\n", enforceSecurityPolicy)
	fmt.Fprintf(w, "  COPY_COMPLIANCE_FRAMEWORKS:  %v\n", copyCompliance)
	fmt.Fprintf(w, "  LINK_MERGE_REQUEST_POLICY:   %v\n", linkMRPolicy)
	fmt.Fprintf(w, "  domains.groups will be set to: %s\n", strings.Join(groupDomains, ", "))

	appIDsRaw := os.Getenv("APP_IDS")
	if strings.TrimSpace(appIDsRaw) == "" {
		// No APP_IDS — skip project resolution but still update domain flags
		fmt.Fprintln(w, "\nAPP_IDS not set or empty — skipping project resolve, updating domain flags only")
		if err := updateDomainFlags(configPath, groupDomains); err != nil {
			return fmt.Errorf("updating domain flags: %w", err)
		}
		return nil
	}

	appIDs := parseAppIDs(appIDsRaw)
	fmt.Fprintf(w, "Resolving %d APPID(s): %s\n", len(appIDs), strings.Join(appIDs, ", "))

	// --- Search source for projects by topic ---
	seen := make(map[string]bool)
	var projects []string

	for _, appID := range appIDs {
		results, err := srcClient.FindProjectsByTopic(appID)
		if err != nil {
			return fmt.Errorf("searching for APPID %q: %w", appID, err)
		}
		for _, p := range results {
			if !seen[p.PathWithNamespace] {
				seen[p.PathWithNamespace] = true
				projects = append(projects, p.PathWithNamespace)
			}
		}
	}

	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects found matching the provided APPID(s)")
		return nil
	}
	sort.Strings(projects)
	fmt.Fprintf(w, "Found %d unique project(s)\n", len(projects))

	// --- Collect unique ancestor groups ---
	groupSet := make(map[string]bool)
	for _, p := range projects {
		for _, g := range ancestorGroups(p) {
			groupSet[g] = true
		}
	}

	// --- Check each group against dest ---
	var groupsToAdd []string
	var groupsSkipped []string

	allGroups := sortedKeys(groupSet)
	for _, g := range allGroups {
		exists, err := dstClient.GroupExists(g)
		if err != nil {
			return fmt.Errorf("checking dest group %q: %w", g, err)
		}
		if exists {
			groupsSkipped = append(groupsSkipped, g)
		} else {
			groupsToAdd = append(groupsToAdd, g)
		}
	}

	// --- Print summary ---
	fmt.Fprintf(w, "\nGroups not on dest (will copy settings): ")
	if len(groupsToAdd) == 0 {
		fmt.Fprintln(w, "none")
	} else {
		fmt.Fprintln(w, strings.Join(groupsToAdd, ", "))
	}

	fmt.Fprintf(w, "Groups already on dest (skipped):        ")
	if len(groupsSkipped) == 0 {
		fmt.Fprintln(w, "none")
	} else {
		fmt.Fprintln(w, strings.Join(groupsSkipped, ", "))
	}

	// --- Write updated config (groups, projects, and domains sections) ---
	if err := writeConfig(configPath, groupsToAdd, projects, groupDomains); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	fmt.Fprintf(w, "\nUpdated %s\n", configPath)
	fmt.Fprintf(w, "  groups.include:   %d group(s)\n", len(groupsToAdd))
	fmt.Fprintf(w, "  projects.include: %d project(s)\n", len(projects))

	// --- Commit and push if in CI ---
	if os.Getenv("CI") == "true" {
		if err := commitAndPush(configPath, appIDs, w); err != nil {
			return fmt.Errorf("committing config: %w", err)
		}
	} else {
		fmt.Fprintln(w, "\nNot running in CI — config updated locally, no git push")
	}

	return nil
}

// ancestorGroups returns all ancestor group paths for a project path.
// e.g. "a/b/c/project" → ["a", "a/b", "a/b/c"]
func ancestorGroups(projectPath string) []string {
	parts := strings.Split(projectPath, "/")
	// last part is the project name, everything before is groups
	if len(parts) < 2 {
		return nil
	}
	groupParts := parts[:len(parts)-1]
	var ancestors []string
	for i := range groupParts {
		ancestors = append(ancestors, strings.Join(groupParts[:i+1], "/"))
	}
	return ancestors
}

func parseAppIDs(raw string) []string {
	var ids []string
	for _, id := range strings.Split(raw, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// updateDomainFlags updates only the domains.groups section in config.yaml,
// leaving groups, projects, and all other sections untouched.
func updateDomainFlags(path string, groupDomains []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	if len(root.Content) == 0 {
		return fmt.Errorf("empty config file")
	}
	updateNestedSection(root.Content[0], "domains", "groups", groupDomains)
	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}

// buildGroupDomains returns the groups domain list based on pipeline flags.
// compliance_frameworks, compliance_assignments, and enforce_security_policy
// are opt-in — excluded by default.
func buildGroupDomains(enforceSecurityPolicy, copyCompliance, linkMRPolicy bool) []string {
	base := []string{
		"push_rules",
		"description",
		"default_branch_name",
		"default_branch_protection",
		"mr_settings",
		"mr_approval_settings",
		"protected_environments",
		"approval_rules",
		"jira_integration",
		"badges",
		"security_policy_project",
		"deploy_tokens",
		"access_tokens",
	}

	if copyCompliance {
		base = append(base, "compliance_frameworks", "compliance_assignments")
	}
	if enforceSecurityPolicy {
		base = append(base, "enforce_security_policy")
	}
	if linkMRPolicy {
		base = append(base, "link_merge_request_policy")
	}

	return base
}

func writeConfig(path string, groupsInclude []string, projectsInclude []string, groupDomains []string) error {
	// Read the raw file as a YAML node tree so we preserve all other
	// fields, comments, and ordering — only groups and projects are touched.
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// The root is a document node; the actual mapping is root.Content[0]
	if len(root.Content) == 0 {
		return fmt.Errorf("empty config file")
	}
	doc := root.Content[0]

	// Always update both sections to ensure config is in a clean known state.
	// Leftover values from a previous batch would cause unintended processing.
	// Empty slice = empty list in YAML, which is valid and intentional.
	updateSection(doc, "groups", map[string]any{
		"include":           groupsInclude,
		"include_subgroups": false,
	})
	updateSection(doc, "projects", map[string]any{
		"include":           projectsInclude,
		"include_subgroups": false,
	})

	// Update domains.groups to reflect the pipeline flag selections
	updateNestedSection(doc, "domains", "groups", groupDomains)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	return os.WriteFile(path, out, 0644)
}

// updateNestedSection finds a top-level key, then updates a sub-key within it.
// e.g. updateNestedSection(doc, "domains", "groups", [...]) sets domains.groups.
func updateNestedSection(mapping *yaml.Node, section, key string, val any) {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value != section {
			continue
		}
		valNode := mapping.Content[i+1]
		if valNode.Kind != yaml.MappingNode {
			continue
		}
		setMappingValue(valNode, key, val)
		return
	}
	// Section not found — append it
	sectionNode := &yaml.Node{Kind: yaml.MappingNode}
	setMappingValue(sectionNode, key, val)
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: section},
		sectionNode,
	)
}

// updateSection finds a mapping key in a YAML mapping node and updates
// only the specified sub-keys, leaving all other sub-keys untouched.
func updateSection(mapping *yaml.Node, section string, updates map[string]any) {
	// Find the section key in the top-level mapping
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		keyNode := mapping.Content[i]
		valNode := mapping.Content[i+1]
		if keyNode.Value != section {
			continue
		}
		if valNode.Kind != yaml.MappingNode {
			continue
		}
		// Update each key within the section
		for updateKey, updateVal := range updates {
			setMappingValue(valNode, updateKey, updateVal)
		}
		return
	}

	// Section not found — append it
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: section}
	valNode := &yaml.Node{Kind: yaml.MappingNode}
	for updateKey, updateVal := range updates {
		setMappingValue(valNode, updateKey, updateVal)
	}
	mapping.Content = append(mapping.Content, keyNode, valNode)
}

// setMappingValue sets or replaces a key in a YAML mapping node.
func setMappingValue(mapping *yaml.Node, key string, val any) {
	newVal := toYAMLNode(val)

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = newVal
			return
		}
	}
	// Key not found — append
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		newVal,
	)
}

// toYAMLNode converts a Go value to a yaml.Node.
func toYAMLNode(val any) *yaml.Node {
	switch v := val.(type) {
	case bool:
		s := "false"
		if v {
			s = "true"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Value: s, Tag: "!!bool"}
	case []string:
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, s := range v {
			seq.Content = append(seq.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: s,
				Tag:   "!!str",
			})
		}
		return seq
	default:
		node := &yaml.Node{}
		_ = node.Encode(val)
		return node
	}
}

func commitAndPush(configPath string, appIDs []string, w io.Writer) error {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITLAB_TOKEN not set — cannot push config update")
	}

	ciURL := os.Getenv("CI_REPOSITORY_URL")
	projectURL := os.Getenv("CI_PROJECT_URL")
	if ciURL == "" || projectURL == "" {
		return fmt.Errorf("CI_REPOSITORY_URL or CI_PROJECT_URL not set")
	}

	// Build authenticated push URL
	// CI_REPOSITORY_URL is like https://gitlab-ci-token:xxx@gitlab.example.com/group/repo.git
	// Replace the token with GITLAB_TOKEN
	parts := strings.SplitN(ciURL, "@", 2)
	if len(parts) != 2 {
		return fmt.Errorf("unexpected CI_REPOSITORY_URL format")
	}
	pushURL := fmt.Sprintf("https://oauth2:%s@%s", token, parts[1])

	branch := os.Getenv("CI_COMMIT_REF_NAME")
	if branch == "" {
		branch = "main"
	}

	commitMsg := fmt.Sprintf("chore: resolve config for APPID(s) %s", strings.Join(appIDs, ", "))

	cmds := [][]string{
		{"git", "config", "user.email", "gitlab-ci@migration"},
		{"git", "config", "user.name", "GitLab Migration Bot"},
		{"git", "add", configPath},
		{"git", "commit", "-m", commitMsg},
		{"git", "push", pushURL, fmt.Sprintf("HEAD:%s", branch)},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			// If nothing to commit, git commit returns non-zero — treat as OK
			if args[1] == "commit" && strings.Contains(err.Error(), "exit status 1") {
				fmt.Fprintln(w, "Nothing to commit — config unchanged")
				return nil
			}
			return fmt.Errorf("git %s failed: %w", args[1], err)
		}
	}

	fmt.Fprintf(w, "Committed and pushed %s to %s\n", configPath, branch)
	return nil
}
