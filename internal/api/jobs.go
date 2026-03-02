package api

// JobWithUrl represents a job with a download URL.
type JobWithUrl struct {
	JobID      int    `json:"jobId"`
	Progress   int    `json:"progress,omitempty"`
	Status     string `json:"status,omitempty"`
	JobFile    string `json:"jobFile,omitempty"`
	JobFileURL string `json:"jobFileUrl,omitempty"`
	VisibleName string `json:"visibleName,omitempty"`
}

// JobsWithUrl wraps a list of jobs.
type JobsWithUrl struct {
	Jobs []JobWithUrl `json:"jobs,omitempty"`
}
