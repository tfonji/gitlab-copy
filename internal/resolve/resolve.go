package resolve

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"gitlab-copy/internal/config"
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
	appIDsRaw := os.Getenv("APP_IDS")
	if strings.TrimSpace(appIDsRaw) == "" {
		fmt.Fprintln(w, "APP_IDS not set or empty — skipping resolve, using existing config")
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

	// --- Load and update config ---
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg.Groups.Include = groupsToAdd
	cfg.Groups.IncludeSubgroups = false
	cfg.Projects.Include = projects
	cfg.Projects.IncludeSubgroups = false

	// --- Write updated config ---
	if err := writeConfig(configPath, cfg); err != nil {
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

func writeConfig(path string, cfg *config.Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(cfg)
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

	commitMsg := fmt.Sprintf("chore: resolve config for APPID(s) %s [skip ci]", strings.Join(appIDs, ", "))

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
