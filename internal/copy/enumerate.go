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

type GroupEntry struct {
	GroupPath string
}

func EnumerateGroups(cfg *config.Config, srcClient *gitlab.Client, allGroups bool) ([]GroupEntry, error) {
	seen := make(map[string]bool)
	var entries []GroupEntry

	var topLevel []string

	if allGroups {
		// Fetch all top-level groups from source instance
		groups, err := srcClient.ListAllTopLevelGroups()
		if err != nil {
			return nil, fmt.Errorf("listing all top-level groups: %w", err)
		}
		for _, g := range groups {
			topLevel = append(topLevel, g.FullPath)
		}
	} else {
		topLevel = cfg.Groups.Include
	}

	for _, groupPath := range topLevel {
		if !seen[groupPath] {
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

func EnumerateProjects(cfg *config.Config, srcClient *gitlab.Client, allGroups bool) ([]ProjectEntry, error) {
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
		var topLevel []string
		if allGroups {
			groups, err := srcClient.ListAllTopLevelGroups()
			if err != nil {
				return nil, fmt.Errorf("listing all top-level groups: %w", err)
			}
			for _, g := range groups {
				topLevel = append(topLevel, g.FullPath)
			}
		} else {
			topLevel = cfg.Groups.Include
		}

		seen := make(map[string]bool)
		for _, groupPath := range topLevel {
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
func isExcluded(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == path {
			return true
		}
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
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

func projectDepth(projectPath, groupPath string) int {
	relative := strings.TrimPrefix(projectPath, groupPath+"/")
	return strings.Count(relative, "/")
}
