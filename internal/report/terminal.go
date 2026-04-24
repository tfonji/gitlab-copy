package report

import (
	"fmt"
	"io"
	"strings"

	"gitlab-copy/internal"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

type Terminal struct {
	w     io.Writer
	color bool
}

func NewTerminal(w io.Writer, color bool) *Terminal {
	return &Terminal{w: w, color: color}
}

func (t *Terminal) Write(result *internal.RunResult) {
	t.writeln("")
	header := "GITLAB MIGRATION COPY REPORT"
	if result.DryRun {
		header += " [DRY-RUN]"
	}
	t.writef("%s%s%s%s\n", colorBold, colorCyan, header, colorReset)
	t.writeln(strings.Repeat("─", 70))

	if len(result.Groups) > 0 {
		t.writef("\n%s%sGROUPS%s\n", colorBold, colorCyan, colorReset)
		t.writeln(strings.Repeat("─", 70))
		for _, gr := range result.Groups {
			t.writeGroup(gr)
		}
	}

	if len(result.ProjectGroups) > 0 {
		t.writef("\n%s%sPROJECTS%s\n", colorBold, colorCyan, colorReset)
		t.writeln(strings.Repeat("─", 70))
		for _, gpg := range result.ProjectGroups {
			t.writeProjectGroup(gpg)
		}
	}

	t.writeln(strings.Repeat("─", 70))

	created, updated, skipped, failed := result.TotalCounts()
	if result.HasFailures {
		t.writef("%s%s✗ Copy finished with failures%s — %d created, %d updated, %d skipped, %d failed\n",
			colorBold, colorRed, colorReset, created, updated, skipped, failed)
	} else if result.DryRun {
		t.writef("%s%s~ Dry-run complete%s — %d would create, %d would update, %d would skip\n",
			colorBold, colorBlue, colorReset, created, updated, skipped)
	} else {
		t.writef("%s%s✓ Copy complete%s — %d created, %d updated, %d skipped, %d failed\n",
			colorBold, colorGreen, colorReset, created, updated, skipped, failed)
	}
	t.writeln("")
}

func (t *Terminal) writeGroup(gr internal.GroupCopyResult) {
	t.writef("\n%s%s%s%s\n", colorBold, colorCyan, gr.GroupPath, colorReset)
	for _, d := range gr.Domains {
		t.writeDomain(d)
	}
}

func (t *Terminal) writeProjectGroup(gpg internal.GroupProjectCopyResults) {
	t.writef("\n%s%s%s%s\n", colorBold, colorCyan, gpg.GroupPath, colorReset)
	for _, pr := range gpg.Projects {
		t.writeProject(pr)
	}
}

func (t *Terminal) writeProject(pr internal.ProjectCopyResult) {
	t.writef("  %s%s%s\n", colorBold, pr.ProjectPath, colorReset)
	for _, d := range pr.Domains {
		t.writeDomain(d)
	}
}

func (t *Terminal) writeDomain(d internal.DomainCopyResult) {
	indent := "    "

	if d.Error != nil {
		t.writef("%s%s%s [ERROR] %v%s\n", indent, colorRed, d.Domain, d.Error, colorReset)
		return
	}

	// Count non-skipped to decide whether to show the domain line
	active := 0
	for _, item := range d.Items {
		if item.Action != internal.ActionSkipped {
			active++
		}
	}

	if active == 0 && len(d.Items) > 0 {
		// All skipped — compact single line
		t.writef("%s%s%s — %s·%s all skipped (%d)\n",
			indent, colorDim, d.Domain, colorDim, colorReset, len(d.Items))
		return
	}

	t.writef("%s%s%s%s\n", indent, colorBold, d.Domain, colorReset)
	for _, item := range d.Items {
		t.writeItem(indent+"  ", item)
	}
}

func (t *Terminal) writeItem(indent string, item internal.ItemResult) {
	symbol, color := itemSymbolColor(item)
	label := item.Label()
	if item.Error != nil {
		t.writef("%s%s%s%s %s — %s: %v%s\n",
			indent, color, symbol, colorReset, item.Key, label, item.Error, colorReset)
	} else {
		t.writef("%s%s%s%s %s — %s\n",
			indent, color, symbol, colorReset, item.Key, label)
	}
}

func itemSymbolColor(item internal.ItemResult) (symbol, color string) {
	if item.DryRun {
		switch item.Action {
		case internal.ActionCreated:
			return "+", colorBlue
		case internal.ActionUpdated:
			return "~", colorBlue
		default:
			return "·", colorDim
		}
	}
	switch item.Action {
	case internal.ActionCreated:
		return "+", colorGreen
	case internal.ActionUpdated:
		return "~", colorYellow
	case internal.ActionSkipped:
		return "·", colorDim
	case internal.ActionFailed:
		return "✗", colorRed
	default:
		return "?", colorDim
	}
}

func (t *Terminal) writef(format string, args ...any) {
	if !t.color {
		// Strip ANSI codes when color is disabled
		format = stripANSI(format)
	}
	fmt.Fprintf(t.w, format, args...)
}

func (t *Terminal) writeln(s string) {
	fmt.Fprintln(t.w, s)
}

// stripANSI removes color escape sequences for no-color output.
func stripANSI(s string) string {
	codes := []string{
		colorReset, colorRed, colorGreen, colorYellow,
		colorCyan, colorBlue, colorBold, colorDim,
	}
	for _, c := range codes {
		s = strings.ReplaceAll(s, c, "")
	}
	return s
}
