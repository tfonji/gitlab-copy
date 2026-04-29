package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab-copy/internal"
)

// TokenEntry holds all context for a single generated token.
type TokenEntry struct {
	Domain string // e.g. "pipeline_triggers", "deploy_tokens", "access_tokens"
	Scope  string // "group" or "project"
	Path   string // group or project path
	Name   string // token name / description
	Token  string // the actual token value
}

// WriteTokensReport collects all generated tokens from the run result and
// writes them to gitlab-copy-tokens.md. Returns true if any tokens were found.
func WriteTokensReport(result *internal.RunResult, dir string) (bool, error) {
	var entries []TokenEntry

	// Collect group-level tokens
	for _, gr := range result.Groups {
		for _, d := range gr.Domains {
			for _, item := range d.Items {
				if item.Token != "" {
					entries = append(entries, TokenEntry{
						Domain: d.Domain,
						Scope:  "group",
						Path:   gr.GroupPath,
						Name:   item.Key,
						Token:  item.Token,
					})
				}
			}
		}
	}

	// Collect project-level tokens
	for _, gpg := range result.ProjectGroups {
		for _, pr := range gpg.Projects {
			for _, d := range pr.Domains {
				for _, item := range d.Items {
					if item.Token != "" {
						entries = append(entries, TokenEntry{
							Domain: d.Domain,
							Scope:  "project",
							Path:   pr.ProjectPath,
							Name:   item.Key,
							Token:  item.Token,
						})
					}
				}
			}
		}
	}

	if len(entries) == 0 {
		return false, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("creating output dir: %w", err)
	}

	path := filepath.Join(dir, "gitlab-copy-tokens.md")
	f, err := os.Create(path)
	if err != nil {
		return false, fmt.Errorf("creating tokens file: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# GitLab Migration — Generated Tokens\n\n")
	fmt.Fprintf(f, "> **SENSITIVE** — This file contains newly generated token values.\n")
	fmt.Fprintf(f, "> Store securely and delete after updating all references.\n\n")
	fmt.Fprintf(f, "These tokens were generated on the destination instance during the migration run.\n")
	fmt.Fprintf(f, "The source token values cannot be recovered — update any CI variables, webhooks,\n")
	fmt.Fprintf(f, "or external services that referenced the old tokens with the new values below.\n\n")

	fmt.Fprintf(f, "| Type | Scope | Path | Name / Description | Token Value |\n")
	fmt.Fprintf(f, "|------|-------|------|--------------------|-------------|\n")

	for _, e := range entries {
		tokenType := tokenTypeLabel(e.Domain)
		fmt.Fprintf(f, "| %s | %s | `%s` | %s | `%s` |\n",
			tokenType,
			e.Scope,
			e.Path,
			mdEsc(e.Name),
			e.Token,
		)
	}

	return true, nil
}

func tokenTypeLabel(domain string) string {
	switch domain {
	case "pipeline_triggers":
		return "Pipeline Trigger"
	case "deploy_tokens":
		return "Deploy Token"
	case "access_tokens":
		return "Access Token"
	default:
		return strings.ReplaceAll(domain, "_", " ")
	}
}

func mdEsc(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	s = strings.ReplaceAll(s, "`", "'")
	return s
}
