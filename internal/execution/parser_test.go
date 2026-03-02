package execution

import (
	"strings"
	"testing"
)

func TestParseYAMLResultsFromString(t *testing.T) {
	yaml := `execution_date: "2024-01-15"
tester: "john"
sut_version: "1.0.0"
results:
  - test_name: "TC-100"
    result: "pass"
    steps:
      - requirement: "SOFT-123"
        actual: "Value matched expected"
        status: "pass"
      - requirement: "SOFT-456"
        actual: "Status verified"
        status: "pass"
  - test_name: "TC-200"
    result: "fail"
    steps:
      - requirement: "SOFT-789"
        actual: "Expected 10 but got 5"
        status: "fail"
`
	results, err := ParseYAMLResultsFromString(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.ExecutionDate != "2024-01-15" {
		t.Errorf("execution_date: got %q", results.ExecutionDate)
	}
	if results.Tester != "john" {
		t.Errorf("tester: got %q", results.Tester)
	}
	if results.SUTVersion != "1.0.0" {
		t.Errorf("sut_version: got %q", results.SUTVersion)
	}
	if len(results.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results.Results))
	}

	// Check normalization to uppercase
	if results.Results[0].Result != "PASS" {
		t.Errorf("result 0 should be normalized to PASS, got %q", results.Results[0].Result)
	}
	if results.Results[1].Result != "FAIL" {
		t.Errorf("result 1 should be normalized to FAIL, got %q", results.Results[1].Result)
	}

	if results.Results[0].Steps[0].Status != "PASS" {
		t.Errorf("step status should be normalized to PASS, got %q", results.Results[0].Steps[0].Status)
	}
}

func TestParseYAMLResultsWithFrontMatter(t *testing.T) {
	yaml := `---
execution_date: "2024-01-15"
tester: "automation"
results:
  - test_name: "TC-100"
    result: "PASS"
    steps:
      - requirement: "SOFT-123"
        actual: "OK"
        status: "PASS"
---`

	results, err := ParseYAMLResultsFromString(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.ExecutionDate != "2024-01-15" {
		t.Errorf("execution_date: got %q", results.ExecutionDate)
	}
	if len(results.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results.Results))
	}
}

func TestParseYAMLResultsDefaults(t *testing.T) {
	yaml := `results:
  - test_name: "TC-100"
    result: "PASS"
    steps:
      - requirement: "SOFT-1"
        actual: "OK"
        status: "PASS"
`
	results, err := ParseYAMLResultsFromString(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should apply defaults
	if results.Tester != "automation" {
		t.Errorf("default tester: got %q, want %q", results.Tester, "automation")
	}
	if results.ExecutionDate == "" {
		t.Error("execution_date should have today's date as default")
	}
}

func TestValidateResultsEmpty(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for empty results")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention empty results: %v", err)
	}
}

func TestValidateResultsMissingTestName(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "",
				Result:   "PASS",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-1", Actual: "OK", Status: "PASS"},
				},
			},
		},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for missing test_name")
	}
}

func TestValidateResultsInvalidStatus(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "TC-100",
				Result:   "MAYBE",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-1", Actual: "OK", Status: "PASS"},
				},
			},
		},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for invalid result status")
	}
	if !strings.Contains(err.Error(), "MAYBE") {
		t.Errorf("error should mention invalid status: %v", err)
	}
}

func TestValidateResultsEmptySteps(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "TC-100",
				Result:   "PASS",
				Steps:    []ExecutionResultStep{},
			},
		},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for empty steps")
	}
}

func TestValidateResultsMissingActual(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "TC-100",
				Result:   "PASS",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-1", Actual: "", Status: "PASS"},
				},
			},
		},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for missing actual")
	}
}

func TestValidateResultsInvalidStepStatus(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "TC-100",
				Result:   "PASS",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-1", Actual: "OK", Status: "UNKNOWN"},
				},
			},
		},
	}
	err := ValidateResults(results)
	if err == nil {
		t.Error("expected error for invalid step status")
	}
}

func TestValidateResultsValid(t *testing.T) {
	results := &ExecutionResults{
		Results: []ExecutionResultTest{
			{
				TestName: "TC-100",
				Result:   "PASS",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-1", Actual: "Value matched", Status: "PASS"},
					{Requirement: "SOFT-2", Actual: "Status correct", Status: "PASS"},
				},
			},
			{
				TestName: "TC-200",
				Result:   "FAIL",
				Steps: []ExecutionResultStep{
					{Requirement: "SOFT-3", Actual: "Unexpected value", Status: "FAIL"},
				},
			},
		},
	}
	err := ValidateResults(results)
	if err != nil {
		t.Errorf("unexpected error for valid results: %v", err)
	}
}

func TestParseYAMLResultsInvalidYAML(t *testing.T) {
	_, err := ParseYAMLResultsFromString("this is not valid yaml: [")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseYAMLResultsCaseInsensitiveStatus(t *testing.T) {
	yaml := `results:
  - test_name: "TC-100"
    result: "Pass"
    steps:
      - requirement: "SOFT-1"
        actual: "OK"
        status: "pass"
`
	results, err := ParseYAMLResultsFromString(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.Results[0].Result != "PASS" {
		t.Errorf("result should be PASS, got %q", results.Results[0].Result)
	}
	if results.Results[0].Steps[0].Status != "PASS" {
		t.Errorf("step status should be PASS, got %q", results.Results[0].Steps[0].Status)
	}
}
