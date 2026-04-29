package copy

import (
	"fmt"
	"strings"

	"gitlab-copy/internal"
	"gitlab-copy/internal/gitlab"
)

// fieldDiff always adds a DiffLine. Match=true means values are equal (shown muted),
// Match=false means they differ (shown highlighted).
func fieldDiff(diffs *[]internal.DiffLine, field string, src, dst any) {
	s := fmt.Sprintf("%v", src)
	d := fmt.Sprintf("%v", dst)
	*diffs = append(*diffs, internal.DiffLine{
		Field: field,
		Src:   s,
		Dst:   d,
		Match: s == d,
	})
}

// hasChanges returns true if any DiffLine has Match=false (i.e. src != dst).
func hasChanges(diffs []internal.DiffLine) bool {
	for _, d := range diffs {
		if !d.Match {
			return true
		}
	}
	return false
}
func pushRuleDiffs(src, dst *gitlab.PushRule) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "commit_message_regex", src.CommitMessageRegex, dst.CommitMessageRegex)
	fieldDiff(&diffs, "commit_message_negative_regex", src.CommitMessageNegativeRegex, dst.CommitMessageNegativeRegex)
	fieldDiff(&diffs, "branch_name_regex", src.BranchNameRegex, dst.BranchNameRegex)
	fieldDiff(&diffs, "author_email_regex", src.AuthorEmailRegex, dst.AuthorEmailRegex)
	fieldDiff(&diffs, "file_name_regex", src.FileNameRegex, dst.FileNameRegex)
	fieldDiff(&diffs, "max_file_size", src.MaxFileSize, dst.MaxFileSize)
	fieldDiff(&diffs, "deny_delete_tag", src.DenyDeleteTag, dst.DenyDeleteTag)
	fieldDiff(&diffs, "member_check", src.MemberCheck, dst.MemberCheck)
	fieldDiff(&diffs, "prevent_secrets", src.PreventSecrets, dst.PreventSecrets)
	fieldDiff(&diffs, "commit_committer_check", src.CommitCommitterCheck, dst.CommitCommitterCheck)
	fieldDiff(&diffs, "commit_committer_name_check", src.CommitCommitterNameCheck, dst.CommitCommitterNameCheck)
	fieldDiff(&diffs, "reject_unsigned_commits", src.RejectUnsignedCommits, dst.RejectUnsignedCommits)
	fieldDiff(&diffs, "reject_non_dco_commits", src.RejectNonDCOCommits, dst.RejectNonDCOCommits)
	return diffs
}

// mrSettingsDiffs returns diff lines for group MR settings fields.
func mrSettingsDiffs(src, dst *gitlab.Group) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "only_allow_merge_if_pipeline_succeeds", src.OnlyAllowMergeIfPipelineSucceeds, dst.OnlyAllowMergeIfPipelineSucceeds)
	fieldDiff(&diffs, "only_allow_merge_if_all_discussions_are_resolved", src.OnlyAllowMergeIfAllDiscussionsAreResolved, dst.OnlyAllowMergeIfAllDiscussionsAreResolved)
	if src.PreventMergeWithoutJiraIssue != nil || dst.PreventMergeWithoutJiraIssue != nil {
		srcVal := fmt.Sprintf("%v", src.PreventMergeWithoutJiraIssue)
		dstVal := fmt.Sprintf("%v", dst.PreventMergeWithoutJiraIssue)
		if srcVal != dstVal {
			diffs = append(diffs, internal.DiffLine{Field: "prevent_merge_without_jira_issue", Src: srcVal, Dst: dstVal})
		}
	}
	return diffs
}

// mrApprovalSettingsDiffs returns diff lines for group MR approval settings.
func mrApprovalSettingsDiffs(src, dst *gitlab.MergeRequestApprovalSettings) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "allow_author_approval", src.AllowAuthorApproval.Value, dst.AllowAuthorApproval.Value)
	fieldDiff(&diffs, "allow_committer_approval", src.AllowCommitterApproval.Value, dst.AllowCommitterApproval.Value)
	fieldDiff(&diffs, "allow_overrides_to_approver_list_per_merge_request", src.AllowOverridesToApproverList.Value, dst.AllowOverridesToApproverList.Value)
	fieldDiff(&diffs, "require_password_to_approve", src.RequirePasswordToApprove.Value, dst.RequirePasswordToApprove.Value)
	fieldDiff(&diffs, "retain_approvals_on_push", src.RetainApprovalsOnPush.Value, dst.RetainApprovalsOnPush.Value)
	fieldDiff(&diffs, "selective_code_owner_removals", src.SelectiveCodeOwnerRemovals.Value, dst.SelectiveCodeOwnerRemovals.Value)
	return diffs
}

