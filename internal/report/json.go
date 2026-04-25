package report

import (
	"encoding/json"
	"fmt"
	"gitlab-copy/internal"
	"os"
	"path/filepath"
)

func WriteJSON(result *internal.RunResult, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating output dir: %w", err)
	}
	filename := fmt.Sprintf("gitlab-copy.json")
	path := filepath.Join(dir, filename)

	// Serialize errors as strings for JSON output
	type jsonItem struct {
		Key    string `json:"key"`
		Action string `json:"action"`
		DryRun bool   `json:"dry_run"`
		Error  string `json:"error,omitempty"`
	}
	type jsonDomain struct {
		Domain string     `json:"domain"`
		Items  []jsonItem `json:"items,omitempty"`
		Error  string     `json:"error,omitempty"`
	}
	type jsonGroup struct {
		GroupPath string       `json:"group_path"`
		Domains   []jsonDomain `json:"domains"`
	}
	type jsonProject struct {
		ProjectPath string       `json:"project_path"`
		GroupPath   string       `json:"group_path"`
		Domains     []jsonDomain `json:"domains"`
	}
	type jsonOutput struct {
		DryRun      bool          `json:"dry_run"`
		HasFailures bool          `json:"has_failures"`
		Groups      []jsonGroup   `json:"groups,omitempty"`
		Projects    []jsonProject `json:"projects,omitempty"`
	}

	toJSONDomain := func(d internal.DomainCopyResult) jsonDomain {
		jd := jsonDomain{Domain: d.Domain}
		if d.Error != nil {
			jd.Error = d.Error.Error()
		}
		for _, item := range d.Items {
			ji := jsonItem{Key: item.Key, Action: item.Label(), DryRun: item.DryRun}
			if item.Error != nil {
				ji.Error = item.Error.Error()
			}
			jd.Items = append(jd.Items, ji)
		}
		return jd
	}

	out := jsonOutput{
		DryRun:      result.DryRun,
		HasFailures: result.HasFailures,
	}
	for _, gr := range result.Groups {
		jg := jsonGroup{GroupPath: gr.GroupPath}
		for _, d := range gr.Domains {
			jg.Domains = append(jg.Domains, toJSONDomain(d))
		}
		out.Groups = append(out.Groups, jg)
	}
	for _, gpg := range result.ProjectGroups {
		for _, pr := range gpg.Projects {
			jp := jsonProject{ProjectPath: pr.ProjectPath, GroupPath: pr.GroupPath}
			for _, d := range pr.Domains {
				jp.Domains = append(jp.Domains, toJSONDomain(d))
			}
			out.Projects = append(out.Projects, jp)
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing JSON file: %w", err)
	}
	return path, nil
}
