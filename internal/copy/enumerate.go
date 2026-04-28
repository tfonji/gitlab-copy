package copy

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gitlab-copy/internal/config"
	"gitlab-copy/internal/gitlab"
)

type ProjectEntry struct {
	ProjectPath string
	GroupPath   string
}

func EnumerateProjects(cfg *config.Config, srcClient *gitlab.Client) ([]ProjectEntry, error) {
	var entries []ProjectEntry

	if len(cfg.Projects.Include) > 0 {

		for _, path := range cfg.Projects.Include {
			if !isExcluded(path, cfg.Projects.Exclude) {

				entries = append(entries, ProjectEntry{
					ProjectPath: path,
					GroupPath:   groupFromProjectPath(path),
				})
			}
		}
	} else {

		seen := make(map[string]bool)
		for _, groupPath := range cfg.Groups.Include {
			projects, err := srcClient.ListGroupProjects(
				groupPath,
				cfg.Projects.IncludeSubgroups,
				cfg.Projects.IncludeArchived,
			)
			if err != nil {
				return nil, err
			}
			for _, p := range projects {
				if seen[p.PathWithNamespace] {
					continue
				}
				if isExcluded(p.PathWithNamespace, cfg.Projects.Exclude) {
					continue
				}
				// Also exclude projects whose group is excluded
				if isExcluded(groupFromProjectPath(p.PathWithNamespace), cfg.Groups.Exclude) {
					continue
				}
				// Apply max_depth — 0 means unlimited
				if cfg.Projects.MaxDepth > 0 {
					depth := projectDepth(p.PathWithNamespace, groupPath)
					if depth > cfg.Projects.MaxDepth {
						continue
					}
				}
				seen[p.PathWithNamespace] = true
				entries = append(entries, ProjectEntry{
					ProjectPath: p.PathWithNamespace,
					GroupPath:   groupFromProjectPath(p.PathWithNamespace),
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].GroupPath != entries[j].GroupPath {
			return entries[i].GroupPath < entries[j].GroupPath
		}
		return entries[i].ProjectPath < entries[j].ProjectPath
	})

	return entries, nil
}

// isExcluded returns true if path matches any of the exclusion patterns.
// Pattern matching rules:
//   - Exact match:        "my-group/subgroup"
//   - Single-level glob: "my-group/sub*"  matches "my-group/subgroup" but not "my-group/sub/child"
//   - Deep glob:         "my-group/*"     matches at any depth below my-group
//     e.g. "my-group/a", "my-group/a/b", "my-group/a/b/c"
func isExcluded(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Exact match
		if pattern == path {
			return true
		}

		// Deep glob: pattern ending in /* matches path at any depth below prefix
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}

		// Standard filepath glob for single-level patterns
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func groupFromProjectPath(projectPath string) string {
	idx := strings.LastIndex(projectPath, "/")
	if idx < 0 {
		return ""
	}
	return projectPath[:idx]
}

// projectDepth returns how many subgroup levels deep a project is relative
// to the top-level group being enumerated.
// Examples (groupPath = "fxpayments"):
//
//	fxpayments/project-a           → depth 0
//	fxpayments/dast/project-a      → depth 1
//	fxpayments/dast/scan/project-a → depth 2
func projectDepth(projectPath, groupPath string) int {
	relative := strings.TrimPrefix(projectPath, groupPath+"/")
	// depth = number of "/" in relative path (each "/" is a subgroup boundary)
	return strings.Count(relative, "/")
}

type GroupEntry struct {
	GroupPath string
}

func EnumerateGroups(cfg *config.Config, srcClient *gitlab.Client) ([]GroupEntry, error) {
	seen := make(map[string]bool)
	var entries []GroupEntry

	for _, groupPath := range cfg.Groups.Include {
		if !seen[groupPath] && !isExcluded(groupPath, cfg.Groups.Exclude) {
			seen[groupPath] = true
			entries = append(entries, GroupEntry{GroupPath: groupPath})
		}

		if cfg.Groups.IncludeSubgroups {
			subgroups, err := srcClient.ListSubgroups(groupPath)
			if err != nil {
				return nil, fmt.Errorf("listing subgroups of %s: %w", groupPath, err)
			}
			for _, sg := range subgroups {
				if seen[sg.FullPath] {
					continue
				}
				if isExcluded(sg.FullPath, cfg.Groups.Exclude) {
					continue
				}
				seen[sg.FullPath] = true
				entries = append(entries, GroupEntry{GroupPath: sg.FullPath})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GroupPath < entries[j].GroupPath
	})

	return entries, nil
}
