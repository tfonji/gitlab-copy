package internal

// CopyAction describes what the tool did (or would do in dry-run) for a single item.
type CopyAction string

const (
	ActionCreated CopyAction = "created"
	ActionUpdated CopyAction = "updated"
	ActionSkipped CopyAction = "skipped"
	ActionFailed  CopyAction = "failed"
)

// DiffLine shows a single field that differs between source and dest.
type DiffLine struct {
	Field string
	Src   string
	Dst   string
}

// ItemResult is one item within a domain — either a singleton (push_rules,
// topics) or one entry in a collection (a named protected environment, etc).
type ItemResult struct {
	Key    string
	Action CopyAction
	DryRun bool       // if true, Action is what WOULD have happened — no write was made
	Error  error      // only set when Action == ActionFailed or as a warning
	Diffs  []DiffLine // populated for Updated items to show what changed
}

// Label returns the display string for the item's action, incorporating
// dry-run state: e.g. "DryRun(Create)", "Created", "Skipped".
func (r ItemResult) Label() string {
	if !r.DryRun {
		return string(r.Action)
	}
	switch r.Action {
	case ActionCreated:
		return "DryRun(Create)"
	case ActionUpdated:
		return "DryRun(Update)"
	case ActionSkipped:
		return "DryRun(Skip)"
	default:
		return "DryRun(" + string(r.Action) + ")"
	}
}

// DomainCopyResult holds all item results for one domain within one group or project.
type DomainCopyResult struct {
	Domain string
	Items  []ItemResult
	Error  error // domain-level error — e.g. could not fetch source; Items may be empty
}

func (d DomainCopyResult) HasFailures() bool {
	if d.Error != nil {
		return true
	}
	for _, item := range d.Items {
		if item.Action == ActionFailed {
			return true
		}
	}
	return false
}

// Counts returns (created, updated, skipped, failed) tallies across all items.
func (d DomainCopyResult) Counts() (created, updated, skipped, failed int) {
	for _, item := range d.Items {
		switch item.Action {
		case ActionCreated:
			created++
		case ActionUpdated:
			updated++
		case ActionSkipped:
			skipped++
		case ActionFailed:
			failed++
		}
	}
	return
}

type GroupCopyResult struct {
	GroupPath string
	Domains   []DomainCopyResult
}

type ProjectCopyResult struct {
	ProjectPath string
	GroupPath   string
	Domains     []DomainCopyResult
}

type GroupProjectCopyResults struct {
	GroupPath string
	Projects  []ProjectCopyResult
}

type RunResult struct {
	DryRun        bool
	Groups        []GroupCopyResult
	ProjectGroups []GroupProjectCopyResults
	HasFailures   bool
}

// TotalCounts sums created, updated, skipped, failed across the entire run.
func (r *RunResult) TotalCounts() (created, updated, skipped, failed int) {
	add := func(d DomainCopyResult) {
		c, u, s, f := d.Counts()
		created += c
		updated += u
		skipped += s
		failed += f
	}
	for _, gr := range r.Groups {
		for _, d := range gr.Domains {
			add(d)
		}
	}
	for _, gpg := range r.ProjectGroups {
		for _, pr := range gpg.Projects {
			for _, d := range pr.Domains {
				add(d)
			}
		}
	}
	return
}
