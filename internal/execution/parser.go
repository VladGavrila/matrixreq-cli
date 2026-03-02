package execution

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ParseYAMLResults parses a YAML execution results file.
func ParseYAMLResults(filePath string) (*ExecutionResults, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading results file: %w", err)
	}
	return ParseYAMLResultsFromString(string(data))
}

// ParseYAMLResultsFromString parses YAML execution results from a string.
func ParseYAMLResultsFromString(content string) (*ExecutionResults, error) {
	// Handle YAML front matter (content between --- markers)
	re := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	match := re.FindStringSubmatch(content)
	yamlContent := content
	if len(match) >= 2 {
		yamlContent = match[1]
	}

	var results ExecutionResults
	if err := yaml.Unmarshal([]byte(yamlContent), &results); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if err := ValidateResults(&results); err != nil {
		return nil, err
	}

	// Apply defaults
	if results.ExecutionDate == "" {
		results.ExecutionDate = time.Now().Format("2006-01-02")
	}
	if results.Tester == "" {
		results.Tester = "automation"
	}

	// Normalize status values to uppercase
	for i := range results.Results {
		results.Results[i].Result = strings.ToUpper(results.Results[i].Result)
		for j := range results.Results[i].Steps {
			results.Results[i].Steps[j].Status = strings.ToUpper(results.Results[i].Steps[j].Status)
		}
	}

	return &results, nil
}

// ValidateResults validates the structure of parsed execution results.
func ValidateResults(results *ExecutionResults) error {
	if len(results.Results) == 0 {
		return fmt.Errorf("results list is empty")
	}

	for i, result := range results.Results {
		if result.TestName == "" {
			return fmt.Errorf("result %d: missing test_name", i)
		}
		status := strings.ToUpper(result.Result)
		if status != "PASS" && status != "FAIL" {
			return fmt.Errorf("result %d (%s): result must be PASS or FAIL, got %q", i, result.TestName, result.Result)
		}
		if len(result.Steps) == 0 {
			return fmt.Errorf("result %d (%s): steps list is empty", i, result.TestName)
		}
		for j, step := range result.Steps {
			if step.Actual == "" {
				return fmt.Errorf("result %d (%s) step %d: missing actual", i, result.TestName, j)
			}
			stepStatus := strings.ToUpper(step.Status)
			if stepStatus != "PASS" && stepStatus != "FAIL" {
				return fmt.Errorf("result %d (%s) step %d: status must be PASS or FAIL, got %q", i, result.TestName, j, step.Status)
			}
		}
	}

	return nil
}
