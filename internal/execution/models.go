package execution

// ExecutionResultStep is a single step result from a YAML results file.
type ExecutionResultStep struct {
	Requirement string `yaml:"requirement"`
	Actual      string `yaml:"actual"`
	Status      string `yaml:"status"` // "PASS" or "FAIL"
}

// ExecutionResultTest is a single test result from a YAML results file.
type ExecutionResultTest struct {
	TestName string                `yaml:"test_name"`
	Result   string                `yaml:"result"` // Overall: "PASS" or "FAIL"
	Steps    []ExecutionResultStep `yaml:"steps"`
}

// ExecutionResults holds parsed YAML execution results.
type ExecutionResults struct {
	ExecutionDate string                `yaml:"execution_date"`
	Tester        string                `yaml:"tester"`
	SUTVersion    string                `yaml:"sut_version"`
	Results       []ExecutionResultTest `yaml:"results"`
}

// TestStep represents a single step in a TC or XTC from the Matrix API.
type TestStep struct {
	Action          string `json:"action"`
	Expected        string `json:"expected"`
	RequirementLink string `json:"RequirementLink"`
	Result          string `json:"result,omitempty"`  // "p", "f"
	Human           string `json:"human,omitempty"`   // "passed", "failed"
	Render          string `json:"render,omitempty"`  // "ok", "error"
	Comment         string `json:"comment,omitempty"` // Execution comment
}

// UploadResult holds the outcome of an execution results upload.
type UploadResult struct {
	Successes map[string]bool // XTC ref → success
	Issues    []string        // TC names with step mismatches
}
