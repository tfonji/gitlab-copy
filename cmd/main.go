package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"

	"gitlab-copy/internal"
	"gitlab-copy/internal/config"
	"gitlab-copy/internal/copy"
	"gitlab-copy/internal/gitlab"
	"gitlab-copy/internal/report"
)

const usage = `gitlab-copy — copy GitLab settings from source to destination instance

Usage:
  gitlab-copy <command> [flags]

Commands:
  groups all          Copy all group-level domains
  projects all        Copy all project-level domains
  all                 Copy everything (groups + projects)

Single-target flags:
  -group      Copy a single group path only
  -project    Copy a single project path only

Other flags:
  -config     Path to config file (default: config.yaml)
  -dry-run    Preview what would be copied without making any changes
  -no-color   Disable terminal color output

Examples:
  gitlab-copy all -config config.yaml -dry-run
  gitlab-copy groups all -config config.yaml
  gitlab-copy projects all -config config.yaml
  gitlab-copy all -config config.yaml -group my-group/subgroup
  gitlab-copy all -config config.yaml -project my-group/my-project
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	first := os.Args[1]
	if first == "-h" || first == "--help" || first == "help" {
		fmt.Print(usage)
		os.Exit(0)
	}

	var subject, verb string
	var flagArgs []string

	switch first {
	case "groups", "projects":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: gitlab-copy %s <all> [flags]\n", first)
			os.Exit(1)
		}
		subject = first
		verb = os.Args[2]
		flagArgs = os.Args[3:]
	case "all":
		subject = "all"
		verb = first
		flagArgs = os.Args[2:]
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", first, usage)
		os.Exit(1)
	}

	if verb != "all" {
		fmt.Fprintf(os.Stderr, "unknown verb %q — expected all\n", verb)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("gitlab-copy", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "path to config file")
	singleGroup := fs.String("group", "", "copy a single group path")
	singleProj := fs.String("project", "", "copy a single project path")
	dryRun := fs.Bool("dry-run", false, "preview without making changes")
	noColor := fs.Bool("no-color", false, "disable color output")
	if err := fs.Parse(flagArgs); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		os.Exit(1)
	}

	scope := "all"
	if subject == "groups" {
		scope = "groups"
	} else if subject == "projects" {
		scope = "projects"
	}
	if *singleProj != "" {
		scope = "projects"
	}

	cfg, err := config.LoadWithOverrides(*configPath, *singleGroup, *singleProj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Fprintf(os.Stderr, "dry-run mode — no changes will be made\n")
	}

	srcClient := gitlab.NewClient(cfg.Source.URL, cfg.Source.Token())
	dstClient := gitlab.NewClient(cfg.Destination.URL, cfg.Destination.Token())

	result := run(scope, *dryRun, cfg, srcClient, dstClient)

	useColor := !*noColor
	term := report.NewTerminal(os.Stdout, useColor)
	term.Write(result)

	for _, format := range cfg.Output.Formats {
		switch format {
		case "json":
			path, err := report.WriteJSON(result, cfg.Output.Dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON report: %v\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "\nJSON report: %s\n", path)
			}
		case "html":
			path, err := report.WriteHTML(result, cfg.Output.Dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing HTML report: %v\n", err)
			} else {
				fmt.Fprintf(os.Stdout, "HTML report: %s\n", path)
			}
		}
	}

	if result.HasFailures {
		os.Exit(1)
	}
}

func run(scope string, dryRun bool, cfg *config.Config, src, dst *gitlab.Client) *internal.RunResult {
	runGroups := scope == "groups" || scope == "all"
	runProjects := scope == "projects" || scope == "all"

	result := &internal.RunResult{DryRun: dryRun}

	if runGroups {
		groupCopier := copy.NewGroupCopier(src, dst, cfg.Domains.Groups, dryRun)

		groups, err := copy.EnumerateGroups(cfg, src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error enumerating groups: %v\n", err)
			os.Exit(1)
		}
		if len(groups) == 0 {
			fmt.Fprintf(os.Stderr, "warning: no groups found\n")
		} else {
			fmt.Fprintf(os.Stderr, "processing %d group(s)...\n", len(groups))
		}

		groupPaths := make([]string, len(groups))
		for i, g := range groups {
			groupPaths[i] = g.GroupPath
		}

		result.Groups = runGroupCopies(groupPaths, cfg.Concurrency.Groups, groupCopier)
	}

	if runProjects {
		projCopier := copy.NewProjectCopier(src, dst, cfg.Domains.Projects, dryRun)

		projects, err := copy.EnumerateProjects(cfg, src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error enumerating projects: %v\n", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Fprintf(os.Stderr, "warning: no projects found\n")
		} else {
			fmt.Fprintf(os.Stderr, "processing %d project(s)...\n", len(projects))
		}

		projectResults := runProjectCopies(projects, cfg.Concurrency.Projects, projCopier)
		result.ProjectGroups = groupProjectResults(projectResults)
	}

	result.HasFailures = computeHasFailures(result)
	return result
}

func runGroupCopies(groups []string, concurrency int, copier *copy.GroupCopier) []internal.GroupCopyResult {
	jobs := make(chan string, len(groups))
	results := make(chan internal.GroupCopyResult, len(groups))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for groupPath := range jobs {
				results <- internal.GroupCopyResult{
					GroupPath: groupPath,
					Domains:   copier.Copy(groupPath),
				}
			}
		}()
	}

	for _, g := range groups {
		jobs <- g
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	order := make(map[string]int, len(groups))
	for i, g := range groups {
		order[g] = i
	}
	ordered := make([]internal.GroupCopyResult, len(groups))
	for r := range results {
		ordered[order[r.GroupPath]] = r
	}
	return ordered
}

func runProjectCopies(projects []copy.ProjectEntry, concurrency int, copier *copy.ProjectCopier) []internal.ProjectCopyResult {
	jobs := make(chan copy.ProjectEntry, len(projects))
	results := make(chan internal.ProjectCopyResult, len(projects))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				results <- internal.ProjectCopyResult{
					ProjectPath: entry.ProjectPath,
					GroupPath:   entry.GroupPath,
					Domains:     copier.Copy(entry.ProjectPath),
				}
			}
		}()
	}

	for _, p := range projects {
		jobs <- p
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	order := make(map[string]int, len(projects))
	for i, p := range projects {
		order[p.ProjectPath] = i
	}
	ordered := make([]internal.ProjectCopyResult, len(projects))
	for r := range results {
		ordered[order[r.ProjectPath]] = r
	}
	return ordered
}

func groupProjectResults(results []internal.ProjectCopyResult) []internal.GroupProjectCopyResults {
	groupMap := make(map[string][]internal.ProjectCopyResult)
	var groupOrder []string
	seen := make(map[string]bool)

	for _, r := range results {
		if !seen[r.GroupPath] {
			seen[r.GroupPath] = true
			groupOrder = append(groupOrder, r.GroupPath)
		}
		groupMap[r.GroupPath] = append(groupMap[r.GroupPath], r)
	}

	sort.Strings(groupOrder)

	grouped := make([]internal.GroupProjectCopyResults, 0, len(groupOrder))
	for _, g := range groupOrder {
		projs := groupMap[g]
		sort.Slice(projs, func(i, j int) bool {
			return projs[i].ProjectPath < projs[j].ProjectPath
		})
		grouped = append(grouped, internal.GroupProjectCopyResults{
			GroupPath: g,
			Projects:  projs,
		})
	}
	return grouped
}

func computeHasFailures(result *internal.RunResult) bool {
	for _, gr := range result.Groups {
		for _, d := range gr.Domains {
			if d.HasFailures() {
				return true
			}
		}
	}
	for _, gpg := range result.ProjectGroups {
		for _, pr := range gpg.Projects {
			for _, d := range pr.Domains {
				if d.HasFailures() {
					return true
				}
			}
		}
	}
	return false
}
