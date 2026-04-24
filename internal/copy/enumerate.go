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

func isExcluded(projectPath string, patterns []string) bool {
	for _, pattern := range patterns {

		if pattern == projectPath {
			return true
		}

		matched, err := filepath.Match(pattern, projectPath)
		if err == nil && matched {
			return true
		}

		parts := strings.Split(projectPath, "/")
		if len(parts) > 0 {
			matched, err = filepath.Match(pattern, parts[len(parts)-1])
			if err == nil && matched {
				return true
			}
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
