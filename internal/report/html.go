package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab-copy/internal"
)

func WriteHTML(result *internal.RunResult, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating output dir: %w", err)
	}
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("gitlab-copy-%s.html", timestamp)
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("creating HTML file: %w", err)
	}
	defer f.Close()

	created, updated, skipped, failed := result.TotalCounts()
	title := "GitLab Copy Report"
	if result.DryRun {
		title += " [DRY-RUN]"
	}

	fmt.Fprintf(f, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
         background: #0f1117; color: #e2e8f0; font-size: 14px; }
  .header { background: #1a1d2e; border-bottom: 1px solid #2d3748; padding: 20px 32px; }
  .header h1 { font-size: 20px; font-weight: 600; color: #7c83e5; }
  .header .meta { color: #718096; font-size: 13px; margin-top: 4px; }
  .summary-bar { display: flex; gap: 24px; padding: 16px 32px;
                 background: #141720; border-bottom: 1px solid #2d3748; }
  .stat { display: flex; flex-direction: column; }
  .stat .num { font-size: 22px; font-weight: 700; }
  .stat .lbl { font-size: 12px; color: #718096; text-transform: uppercase; letter-spacing: .05em; }
  .stat.created .num { color: #68d391; }
  .stat.updated .num { color: #f6e05e; }
  .stat.skipped .num { color: #718096; }
  .stat.failed  .num { color: #fc8181; }
  .tabs { display: flex; background: #1a1d2e; border-bottom: 1px solid #2d3748; padding: 0 32px; }
  .tab { padding: 12px 20px; cursor: pointer; color: #718096; font-weight: 500;
         border-bottom: 2px solid transparent; transition: all .15s; }
  .tab.active { color: #7c83e5; border-color: #7c83e5; }
  .tab:hover:not(.active) { color: #a0aec0; }
  .pane { display: none; padding: 24px 32px; }
  .pane.active { display: block; }
  .group-block { margin-bottom: 28px; }
  .group-title { font-size: 15px; font-weight: 600; color: #7c83e5;
                 padding: 8px 0; border-bottom: 1px solid #2d3748; margin-bottom: 12px; }
  .project-block { margin-left: 16px; margin-bottom: 20px; }
  .project-title { font-size: 13px; font-weight: 600; color: #a0aec0; margin-bottom: 8px; }
  .domain-block { margin-bottom: 10px; }
  .domain-name { font-size: 12px; font-weight: 600; color: #718096;
                 text-transform: uppercase; letter-spacing: .05em; margin-bottom: 4px; }
  .domain-error { color: #fc8181; font-size: 12px; padding: 4px 8px;
                  background: #2d1515; border-radius: 4px; }
  .items { display: flex; flex-direction: column; gap: 2px; }
  .item { display: flex; align-items: center; gap: 8px; padding: 3px 8px;
          border-radius: 4px; font-size: 13px; }
  .item:hover { background: #1e2235; }
  .badge { font-size: 11px; font-weight: 600; padding: 1px 7px; border-radius: 10px;
           text-transform: uppercase; letter-spacing: .04em; white-space: nowrap; }
  .badge.created  { background: #1a3d2b; color: #68d391; }
  .badge.updated  { background: #3d3519; color: #f6e05e; }
  .badge.skipped  { background: #1e2235; color: #718096; }
  .badge.failed   { background: #3d1515; color: #fc8181; }
  .badge.drycreate { background: #1a2a3d; color: #63b3ed; }
  .badge.dryupdate { background: #1a2a3d; color: #76e4f7; }
  .badge.dryskip   { background: #1e2235; color: #718096; }
  .item-key { color: #e2e8f0; }
  .item-error { color: #fc8181; font-size: 12px; margin-left: 4px; }
  .all-skipped { color: #4a5568; font-size: 12px; font-style: italic; padding: 2px 8px; }
  .drybanner { background: #1a2a3d; border: 1px solid #2c4a6e; border-radius: 6px;
               padding: 10px 16px; margin-bottom: 20px; color: #63b3ed; font-size: 13px; }
</style>
</head>
<body>
<div class="header">
  <h1>%s</h1>
  <div class="meta">Generated %s</div>
</div>
<div class="summary-bar">
  <div class="stat created"><span class="num">%d</span><span class="lbl">Created</span></div>
  <div class="stat updated"><span class="num">%d</span><span class="lbl">Updated</span></div>
  <div class="stat skipped"><span class="num">%d</span><span class="lbl">Skipped</span></div>
  <div class="stat failed"><span class="num">%d</span><span class="lbl">Failed</span></div>
</div>
<div class="tabs">
  <div class="tab active" onclick="showTab('groups')">Groups</div>
  <div class="tab" onclick="showTab('projects')">Projects</div>
</div>
`, title, title, time.Now().Format("2006-01-02 15:04:05"), created, updated, skipped, failed)

	// Groups pane
	fmt.Fprintf(f, `<div id="pane-groups" class="pane active">`)
	if result.DryRun {
		fmt.Fprintf(f, `<div class="drybanner">🔍 Dry-run mode — no changes were made. Actions show what <em>would</em> happen.</div>`)
	}
	for _, gr := range result.Groups {
		fmt.Fprintf(f, `<div class="group-block"><div class="group-title">%s</div>`, htmlEsc(gr.GroupPath))
		for _, d := range gr.Domains {
			writeDomainHTML(f, d)
		}
		fmt.Fprintf(f, `</div>`)
	}
	fmt.Fprintf(f, `</div>`)

	// Projects pane
	fmt.Fprintf(f, `<div id="pane-projects" class="pane">`)
	if result.DryRun {
		fmt.Fprintf(f, `<div class="drybanner">🔍 Dry-run mode — no changes were made. Actions show what <em>would</em> happen.</div>`)
	}
	for _, gpg := range result.ProjectGroups {
		fmt.Fprintf(f, `<div class="group-block"><div class="group-title">%s</div>`, htmlEsc(gpg.GroupPath))
		for _, pr := range gpg.Projects {
			fmt.Fprintf(f, `<div class="project-block"><div class="project-title">%s</div>`, htmlEsc(pr.ProjectPath))
			for _, d := range pr.Domains {
				writeDomainHTML(f, d)
			}
			fmt.Fprintf(f, `</div>`)
		}
		fmt.Fprintf(f, `</div>`)
	}
	fmt.Fprintf(f, `</div>`)

	fmt.Fprintf(f, `
<script>
function showTab(name) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.pane').forEach(p => p.classList.remove('active'));
  document.querySelector('[onclick="showTab(\''+name+'\')"]').classList.add('active');
  document.getElementById('pane-'+name).classList.add('active');
}
</script>
</body></html>`)

	return path, nil
}

func writeDomainHTML(f *os.File, d internal.DomainCopyResult) {
	fmt.Fprintf(f, `<div class="domain-block"><div class="domain-name">%s</div>`, htmlEsc(d.Domain))

	if d.Error != nil {
		fmt.Fprintf(f, `<div class="domain-error">Error: %s</div>`, htmlEsc(d.Error.Error()))
		fmt.Fprintf(f, `</div>`)
		return
	}

	// Check if all items are skipped for compact rendering
	allSkipped := true
	for _, item := range d.Items {
		if item.Action != internal.ActionSkipped {
			allSkipped = false
			break
		}
	}

	if allSkipped && len(d.Items) > 0 {
		fmt.Fprintf(f, `<div class="all-skipped">all %d skipped</div>`, len(d.Items))
		fmt.Fprintf(f, `</div>`)
		return
	}

	fmt.Fprintf(f, `<div class="items">`)
	for _, item := range d.Items {
		badgeClass, label := itemBadge(item)
		errStr := ""
		if item.Error != nil {
			errStr = fmt.Sprintf(`<span class="item-error">%s</span>`, htmlEsc(item.Error.Error()))
		}
		fmt.Fprintf(f, `<div class="item"><span class="badge %s">%s</span><span class="item-key">%s</span>%s</div>`,
			badgeClass, label, htmlEsc(item.Key), errStr)
	}
	fmt.Fprintf(f, `</div></div>`)
}

func itemBadge(item internal.ItemResult) (class, label string) {
	if item.DryRun {
		switch item.Action {
		case internal.ActionCreated:
			return "drycreate", "DryRun(Create)"
		case internal.ActionUpdated:
			return "dryupdate", "DryRun(Update)"
		default:
			return "dryskip", "DryRun(Skip)"
		}
	}
	switch item.Action {
	case internal.ActionCreated:
		return "created", "Created"
	case internal.ActionUpdated:
		return "updated", "Updated"
	case internal.ActionSkipped:
		return "skipped", "Skipped"
	case internal.ActionFailed:
		return "failed", "Failed"
	default:
		return "skipped", string(item.Action)
	}
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
