# gitlab-copy

Copies GitLab settings that [Congregate](https://gitlab.com/gitlab-org/professional-services-automation/tools/migration/congregate) cannot migrate from a self-managed GitLab instance to GitLab SaaS Dedicated.

Congregate handles repository content (branches, commits, merge requests, wikis). `gitlab-copy` fills the gap by copying settings like environments, protected branches, push rules, approval settings, topics, badges, and more.

---

## Prerequisites

- Go 1.22+ (to build)
- API-scoped tokens for both source and destination instances
- Destination token needs **Owner** on groups and **Maintainer** on projects

```bash
export SOURCE_GITLAB_TOKEN=glpat-xxxx
export DEST_GITLAB_TOKEN=glpat-yyyy
```

---

## Build

```bash
go mod tidy
go build -o gitlab-copy ./cmd/
```

Cross-platform builds via Makefile:

```bash
make all          # all platforms
make linux        # linux/amd64
make mac-arm      # darwin/arm64
make windows      # windows/amd64
```

---

## Configuration

Copy `config.yaml` and edit for your environment:

```yaml
source:
  url: https://source.example.com
  token_env: SOURCE_GITLAB_TOKEN

destination:
  url: https://dest.example.com
  token_env: DEST_GITLAB_TOKEN

groups:
  include:
    - my-group
  exclude:
    - my-group/dast/*          # deep glob — excludes all descendants
    - my-group/dast_rest_scan/*
  include_subgroups: true

projects:
  include: []                  # empty = derive from groups
  exclude:
    - my-group/obsolete-*
  include_subgroups: true
  include_archived: false

concurrency:
  groups: 5
  projects: 10

domains:
  groups:
    - push_rules
    - description
    - default_branch_name
    - mr_settings
    - mr_approval_settings
    - protected_environments
    - approval_rules
    - jira_integration
  projects:
    - topics
    - environments
    - protected_environments
    - jira_integration
    - pipeline_triggers
    - deploy_keys
    - project_push_rules
    - project_mr_approvals
    - project_approval_rules
    - badges
    - project_protected_branches
    - project_protected_tags

output:
  dir: ./gl-copy-report
  formats:
    - terminal
    - html
    - json
```

### Exclusion patterns

| Pattern | Matches |
|---|---|
| `my-group/subgroup` | exact path only |
| `my-group/dast_*` | single-level glob |
| `my-group/*` | all descendants at any depth |

---

## Usage

### Always dry-run first

```bash
./gitlab-copy all -config config.yaml -dry-run
```

Reads both source and dest, shows what *would* happen — no writes made.

### Copy everything (new group batch)

```bash
./gitlab-copy all -config config.yaml -group my-group
```

Copies group-level settings and all project settings under that group.

### Copy projects only (group already migrated)

```bash
./gitlab-copy projects all -config config.yaml -group my-group
```

Skips group domains entirely. Use this for subsequent migration batches where the group is already set up on dest.

### Copy a single project

```bash
./gitlab-copy projects all -config config.yaml -project my-group/my-project
```

No `groups.include` needed in config when targeting a specific project.

### Scope to groups only

```bash
./gitlab-copy groups all -config config.yaml -group my-group
```

### Disable terminal color (for CI)

```bash
./gitlab-copy all -config config.yaml -dry-run -no-color
```

---

## Migration workflow

The recommended sequence for each migration batch:

```
1. Run Congregate        — migrates repo content to dest
2. gitlab-copy dry-run   — preview what will be copied
3. gitlab-copy           — apply the copy
4. gitlab-diff           — validate source and dest match
```

For groups migrated in earlier batches, use `projects all` in step 2 and 3 to avoid re-running group domains unnecessarily.

---

## Domains reference

### Group domains

| Domain | What it copies |
|---|---|
| `push_rules` | Commit message regex, branch name rules, file size limits |
| `description` | Group description |
| `default_branch_name` | Default branch for new projects |
| `mr_settings` | Merge-if-pipeline-succeeds, resolve-all-discussions policies |
| `mr_approval_settings` | Author/committer approval, override permissions, retain approvals on push |
| `protected_environments` | Which roles can deploy to named environments |
| `approval_rules` | MR approval rule names and required counts (approvers need manual assignment) |
| `jira_integration` | Jira integration config (credentials require manual verification) |
| `badges` | Group-level badges inherited by all projects in the group |

### Project domains

| Domain | What it copies |
|---|---|
| `topics` | Project tags/labels (AppID, CMDB ID, etc.) |
| `environments` | Deployment environments with external URLs |
| `protected_environments` | Role-based deploy access per environment |
| `jira_integration` | Jira integration config (credentials require manual verification) |
| `pipeline_triggers` | CI trigger tokens (new token generated on dest — update any references) |
| `deploy_keys` | SSH deploy keys (existing global keys flagged for manual enable) |
| `project_push_rules` | Project-level commit/branch push rules |
| `project_mr_approvals` | Project-level MR approval settings |
| `project_approval_rules` | Project-level approval rules (approvers need manual assignment) |
| `badges` | Pipeline status and coverage badges |
| `project_protected_branches` | Branch protection rules (role-based access levels only) |
| `project_protected_tags` | Tag protection rules (role-based access levels only) |

### What cannot be automated

| Item | Reason |
|---|---|
| `security_policies` | No API for linking a policy project — configure via UI |
| `access_tokens` | Token value only shown at creation — create manually on dest |
| `deploy_tokens` | Secrets cannot be exported — regenerate on dest |
| `variables` (masked/hidden) | Values are masked in API responses — create manually on dest |
| Jira credentials | Password/token masked in GET response — verify manually after copy |

---

## Report

Each run produces a report in `./gl-copy-report/` (configurable):

- **HTML** — tabbed Groups/Projects view with collapsible cards, color-coded actions
- **JSON** — machine-readable output for pipeline integration
- **Terminal** — colored output during the run

### Action labels

| Label | Meaning |
|---|---|
| `Created` | Item didn't exist on dest, was created |
| `Updated` | Item existed but differed, was updated |
| `Skipped` | Item already matches, nothing done |
| `Failed` | Write attempt failed — see error message |
| `DryRun(Create)` | Would be created (dry-run mode) |
| `DryRun(Update)` | Would be updated (dry-run mode) |
| `DryRun(Skip)` | Already matches (dry-run mode) |

Exit code `0` = clean run, `1` = failures present (pipeline-friendly).

---

## Related tools

- **gitlab-diff** — validates that source and dest match after migration
- **Congregate** — handles repository content migration