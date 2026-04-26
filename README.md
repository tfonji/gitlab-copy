# gitlab-copy

`gitlab-copy` copies GitLab settings that [Congregate](https://gitlab.com/gitlab-org/professional-services-automation/tools/migration/congregate) cannot migrate when moving from a self-managed GitLab instance to GitLab SaaS Dedicated.

**Division of responsibility:**

| Tool | What it handles |
|---|---|
| Congregate | Repository content — branches, commits, merge requests, wikis, issues |
| **gitlab-copy** | Settings — environments, push rules, approval rules, topics, badges, protected branches, Jira integration, and more |
| gitlab-diff | Validation — compares source and dest after migration to confirm everything matches |

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Build](#build)
3. [Quick Start](#quick-start)
4. [Migration Workflow](#migration-workflow)
5. [Configuration](#configuration)
6. [Controlling What Gets Copied](#controlling-what-gets-copied)
7. [CLI Reference](#cli-reference)
8. [Reading the Report](#reading-the-report)
9. [Domains Reference](#domains-reference)
10. [What Cannot Be Automated](#what-cannot-be-automated)
11. [FAQ](#faq)

---

## Prerequisites

**API tokens** — you need tokens for both instances:

- Source token: `api` scope, any role (read-only)
- Dest token: `api` scope, **Owner** on groups, **Maintainer** on projects

Set them as environment variables before running:

```bash
export SOURCE_GITLAB_TOKEN=glpat-xxxx
export DEST_GITLAB_TOKEN=glpat-yyyy
```

These variable names must match `source.token_env` and `destination.token_env` in your config file.

---

## Build

Requires Go 1.22+.

```bash
go mod tidy
go build -o gitlab-copy ./cmd/
```

Cross-platform builds via Makefile:

```bash
make all          # builds for linux/amd64, darwin/arm64, windows/amd64
make linux        # linux/amd64 only
make mac-arm      # darwin/arm64 only
make windows      # windows/amd64 only
```

---

## Quick Start

**Step 1 — Copy and edit the config:**

```bash
cp config.yaml my-batch.yaml
# Edit my-batch.yaml — set source/dest URLs, group name
```

**Step 2 — Dry-run first. Always:**

```bash
./gitlab-copy all -config my-batch.yaml -group my-group -dry-run
```

Open the HTML report in `./gl-copy-report/` and review what would change. Nothing is written in dry-run mode.

**Step 3 — Apply:**

```bash
./gitlab-copy all -config my-batch.yaml -group my-group
```

**Step 4 — Validate:**

```bash
./gitlab-diff -config my-batch.yaml -group my-group
```

---

## Migration Workflow

Each migration batch follows this four-step sequence:

```
1. Congregate     →  migrates repository content (branches, commits, MRs, wikis)
2. gitlab-copy    →  dry-run: preview what settings will be copied
3. gitlab-copy    →  apply: copy the settings
4. gitlab-diff    →  validate: confirm source and dest match
```

### Batch 1 — New group (first batch for this group)

The group has never been migrated. Run `all` to copy both group-level settings and all project settings:

```bash
# Dry-run
./gitlab-copy all -config config.yaml -group group1 -dry-run

# Apply
./gitlab-copy all -config config.yaml -group group1
```

### Batch 2+ — Existing group (additional APPID batches)

The group already exists on dest with correct settings from Batch 1. Only copy project settings — skip group domains to avoid unnecessary API calls:

```bash
# Dry-run
./gitlab-copy projects all -config config.yaml -group group1 -dry-run

# Apply
./gitlab-copy projects all -config config.yaml -group group1
```

### Single project

Useful for re-running a specific project after a fix, or for one-off migrations:

```bash
./gitlab-copy projects all -config config.yaml -project group1/fx-posting-soap -dry-run
./gitlab-copy projects all -config config.yaml -project group1/fx-posting-soap
```

No group needs to be specified in the config for single-project runs.

---

## Configuration

The config file controls source/dest connections, which groups and projects to target, and which domains to copy. Use `config.yaml` as your starting point and create a copy per migration batch if needed.

```yaml
source:
  url: https://source.example.com
  token_env: SOURCE_GITLAB_TOKEN      # name of the env var holding the token

destination:
  url: https://dest.example.com
  token_env: DEST_GITLAB_TOKEN

groups:
  include:
    - group1                      # group path(s) to process
  exclude:
    - group1/dast/*               # exclude all descendants of group1/dast
    - group1/dast_rest_scan/*
  include_subgroups: true

projects:
  include: []                         # leave empty to derive from groups above
  exclude:
    - group1/OBSOLETE-*           # exclude by name pattern
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
    - default_branch_protection
    - mr_settings
    - mr_approval_settings
    - protected_environments
    - approval_rules
    - jira_integration
    - badges
    - compliance_frameworks
    - compliance_assignments
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

---

## Controlling What Gets Copied

There are two independent controls:

### 1. Domains — which settings categories to copy

The `domains` section in config.yaml is the primary control. **Comment out any domain you want to skip entirely.**

Example — skip Jira integration for all projects in this run:

```yaml
domains:
  projects:
    - topics
    - environments
    # - jira_integration   ← commented out, won't run
    - pipeline_triggers
```

Domains run in the order listed. For compliance, `compliance_frameworks` must appear before `compliance_assignments`.

### 2. Scope — which groups and projects to target

Control scope through a combination of config and CLI flags.

**Exclude specific subgroups or projects:**

```yaml
groups:
  exclude:
    - group1/dast/*              # skip all dast projects
    - group1/dast_rest_scan/*    # skip all dast_rest_scan projects

projects:
  exclude:
    - group1/OBSOLETE-*          # skip projects prefixed with OBSOLETE-
```

**Exclusion pattern syntax:**

| Pattern | What it matches |
|---|---|
| `group1/dast` | Exact path only |
| `group1/dast_*` | Single-level glob — direct children matching the pattern |
| `group1/dast/*` | Deep glob — all descendants at any depth below `group1/dast` |

**Target a specific group at runtime (overrides config):**

```bash
./gitlab-copy all -config config.yaml -group group1
```

**Target a single project at runtime:**

```bash
./gitlab-copy projects all -config config.yaml -project group1/fx-posting-soap
```

**Run only group-level domains (no projects):**

```bash
./gitlab-copy groups all -config config.yaml -group group1
```

**Run only project-level domains (no group settings):**

```bash
./gitlab-copy projects all -config config.yaml -group group1
```

---

## CLI Reference

### Command structure

```
./gitlab-copy <scope> [flags]
```

**Scopes:**

| Scope | What runs |
|---|---|
| `all` | Group domains first, then all project domains |
| `groups all` | Group domains only — no projects touched |
| `projects all` | Project domains only — no group settings touched |

**Flags:**

| Flag | Description |
|---|---|
| `-config <path>` | Path to config file (default: `config.yaml`) |
| `-group <path>` | Target a single group, overrides `groups.include` in config |
| `-project <path>` | Target a single project path, no group config required |
| `-dry-run` | Preview mode — no writes made, shows what would change |
| `-no-color` | Disable ANSI color in terminal output (useful in CI) |

**Exit codes:**

| Code | Meaning |
|---|---|
| `0` | Clean run — no failures |
| `1` | One or more domains failed — review the report |

---

### All commands

#### Dry-run (preview) — always run this first

```bash
./gitlab-copy all -config config.yaml -group my-group -dry-run
```

Reads source and dest, shows exactly what would be created/updated/skipped — no writes made. Always do this before applying. Review the HTML report before proceeding.

---

#### Copy everything for a new group

```bash
./gitlab-copy all -config config.yaml -group my-group
```

Runs all group domains followed by all project domains for every project under `my-group`. Use this on **Batch 1** when the group has never been migrated to dest.

---

#### Copy projects only (group already set up)

```bash
./gitlab-copy projects all -config config.yaml -group my-group
```

Skips all group-level domains entirely. Runs only project domains for every project under `my-group`. Use this on **Batch 2+** when the group was already configured in a previous batch.

---

#### Copy group settings only (no projects)

```bash
./gitlab-copy groups all -config config.yaml -group my-group
```

Runs group domains only — push rules, MR settings, approval settings, etc. No project domains run at all. Use this when you need to re-apply or fix group-level settings without touching projects.

---

#### Copy a single project

```bash
./gitlab-copy projects all -config config.yaml -project my-group/my-project
```

Targets one specific project. No `-group` flag needed, no `groups.include` required in config. Useful for:
- Re-running a specific project after an error
- Migrating a project that was missed in a previous batch
- Testing settings on one project before running the full batch

---

#### Dry-run a single project

```bash
./gitlab-copy projects all -config config.yaml -project my-group/my-project -dry-run
```

Preview what would change for one project only. Good for investigating a specific issue before committing.

---

#### Copy everything using config file only (no CLI group override)

```bash
./gitlab-copy all -config my-batch-config.yaml
```

Uses whatever `groups.include` is set in the config file. Useful when you have a batch-specific config that already lists the groups to process and you want to run it without any CLI overrides.

---

#### Run with no terminal color (for CI pipelines)

```bash
./gitlab-copy all -config config.yaml -group my-group -no-color
```

Disables ANSI color codes. Use in CI environments where colored output produces garbled logs.

---

#### Dry-run with no color (CI preview step)

```bash
./gitlab-copy all -config config.yaml -group my-group -dry-run -no-color
```

Preview mode for CI. Combine with artifact archiving of the JSON report for audit trails.

---

#### Re-run after fixing a failure

If a previous run had failures, fix the root cause and re-run the same command. The tool is idempotent — items that already succeeded will be `Skipped`, only the previously failed items will be retried.

```bash
# Same command as before — safe to re-run
./gitlab-copy projects all -config config.yaml -group my-group
```

---

### Combining scopes and config for common scenarios

**Run only specific domains for a batch:**

Edit config.yaml to comment out domains you don't want, or create a batch-specific config:

```yaml
# batch-3-push-rules-only.yaml
domains:
  projects:
    - project_push_rules    # only copy push rules this run
```

```bash
./gitlab-copy projects all -config batch-3-push-rules-only.yaml -group my-group
```

**Exclude specific subgroups:**

```yaml
# config.yaml
groups:
  exclude:
    - my-group/dast/*
    - my-group/dast_rest_scan/*
```

```bash
./gitlab-copy all -config config.yaml -group my-group
```

Projects under `dast` and `dast_rest_scan` will be enumerated but skipped due to the exclusion rules.

---

## Reading the Report

Each run writes a report to `./gl-copy-report/` (configurable). Three formats are produced:

- **HTML** — open in a browser; tabbed Groups/Projects view with collapsible cards
- **JSON** — machine-readable; suitable for pipeline integration or scripting
- **Terminal** — printed live during the run

### Action labels

| Label | Meaning |
|---|---|
| `Created` | Item didn't exist on dest — was created |
| `Updated` | Item existed but differed — was updated. Diff lines show what changed |
| `Skipped` | Item already matches on dest — nothing to do |
| `Failed` | Write attempt failed — error message shown, investigate before re-running |
| `DryRun(Create)` | Would be created in a real run |
| `DryRun(Update)` | Would be updated — diff lines show what would change |
| `DryRun(Skip)` | Already matches — would be skipped |

### Understanding warnings on Skipped items

Some items are reported as `Skipped` with a warning message rather than `Failed`. This is intentional — it means the tool detected a condition where attempting the write would definitely fail or produce incorrect results. Common examples:

- **Jira integration** — `credentials masked in source API response` — the Jira password/token cannot be read from the API. Configure Jira manually on dest.
- **any-approver rule already exists** — dest already has an any-approver approval rule. Only one is allowed per project; the duplicate is correctly skipped.
- **user/group-specific access levels not copied** — a protected branch or tag has user- or group-specific access rules. These IDs are instance-specific and can't transfer. Role-based rules are copied; user/group rules need manual follow-up.

### Updated items with diffs

When a setting differs between source and dest, the Updated item shows a diff:

```
master — Updated
  merge_access_levels    dest: Developers+Maintainers  → source: Maintainers
  push_access_levels     dest: Developers+Maintainers  → source: Maintainers
```

Red shows the current dest value. Green shows the source value that will be applied. In dry-run mode the same diff is shown so you can review before committing.

---

## Domains Reference

### Group domains

| Domain | What it copies | Notes |
|---|---|---|
| `push_rules` | Commit message regex, branch name rules, author email regex, file size limits | |
| `description` | Group description text | |
| `default_branch_name` | Default branch name for new projects in the group | |
| `default_branch_protection` | Default branch protection applied to new projects (who can push/merge, force push rules) | |
| `mr_settings` | Merge-if-pipeline-succeeds, resolve-all-discussions, Jira issue requirement | |
| `mr_approval_settings` | Author/committer approval, override permissions, retain approvals on push | |
| `protected_environments` | Role-based deploy access per named environment | |
| `approval_rules` | MR approval rule names and required approver counts | Approvers must be assigned manually after copy |
| `jira_integration` | Jira integration configuration | Credentials are masked in source API — verify password/token on dest manually |
| `badges` | Group-level badges inherited by all projects in the group | Badge URLs may reference source instance — verify after copy |
| `compliance_frameworks` | Compliance framework definitions (name, description, color, pipeline config path) | Pipeline config path must exist on dest |
| `compliance_assignments` | Which projects have which compliance frameworks assigned | Must run after `compliance_frameworks` |

### Project domains

| Domain | What it copies | Notes |
|---|---|---|
| `topics` | Project topics/tags (AppID, CMDB ID, etc.) | |
| `environments` | Deployment environments with external URLs | Environment state (stopped/active) is not copied |
| `protected_environments` | Role-based deploy access per environment | |
| `jira_integration` | Jira integration configuration | Credentials masked — verify manually on dest |
| `pipeline_triggers` | CI pipeline trigger tokens | New token is generated on dest — update any CI variable references |
| `deploy_keys` | SSH deploy keys | Globally registered keys may need manual enabling on dest |
| `project_push_rules` | Project-level commit and branch push rules | |
| `project_mr_approvals` | Project-level MR approval settings | |
| `project_approval_rules` | Project-level MR approval rules | Approvers must be assigned manually; any-approver duplicates are skipped gracefully |
| `badges` | Project-level pipeline status and coverage badges | |
| `project_protected_branches` | Branch protection rules | User/group-specific access levels are not copied — role-based only |
| `project_protected_tags` | Tag protection rules | User/group-specific access levels are not copied — role-based only |

---

## What Cannot Be Automated

These items cannot be copied by the tool. They require manual action on the dest instance after migration.

| Item | Why | Action required |
|---|---|---|
| **Security policies** | No API exists for linking a security policy project | Configure via GitLab UI → Group → Security → Policies |
| **Group/project access tokens** | Token value is only shown once at creation — cannot be read back | Create new tokens on dest with the same name and scopes |
| **Deploy tokens** | Token secret cannot be exported | Regenerate on dest and update any services using them |
| **Masked/hidden variables** | Values are masked in API responses — reads return empty | Create variables manually on dest with the correct values |
| **Jira credentials** | Password/token are masked in source API response | After copy, open the Jira integration on dest and re-enter credentials |
| **MR approval rule approvers** | User IDs are instance-specific — they don't map across instances | After rules are created, assign approvers manually on dest |
| **User/group access on protected branches/tags** | Same reason — user and group IDs don't transfer | After protection is applied, add specific user/group access manually |

---

## FAQ

**Q: Do I need to run gitlab-copy every time I run a new Congregate batch for the same group?**

For most batches, yes — but use `projects all` rather than `all` once the group is set up:

```bash
# Batch 1 — group is new
./gitlab-copy all -config config.yaml -group my-group

# Batch 2, 3, ... — group already exists
./gitlab-copy projects all -config config.yaml -group my-group
```

Running `projects all` skips all group-level domains and only processes project settings for the projects in scope.

---

**Q: What happens if I run gitlab-copy twice on the same project?**

It is safe to re-run. The tool is idempotent — if a setting already matches on dest it is reported as `Skipped` and no write is made. If something changed since the last run it will be updated.

---

**Q: A project isn't on dest yet. Will gitlab-copy fail?**

For domains that target the project directly (topics, push rules, etc.) those will fail with a `404 Project Not Found` error. This is expected — run gitlab-copy after Congregate has migrated the project. For domains that don't require the project to exist on dest (environments is one edge case), they may partially succeed.

---

**Q: I see `Failed` on jira_integration for every project. What's wrong?**

This was a known issue that is now fixed. The tool detects that Jira credentials are masked in the source API response and reports the item as `Skipped` with a warning rather than attempting a write that would fail. If you're seeing actual `Failed` status on Jira integration, check that your source token has sufficient permissions to read integrations.

---

**Q: I want to skip a specific domain for one batch without editing config.yaml permanently.**

Comment it out in a batch-specific copy of the config:

```bash
cp config.yaml batch-3.yaml
# edit batch-3.yaml, comment out the domain
./gitlab-copy all -config batch-3.yaml -group my-group
```

---

**Q: Can I run just one domain to fix a specific issue?**

Yes — edit the config to include only that domain:

```yaml
domains:
  projects:
    - project_protected_branches   # only run this
```

Or create a minimal config file for that purpose.

---

**Q: The report shows `Updated` on protected branches every time I run it, but nothing looks wrong.**

Check whether the branch has user- or group-specific access levels on source. These can't be transferred (user/group IDs are instance-specific) so the tool copies only the role-based levels. On subsequent runs, if the role-based levels match, the branch will be `Skipped`. If it keeps showing `Updated`, there is a genuine difference in the role-based access levels between source and dest — look at the diff lines in the report to see exactly what differs.

---

**Q: I got `any-approver rule already exists on dest — skipped`. Is that a problem?**

No — this is correct behavior. GitLab only allows one `any-approver` approval rule per project. If dest already has one (typically named `All Members`), the source rule with a different name but the same type is skipped rather than creating a duplicate that GitLab would reject. Review the existing rule on dest to confirm it has the correct approver count.

---

**Q: How do I know if the migration is complete?**

Run gitlab-diff after gitlab-copy:

```bash
./gitlab-diff -config config.yaml -group my-group
```

A clean gitlab-diff report (all green, no diffs) means source and dest match on all tracked settings. Remaining red items in gitlab-diff after gitlab-copy are either in the "cannot be automated" list above, or indicate something that needs investigation.

---

**Q: compliance_frameworks ran but compliance_assignments failed with "framework not found on dest".**

This means `compliance_frameworks` ran but the framework wasn't created — likely because it failed silently or the pipeline config path was invalid. Check the gitlab-copy report for the `compliance_frameworks` domain. Fix any errors there and re-run — `compliance_assignments` will pick up the framework IDs on the next run.

---

**Q: Can I run this in a CI pipeline?**

Yes. Use `-no-color` to disable ANSI codes in CI logs, and check the exit code:

```yaml
# GitLab CI example
gitlab-copy:
  script:
    - ./gitlab-copy all -config config.yaml -group $GROUP -no-color
  # Exit code 1 = failures, which will fail the job
```

The JSON report in `./gl-copy-report/` can be archived as a CI artifact for auditing.