// projectMRApprovalsDiffs returns diff lines for project MR approval settings.
func projectMRApprovalsDiffs(src, dst *gitlab.ProjectApprovalSettings) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "approvals_before_merge", src.ApprovalsBeforeMerge, dst.ApprovalsBeforeMerge)
	fieldDiff(&diffs, "reset_approvals_on_push", src.ResetApprovalsOnPush, dst.ResetApprovalsOnPush)
	fieldDiff(&diffs, "selective_code_owner_removals", src.SelectiveCodeOwnerRemovals, dst.SelectiveCodeOwnerRemovals)
	fieldDiff(&diffs, "disable_overriding_approvers_per_merge_request", src.DisableOverridingApproversPerMergeRequest, dst.DisableOverridingApproversPerMergeRequest)
	fieldDiff(&diffs, "merge_requests_author_approval", src.MergeRequestsAuthorApproval, dst.MergeRequestsAuthorApproval)
	fieldDiff(&diffs, "merge_requests_disable_committers_approval", src.MergeRequestsDisableCommittersApproval, dst.MergeRequestsDisableCommittersApproval)
	fieldDiff(&diffs, "require_password_to_approve", src.RequirePasswordToApprove, dst.RequirePasswordToApprove)
	return diffs
}

// protectedBranchDiffs returns diff lines for a protected branch comparison.
// It only compares role-based access levels since user/group IDs are instance-specific.
func protectedBranchDiffs(src, dst gitlab.ProtectedBranch) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "allow_force_push", src.AllowForcePush, dst.AllowForcePush)
	fieldDiff(&diffs, "code_owner_approval_required", src.CodeOwnerApprovalRequired, dst.CodeOwnerApprovalRequired)

	srcPush := accessLevelDesc(src.PushAccessLevels)
	dstPush := accessLevelDesc(dst.PushAccessLevels)
	if srcPush != dstPush {
		diffs = append(diffs, internal.DiffLine{Field: "push_access_levels", Src: srcPush, Dst: dstPush})
	}

	srcMerge := accessLevelDesc(src.MergeAccessLevels)
	dstMerge := accessLevelDesc(dst.MergeAccessLevels)
	if srcMerge != dstMerge {
		diffs = append(diffs, internal.DiffLine{Field: "merge_access_levels", Src: srcMerge, Dst: dstMerge})
	}

	srcUnprotect := accessLevelDesc(src.UnprotectAccessLevels)
	dstUnprotect := accessLevelDesc(dst.UnprotectAccessLevels)
	if srcUnprotect != dstUnprotect {
		diffs = append(diffs, internal.DiffLine{Field: "unprotect_access_levels", Src: srcUnprotect, Dst: dstUnprotect})
	}

	return diffs
}

// protectedTagDiffs returns diff lines for a protected tag comparison.
func protectedTagDiffs(src, dst gitlab.ProtectedTag) []internal.DiffLine {
	var diffs []internal.DiffLine
	srcCreate := accessLevelDesc(src.CreateAccessLevels)
	dstCreate := accessLevelDesc(dst.CreateAccessLevels)
	if srcCreate != dstCreate {
		diffs = append(diffs, internal.DiffLine{Field: "create_access_levels", Src: srcCreate, Dst: dstCreate})
	}
	return diffs
}

// defaultBranchProtectionDiffs returns diff lines for group default branch protection.
func defaultBranchProtectionDiffs(src, dst *gitlab.Group) []internal.DiffLine {
	var diffs []internal.DiffLine
	fieldDiff(&diffs, "default_branch_protection", src.DefaultBranchProtection, dst.DefaultBranchProtection)

	// Compare defaults struct if both have it
	sd := src.DefaultBranchProtectionDefaults
	dd := dst.DefaultBranchProtectionDefaults
	if sd != nil || dd != nil {
		srcForcePush := false
		dstForcePush := false
		srcDevInitial := false
		dstDevInitial := false
		if sd != nil {
			srcForcePush = sd.AllowForcePush
			srcDevInitial = sd.DeveloperCanInitialPush
		}
		if dd != nil {
			dstForcePush = dd.AllowForcePush
			dstDevInitial = dd.DeveloperCanInitialPush
		}
		fieldDiff(&diffs, "default_branch_protection_defaults.allow_force_push", srcForcePush, dstForcePush)
		fieldDiff(&diffs, "default_branch_protection_defaults.developer_can_initial_push", srcDevInitial, dstDevInitial)

		srcPush := "[]"
		dstPush := "[]"
		srcMerge := "[]"
		dstMerge := "[]"
		if sd != nil {
			srcPush = fmt.Sprintf("%v", sd.AllowedToPush)
			srcMerge = fmt.Sprintf("%v", sd.AllowedToMerge)
		}
		if dd != nil {
			dstPush = fmt.Sprintf("%v", dd.AllowedToPush)
			dstMerge = fmt.Sprintf("%v", dd.AllowedToMerge)
		}
		if srcPush != dstPush {
			diffs = append(diffs, internal.DiffLine{Field: "default_branch_protection_defaults.allowed_to_push", Src: srcPush, Dst: dstPush})
		}
		if srcMerge != dstMerge {
			diffs = append(diffs, internal.DiffLine{Field: "default_branch_protection_defaults.allowed_to_merge", Src: srcMerge, Dst: dstMerge})
		}
	}
	return diffs
}
func accessLevelDesc(levels []gitlab.BranchAccessLevel) string {
	var parts []string
	for _, l := range levels {
		if l.IsRoleBased() {
			parts = append(parts, l.AccessLevelDescription)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
