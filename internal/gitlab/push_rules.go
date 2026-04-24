package gitlab

// PushRule represents a group or project push rule.
type PushRule struct {
	ID                         int    `json:"id"`
	CreatedAt                  string `json:"created_at"`
	CommitMessageRegex         string `json:"commit_message_regex"`
	CommitMessageNegativeRegex string `json:"commit_message_negative_regex"`
	BranchNameRegex            string `json:"branch_name_regex"`
	DenyDeleteTag              bool   `json:"deny_delete_tag"`
	MemberCheck                bool   `json:"member_check"`
	PreventSecrets             bool   `json:"prevent_secrets"`
	AuthorEmailRegex           string `json:"author_email_regex"`
	FileNameRegex              string `json:"file_name_regex"`
	MaxFileSize                int    `json:"max_file_size"`
	CommitCommitterCheck       bool   `json:"commit_committer_check"`
	CommitCommitterNameCheck   bool   `json:"commit_committer_name_check"`
	RejectUnsignedCommits      bool   `json:"reject_unsigned_commits"`
	RejectNonDCOCommits        bool   `json:"reject_non_dco_commits"`
}

// IsEmpty returns true when a PushRule is the zero value — i.e. a 404 response
// from GitLab meaning no rules are configured.
func (pr *PushRule) IsEmpty() bool {
	return pr.CommitMessageRegex == "" &&
		pr.CommitMessageNegativeRegex == "" &&
		pr.BranchNameRegex == "" &&
		!pr.DenyDeleteTag &&
		!pr.MemberCheck &&
		!pr.PreventSecrets &&
		pr.AuthorEmailRegex == "" &&
		pr.FileNameRegex == "" &&
		pr.MaxFileSize == 0 &&
		!pr.CommitCommitterCheck &&
		!pr.CommitCommitterNameCheck &&
		!pr.RejectUnsignedCommits &&
		!pr.RejectNonDCOCommits
}

// Equal returns true when all meaningful fields match between two push rules.
func (pr *PushRule) Equal(other *PushRule) bool {
	return pr.CommitMessageRegex == other.CommitMessageRegex &&
		pr.CommitMessageNegativeRegex == other.CommitMessageNegativeRegex &&
		pr.BranchNameRegex == other.BranchNameRegex &&
		pr.DenyDeleteTag == other.DenyDeleteTag &&
		pr.MemberCheck == other.MemberCheck &&
		pr.PreventSecrets == other.PreventSecrets &&
		pr.AuthorEmailRegex == other.AuthorEmailRegex &&
		pr.FileNameRegex == other.FileNameRegex &&
		pr.MaxFileSize == other.MaxFileSize &&
		pr.CommitCommitterCheck == other.CommitCommitterCheck &&
		pr.CommitCommitterNameCheck == other.CommitCommitterNameCheck &&
		pr.RejectUnsignedCommits == other.RejectUnsignedCommits &&
		pr.RejectNonDCOCommits == other.RejectNonDCOCommits
}

// PushRuleRequest is the write body for POST/PUT. Excludes id and created_at.
type PushRuleRequest struct {
	CommitMessageRegex         string `json:"commit_message_regex"`
	CommitMessageNegativeRegex string `json:"commit_message_negative_regex"`
	BranchNameRegex            string `json:"branch_name_regex"`
	DenyDeleteTag              bool   `json:"deny_delete_tag"`
	MemberCheck                bool   `json:"member_check"`
	PreventSecrets             bool   `json:"prevent_secrets"`
	AuthorEmailRegex           string `json:"author_email_regex"`
	FileNameRegex              string `json:"file_name_regex"`
	MaxFileSize                int    `json:"max_file_size"`
	CommitCommitterCheck       bool   `json:"commit_committer_check"`
	CommitCommitterNameCheck   bool   `json:"commit_committer_name_check"`
	RejectUnsignedCommits      bool   `json:"reject_unsigned_commits"`
	RejectNonDCOCommits        bool   `json:"reject_non_dco_commits"`
}

// PushRuleRequestFrom converts a PushRule into a write request.
func PushRuleRequestFrom(pr *PushRule) PushRuleRequest {
	return PushRuleRequest{
		CommitMessageRegex:         pr.CommitMessageRegex,
		CommitMessageNegativeRegex: pr.CommitMessageNegativeRegex,
		BranchNameRegex:            pr.BranchNameRegex,
		DenyDeleteTag:              pr.DenyDeleteTag,
		MemberCheck:                pr.MemberCheck,
		PreventSecrets:             pr.PreventSecrets,
		AuthorEmailRegex:           pr.AuthorEmailRegex,
		FileNameRegex:              pr.FileNameRegex,
		MaxFileSize:                pr.MaxFileSize,
		CommitCommitterCheck:       pr.CommitCommitterCheck,
		CommitCommitterNameCheck:   pr.CommitCommitterNameCheck,
		RejectUnsignedCommits:      pr.RejectUnsignedCommits,
		RejectNonDCOCommits:        pr.RejectNonDCOCommits,
	}
}

// --- Read ---

// GetGroupPushRules fetches push rules for a group.
// Returns empty struct (not nil) on 404 — meaning no rules configured.
// Returns nil on 403 — meaning not accessible.
func (c *Client) GetGroupPushRules(groupPath string) (*PushRule, error) {
	var pr PushRule
	err := c.get("/groups/"+encodePath(groupPath)+"/push_rule", nil, &pr)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return &PushRule{}, nil
		}
		if apiErr, ok := err.(*APIError); ok && apiErr.IsForbidden() {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

// --- Write ---

// CreateGroupPushRules creates push rules on the destination group via POST.
func (c *Client) CreateGroupPushRules(groupPath string, req PushRuleRequest) error {
	return c.post("/groups/"+encodePath(groupPath)+"/push_rule", req, nil)
}

// UpdateGroupPushRules updates existing push rules on the destination group via PUT.
func (c *Client) UpdateGroupPushRules(groupPath string, req PushRuleRequest) error {
	return c.put("/groups/"+encodePath(groupPath)+"/push_rule", req, nil)
}
