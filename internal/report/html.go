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
	title := "GitLab Migration Copy Report"
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
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f5f5f5; color: #333; font-size: 14px; }
  header { background: #1a1a2e; color: white; padding: 20px 32px; display: flex; align-items: center; justify-content: space-between; }
  header h1 { font-size: 20px; font-weight: 600; }
  header .meta { font-size: 12px; color: #aaa; }
  .status-banner { padding: 12px 32px; font-weight: 600; font-size: 14px; }
  .status-banner.has-failures { background: #fdecea; color: #c0392b; border-left: 4px solid #c0392b; }
  .status-banner.clean { background: #e8f8f0; color: #1e8449; border-left: 4px solid #1e8449; }
  .status-banner.dry-run { background: #eaf4ff; color: #1a73e8; border-left: 4px solid #1a73e8; }
  .container { max-width: 1100px; margin: 24px auto; padding: 0 24px; }
  .summary-table { background: white; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.08); margin-bottom: 24px; overflow: hidden; }
  .summary-table table { width: 100%%; border-collapse: collapse; }
  .summary-table th { background: #f8f8f8; text-align: left; padding: 10px 16px; font-weight: 600; font-size: 12px; text-transform: uppercase; letter-spacing: .05em; color: #666; border-bottom: 1px solid #eee; }
  .summary-table td { padding: 10px 16px; border-bottom: 1px solid #f0f0f0; }
  .summary-table tr:last-child td { border-bottom: none; }
  .summary-table tr:hover td { background: #fafafa; }
  .summary-table a { color: #1a73e8; text-decoration: none; font-weight: 500; }
  .group-card { background: white; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.08); margin-bottom: 16px; overflow: hidden; }
  .group-header { padding: 14px 20px; cursor: pointer; display: flex; align-items: center; justify-content: space-between; user-select: none; }
  .group-header:hover { background: #fafafa; }
  .group-header h2 { font-size: 15px; font-weight: 600; font-family: monospace; }
  .group-header .badges { display: flex; gap: 8px; align-items: center; }
  .badge { padding: 2px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; }
  .badge.has-changes { background: #e8f4fd; color: #1a73e8; }
  .badge.has-failures { background: #fdecea; color: #c0392b; }
  .badge.clean { background: #e8f8f0; color: #1e8449; }
  .badge.dry-run { background: #eaf4ff; color: #1a73e8; }
  .chevron { transition: transform .2s; color: #999; }
  .group-body { padding: 0 20px 16px; display: none; }
  .group-body.open { display: block; }
  .section-title { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: #888; margin: 16px 0 8px; }
  .domain-row { display: flex; align-items: flex-start; padding: 6px 0; border-bottom: 1px solid #f5f5f5; gap: 16px; }
  .domain-row:last-child { border-bottom: none; }
  .domain-name { min-width: 220px; flex-shrink: 0; font-family: monospace; font-size: 13px; color: #555; }
  .domain-status { flex: 1; min-width: 0; }
  .domain-error { color: #e67e22; font-size: 13px; }
  .item-list { margin-top: 2px; }
  .item-row { display: flex; align-items: center; gap: 8px; padding: 2px 0; font-size: 13px; }
  .item-label { font-size: 11px; font-weight: 700; padding: 1px 6px; border-radius: 10px; text-transform: uppercase; white-space: nowrap; }
  .item-label.created   { background: #e8f8f0; color: #1e8449; }
  .item-label.updated   { background: #fff8e1; color: #e67e22; }
  .item-label.skipped   { background: #f5f5f5; color: #999; }
  .item-label.failed    { background: #fdecea; color: #c0392b; }
  .item-label.drycreate { background: #eaf4ff; color: #1a73e8; }
  .item-label.dryupdate { background: #eaf4ff; color: #1a73e8; }
  .item-label.dryskip   { background: #f5f5f5; color: #999; }
  .item-key { color: #333; font-family: monospace; }
  .item-warn { color: #e67e22; font-size: 12px; margin-left: 4px; }
  .item-err  { color: #c0392b; font-size: 12px; margin-left: 4px; }
  .all-skipped { color: #aaa; font-size: 12px; font-style: italic; }
  .diff-list { margin: 4px 0 6px 0; background: #fafafa; border-radius: 4px; padding: 6px 10px; border-left: 3px solid #f0c040; }
  .diff-row { display: flex; gap: 8px; align-items: baseline; font-size: 12px; padding: 3px 0; font-family: monospace; flex-wrap: wrap; }
  .diff-field { color: #555; min-width: 220px; flex-shrink: 0; font-weight: 500; }
  .diff-label-src { color: #c0392b; font-weight: 700; font-size: 11px; }
  .diff-label-dst { color: #1e8449; font-weight: 700; font-size: 11px; }
  .diff-val-src { color: #c0392b; }
  .diff-val-dst { color: #1e8449; }
  .toggle-all { background: none; border: 1px solid #ddd; border-radius: 6px; padding: 6px 14px; font-size: 12px; cursor: pointer; color: #555; margin-bottom: 16px; }
  .toggle-all:hover { background: #f5f5f5; }
  .tabs { display: flex; gap: 4px; margin-bottom: 20px; border-bottom: 2px solid #eee; padding-bottom: 0; }
  .tab-btn { background: none; border: none; border-bottom: 3px solid transparent; padding: 8px 20px; font-size: 14px; font-weight: 600; cursor: pointer; color: #666; margin-bottom: -2px; }
  .tab-btn.active { color: #1a73e8; border-bottom-color: #1a73e8; }
  .tab-panel { display: none; }
  .tab-panel.active { display: block; }
  .project-group-header { font-size: 14px; font-weight: 700; color: #444; padding: 16px 0 8px; border-bottom: 1px solid #eee; margin-bottom: 8px; }
  .project-card { background: white; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.08); margin-bottom: 12px; overflow: hidden; }
  .project-header { display: flex; align-items: center; justify-content: space-between; padding: 12px 20px; cursor: pointer; }
  .project-header h3 { font-size: 14px; margin: 0; font-family: monospace; }
  .project-header .project-path { font-size: 11px; color: #888; font-weight: 400; margin-left: 8px; font-family: monospace; }
  .project-body { padding: 0 20px 16px; display: none; }
  .project-body.open { display: block; }
  .dry-run-banner { background: #eaf4ff; border: 1px solid #b3d4ff; border-radius: 6px; padding: 8px 14px; color: #1a73e8; font-size: 13px; margin-bottom: 16px; }
  .stats-grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px; margin-bottom: 24px; }
  .stats-card { background: white; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,.08); padding: 16px 20px; }
  .stats-card h3 { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: #888; margin-bottom: 12px; }
  .stats-numbers { display: flex; gap: 24px; }
  .stats-num { text-align: center; }
  .stats-num .val { font-size: 28px; font-weight: 700; line-height: 1; }
  .stats-num .lbl { font-size: 11px; color: #888; margin-top: 3px; }
  .stats-num.created .val { color: #1e8449; }
  .stats-num.updated .val { color: #e67e22; }
  .stats-num.skipped .val { color: #999; }
  .stats-num.failed .val  { color: #c0392b; }
  .stats-num.neutral .val { color: #333; }
  .stats-num.good .val    { color: #1e8449; }
  .stats-num.warn .val    { color: #e67e22; }
  .stats-num.bad .val     { color: #c0392b; }
  .domain-stats-table { width: 100%%; border-collapse: collapse; font-size: 12px; }
  .domain-stats-table th { text-align: right; padding: 4px 10px; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; color: #888; border-bottom: 1px solid #eee; }
  .domain-stats-table th:first-child { text-align: left; }
  .domain-stats-table td { padding: 5px 10px; text-align: right; border-bottom: 1px solid #f5f5f5; font-family: monospace; }
  .domain-stats-table td:first-child { text-align: left; font-family: monospace; color: #555; }
  .domain-stats-table tr:last-child td { border-bottom: none; }
  .domain-stats-table tr:hover td { background: #fafafa; }
  .dst-created { color: #1e8449; }
  .dst-updated { color: #e67e22; }
  .dst-failed  { color: #c0392b; font-weight: 600; }
  .dst-zero    { color: #ddd; }
</style>
</head>
<body>
<header>
  <h1>%s</h1>
  <div class="meta">Generated: %s UTC</div>
</header>
`, title, title, time.Now().UTC().Format("2006-01-02 15:04:05"))

	// Status banner
	if result.DryRun {
		fmt.Fprintf(f, `<div class="status-banner dry-run">🔍 Dry-run mode — no changes were made. Actions show what <em>would</em> happen.</div>`)
	} else if result.HasFailures {
		fmt.Fprintf(f, `<div class="status-banner has-failures">✗ Copy finished with failures — review required</div>`)
	} else {
		fmt.Fprintf(f, `<div class="status-banner clean">✓ Copy complete — %d created, %d updated, %d skipped, %d failed</div>`,
			created, updated, skipped, failed)
	}

	fmt.Fprintf(f, `<div class="container">`)

	// Compute counts needed for stats and summary table
	groupCount := len(result.Groups)
	projectCount := 0
	for _, gpg := range result.ProjectGroups {
		projectCount += len(gpg.Projects)
	}

	// --- Stats section ---
	writeStats(f, result, created, updated, skipped, failed, groupCount, projectCount)

	// Summary table

	fmt.Fprintf(f, `<div class="summary-table"><table>
<thead><tr><th>Scope</th><th>Created</th><th>Updated</th><th>Skipped</th><th>Failed</th></tr></thead><tbody>`)
	for _, gr := range result.Groups {
		c, u, s, fa := 0, 0, 0, 0
		for _, d := range gr.Domains {
			dc, du, ds, df := d.Counts()
			c += dc
			u += du
			s += ds
			fa += df
		}
		fmt.Fprintf(f, `<tr><td><a href="#group-%s">%s</a> <span style="color:#888;font-size:11px">(group)</span></td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`,
			htmlID(gr.GroupPath), htmlEsc(gr.GroupPath), c, u, s, fa)
	}
	fmt.Fprintf(f, `<tr style="font-weight:600;background:#f8f8f8"><td>Total (%d groups, %d projects)</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`,
		groupCount, projectCount, created, updated, skipped, failed)
	fmt.Fprintf(f, `</tbody></table></div>`)

	// Tabs
	fmt.Fprintf(f, `<div class="tabs">
  <button class="tab-btn active" id="btn-groups" onclick="showTab('groups')">Groups (%d)</button>
  <button class="tab-btn" id="btn-projects" onclick="showTab('projects')">Projects (%d)</button>
</div>`, groupCount, projectCount)

	// Groups pane
	fmt.Fprintf(f, `<div id="tab-groups" class="tab-panel active">`)
	if result.DryRun {
		fmt.Fprintf(f, `<div class="dry-run-banner">🔍 Dry-run mode — no changes were made. Actions show what <em>would</em> happen.</div>`)
	}
	fmt.Fprintf(f, `<button class="toggle-all" onclick="toggleAll()">Expand All</button>`)
	for _, gr := range result.Groups {
		writeGroupHTML(f, gr, result.DryRun)
	}
	fmt.Fprintf(f, `</div>`)

	// Projects pane
	fmt.Fprintf(f, `<div id="tab-projects" class="tab-panel">`)
	if result.DryRun {
		fmt.Fprintf(f, `<div class="dry-run-banner">🔍 Dry-run mode — no changes were made. Actions show what <em>would</em> happen.</div>`)
	}
	for _, gpg := range result.ProjectGroups {
		fmt.Fprintf(f, `<div class="project-group-header">%s</div>`, htmlEsc(gpg.GroupPath))
		for _, pr := range gpg.Projects {
			writeProjectHTML(f, pr, result.DryRun)
		}
	}
	fmt.Fprintf(f, `</div>`)

	fmt.Fprintf(f, `</div>`) // container

	fmt.Fprintf(f, `
<script>
function toggleGroup(header) {
  const body = header.nextElementSibling;
  const chevron = header.querySelector('.chevron');
  const isOpen = body.classList.toggle('open');
  if (chevron) chevron.style.transform = isOpen ? 'rotate(180deg)' : '';
}
function toggleProject(header) {
  const body = header.nextElementSibling;
  const chevron = header.querySelector('.chevron');
  const isOpen = body.classList.toggle('open');
  if (chevron) chevron.style.transform = isOpen ? 'rotate(180deg)' : '';
}
function toggleAll() {
  const bodies = document.querySelectorAll('.group-body, .project-body');
  const btn = document.querySelector('.toggle-all');
  const anyOpen = Array.from(bodies).some(b => b.classList.contains('open'));
  bodies.forEach(b => b.classList.toggle('open', !anyOpen));
  document.querySelectorAll('.chevron').forEach(c => c.style.transform = anyOpen ? '' : 'rotate(180deg)');
  btn.textContent = anyOpen ? 'Expand All' : 'Collapse All';
}
function showTab(name) {
  document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
  const panel = document.getElementById('tab-' + name);
  if (panel) panel.classList.add('active');
  const btn = document.getElementById('btn-' + name);
  if (btn) btn.classList.add('active');
}
document.querySelectorAll('.group-card, .project-card').forEach(card => {
  if (card.querySelector('.badge.has-failures')) {
    const body = card.querySelector('.group-body, .project-body');
    const chevron = card.querySelector('.chevron');
    if (body) body.classList.add('open');
    if (chevron) chevron.style.transform = 'rotate(180deg)';
  }
});
</script>
</body></html>`)

	return path, nil
}

func writeGroupHTML(f *os.File, gr internal.GroupCopyResult, dryRun bool) {
	c, u, s, fa := 0, 0, 0, 0
	for _, d := range gr.Domains {
		dc, du, ds, df := d.Counts()
		c += dc
		u += du
		s += ds
		fa += df
	}
	_ = s
	badgeClass, badgeText := groupBadge(c, u, fa, dryRun)

	fmt.Fprintf(f, `<div class="group-card" id="group-%s">
  <div class="group-header" onclick="toggleGroup(this)">
    <h2>%s</h2>
    <div class="badges"><span class="badge %s">%s</span><span class="chevron">▼</span></div>
  </div>
  <div class="group-body">
    <div class="section-title">Domains</div>`,
		htmlID(gr.GroupPath), htmlEsc(gr.GroupPath), badgeClass, badgeText)

	for _, d := range gr.Domains {
		writeDomainRowHTML(f, d)
	}
	fmt.Fprintf(f, `</div></div>`)
}

func writeProjectHTML(f *os.File, pr internal.ProjectCopyResult, dryRun bool) {
	c, u, s, fa := 0, 0, 0, 0
	for _, d := range pr.Domains {
		dc, du, ds, df := d.Counts()
		c += dc
		u += du
		s += ds
		fa += df
	}
	_ = s
	badgeClass, badgeText := groupBadge(c, u, fa, dryRun)

	fmt.Fprintf(f, `<div class="project-card" id="proj-%s">
  <div class="project-header" onclick="toggleProject(this)">
    <h3>%s<span class="project-path">%s</span></h3>
    <div class="badges"><span class="badge %s">%s</span><span class="chevron">▼</span></div>
  </div>
  <div class="project-body">`,
		htmlID(pr.ProjectPath),
		htmlEsc(projectName(pr.ProjectPath)),
		htmlEsc(pr.ProjectPath),
		badgeClass, badgeText)

	for _, d := range pr.Domains {
		writeDomainRowHTML(f, d)
	}
	fmt.Fprintf(f, `</div></div>`)
}

func writeDomainRowHTML(f *os.File, d internal.DomainCopyResult) {
	fmt.Fprintf(f, `<div class="domain-row"><div class="domain-name">%s</div><div class="domain-status">`,
		htmlEsc(d.Domain))

	if d.Error != nil {
		fmt.Fprintf(f, `<span class="domain-error">⚠ %s</span>`, htmlEsc(d.Error.Error()))
		fmt.Fprintf(f, `</div></div>`)
		return
	}

	if len(d.Items) == 0 {
		fmt.Fprintf(f, `<span class="all-skipped">— no items</span>`)
		fmt.Fprintf(f, `</div></div>`)
		return
	}

	allSkipped := true
	for _, item := range d.Items {
		if item.Action != internal.ActionSkipped {
			allSkipped = false
			break
		}
	}
	if allSkipped {
		fmt.Fprintf(f, `<span class="all-skipped">all %d skipped</span>`, len(d.Items))
		fmt.Fprintf(f, `</div></div>`)
		return
	}

	fmt.Fprintf(f, `<div class="item-list">`)
	for _, item := range d.Items {
		labelClass, labelText := itemLabelClassText(item)
		extra := ""
		if item.Error != nil {
			if item.Action == internal.ActionFailed {
				extra = fmt.Sprintf(`<span class="item-err">⚠ %s</span>`, htmlEsc(item.Error.Error()))
			} else {
				extra = fmt.Sprintf(`<span class="item-warn">⚠ %s</span>`, htmlEsc(item.Error.Error()))
			}
		}
		fmt.Fprintf(f, `<div class="item-row"><span class="item-label %s">%s</span><span class="item-key">%s</span>%s</div>`,
			labelClass, labelText, htmlEsc(item.Key), extra)
		if len(item.Diffs) > 0 {
			fmt.Fprintf(f, `<div class="diff-list">`)
			for _, d := range item.Diffs {
				fmt.Fprintf(f, `<div class="diff-row"><span class="diff-field">%s</span><span class="diff-label-src">dest:</span> <span class="diff-val-src">%s</span><span class="diff-label-dst">→ source:</span> <span class="diff-val-dst">%s</span></div>`,
					htmlEsc(d.Field), htmlEsc(d.Dst), htmlEsc(d.Src))
			}
			fmt.Fprintf(f, `</div>`)
		}
	}
	fmt.Fprintf(f, `</div>`)
	fmt.Fprintf(f, `</div></div>`)
}

func groupBadge(created, updated, failed int, dryRun bool) (class, text string) {
	if failed > 0 {
		return "has-failures", fmt.Sprintf("✗ %d failed", failed)
	}
	changes := created + updated
	if dryRun {
		if changes > 0 {
			return "dry-run", fmt.Sprintf("~ %d would change", changes)
		}
		return "clean", "✓ no changes needed"
	}
	if changes > 0 {
		return "has-changes", fmt.Sprintf("✓ %d changed", changes)
	}
	return "clean", "✓ no changes needed"
}

func itemLabelClassText(item internal.ItemResult) (class, text string) {
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

func projectName(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func htmlID(s string) string {
	r := strings.NewReplacer("/", "-", " ", "-", ".", "-", "_", "-")
	return r.Replace(s)
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

type domainStat struct {
	name    string
	created int
	updated int
	skipped int
	failed  int
}

func writeStats(f *os.File, result *internal.RunResult, created, updated, skipped, failed, groupCount, projectCount int) {
	// --- Domain breakdown ---
	domainMap := make(map[string]*domainStat)
	domainOrder := []string{}

	addDomainItems := func(domainName string, items []internal.ItemResult) {
		ds, ok := domainMap[domainName]
		if !ok {
			ds = &domainStat{name: domainName}
			domainMap[domainName] = ds
			domainOrder = append(domainOrder, domainName)
		}
		for _, item := range items {
			switch item.Action {
			case internal.ActionCreated:
				ds.created++
			case internal.ActionUpdated:
				ds.updated++
			case internal.ActionSkipped:
				ds.skipped++
			case internal.ActionFailed:
				ds.failed++
			}
		}
	}

	for _, gr := range result.Groups {
		for _, d := range gr.Domains {
			addDomainItems(d.Domain, d.Items)
		}
	}
	for _, gpg := range result.ProjectGroups {
		for _, pr := range gpg.Projects {
			for _, d := range pr.Domains {
				addDomainItems(d.Domain, d.Items)
			}
		}
	}

	// --- Project health ---
	inSync, hadChanges, hadFailures := 0, 0, 0
	for _, gpg := range result.ProjectGroups {
		for _, pr := range gpg.Projects {
			c, u, _, fa := 0, 0, 0, 0
			for _, d := range pr.Domains {
				dc, du, _, df := d.Counts()
				c += dc
				u += du
				fa += df
			}
			if fa > 0 {
				hadFailures++
			} else if c+u > 0 {
				hadChanges++
			} else {
				inSync++
			}
		}
	}

	// --- Count domains that actually ran (had any items) ---
	domainsRun := len(domainOrder)

	fmt.Fprintf(f, `<div class="stats-grid">`)

	// Panel 1 — Run summary
	fmt.Fprintf(f, `<div class="stats-card"><h3>Run Summary</h3><div class="stats-numbers">`)
	fmt.Fprintf(f, `<div class="stats-num created"><div class="val">%d</div><div class="lbl">Created</div></div>`, created)
	fmt.Fprintf(f, `<div class="stats-num updated"><div class="val">%d</div><div class="lbl">Updated</div></div>`, updated)
	fmt.Fprintf(f, `<div class="stats-num skipped"><div class="val">%d</div><div class="lbl">Skipped</div></div>`, skipped)
	fmt.Fprintf(f, `<div class="stats-num failed"><div class="val">%d</div><div class="lbl">Failed</div></div>`, failed)
	fmt.Fprintf(f, `</div>`)
	fmt.Fprintf(f, `<div style="margin-top:12px;font-size:12px;color:#888">%d group(s) &nbsp;·&nbsp; %d project(s) &nbsp;·&nbsp; %d domain(s)</div>`, groupCount, projectCount, domainsRun)
	fmt.Fprintf(f, `</div>`)

	// Panel 2 — Project health
	fmt.Fprintf(f, `<div class="stats-card"><h3>Project Health</h3><div class="stats-numbers">`)
	fmt.Fprintf(f, `<div class="stats-num good"><div class="val">%d</div><div class="lbl">In Sync</div></div>`, inSync)
	fmt.Fprintf(f, `<div class="stats-num warn"><div class="val">%d</div><div class="lbl">Had Changes</div></div>`, hadChanges)
	fmt.Fprintf(f, `<div class="stats-num bad"><div class="val">%d</div><div class="lbl">Had Failures</div></div>`, hadFailures)
	fmt.Fprintf(f, `</div></div>`)

	// Panel 3 — Domain breakdown
	fmt.Fprintf(f, `<div class="stats-card"><h3>By Domain</h3>`)
	fmt.Fprintf(f, `<table class="domain-stats-table"><thead><tr><th>Domain</th><th>Created</th><th>Updated</th><th>Skipped</th><th>Failed</th></tr></thead><tbody>`)
	for _, name := range domainOrder {
		ds := domainMap[name]
		fmtNum := func(n int, cls string) string {
			if n == 0 {
				return fmt.Sprintf(`<td class="dst-zero">—</td>`)
			}
			return fmt.Sprintf(`<td class="%s">%d</td>`, cls, n)
		}
		fmt.Fprintf(f, `<tr><td>%s</td>%s%s%s%s</tr>`,
			htmlEsc(name),
			fmtNum(ds.created, "dst-created"),
			fmtNum(ds.updated, "dst-updated"),
			fmtNum(ds.skipped, ""),
			fmtNum(ds.failed, "dst-failed"),
		)
	}
	fmt.Fprintf(f, `</tbody></table></div>`)

	fmt.Fprintf(f, `</div>`) // close stats-grid
}
