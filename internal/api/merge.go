package api

// MergeParam describes a merge operation parameter.
type MergeParam struct {
	PushOrPull   string `json:"pushOrPull,omitempty"`
	Project      string `json:"project,omitempty"`
	BranchProject string `json:"branchProject,omitempty"`
	Categories   string `json:"categories,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

// MergeHistory represents a merge history entry.
type MergeHistory struct {
	MergeID    int    `json:"mergeId"`
	Date       string `json:"date,omitempty"`
	User       string `json:"user,omitempty"`
	Details    string `json:"details,omitempty"`
	PushOrPull string `json:"pushOrPull,omitempty"`
}

// BranchInfo represents branch status information.
type BranchInfo struct {
	MainProject   string `json:"mainProject,omitempty"`
	BranchProject string `json:"branchProject,omitempty"`
	BranchDate    string `json:"branchDate,omitempty"`
	BranchTag     string `json:"branchTag,omitempty"`
}
