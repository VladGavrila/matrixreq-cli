// Standalone mxreq execution result recorder for Go.
// Copy this file into your test repository.
//
// Usage:
//
//	import "yourproject/mxreq"
//
//	func TestExample(t *testing.T) {
//	    mxreq.StartTest("")  // Auto-detects test name from function
//	    // Or explicitly: mxreq.StartTest("TestExample")
//	    defer mxreq.EndTest(t)
//
//	    // Staging step - not recorded (no requirement)
//	    data := setup()
//
//	    // Verification step - recorded
//	    result := calculate(data)
//	    mxreq.VerifyEqual("SOFT-123", result, 10, false)
//	}
//
// Results are written incrementally - each EndTest() writes a valid YAML file.

package mxreq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// defaultOutputFile is generated once at package initialization
// Uses date only (YYYYMMDD) to ensure all tests in a run write to the same file
var defaultOutputFile = fmt.Sprintf("results_%s.yaml", time.Now().Format("20060102"))

type stepResult struct {
	Actual      string
	Status      string
	Requirement string
	Line        int
}

type testResult struct {
	TestName string
	Result   string
	Steps    []stepResult
}

type recorder struct {
	mu          sync.Mutex
	currentTest *testResult
	tester      string
	version     string
	outputFile  string
}

var globalRecorder = &recorder{tester: "automation"}

// Configure sets global metadata for all tests.
func Configure(tester, version, outputFile string) {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()
	globalRecorder.tester = tester
	globalRecorder.version = version
	if outputFile != "" {
		globalRecorder.outputFile = outputFile
	}
}

// StartTest begins recording a test.
// If testName is empty, auto-detects the calling function name.
// Extracts only the TC-{number} portion from the test name for matching with Matrix XTCs.
func StartTest(testName string) {
	if testName == "" {
		// Auto-detect from caller's function name
		pc, _, _, ok := runtime.Caller(1)
		if !ok {
			panic("Could not auto-detect test name")
		}
		fullName := runtime.FuncForPC(pc).Name()
		// Extract function name from full path (e.g., "package.TestExample" -> "TestExample")
		parts := strings.Split(fullName, ".")
		testName = parts[len(parts)-1]
		// Remove any suffix like "-fm" from method values
		testName = strings.Split(testName, "-")[0]
	}

	// Extract only the TC-{number} portion from the test name
	tcPattern := regexp.MustCompile(`(?i)TC[-_]\d+`)
	if match := tcPattern.FindString(testName); match != "" {
		// Normalize to TC-{number} format (uppercase, dash separator)
		testName = strings.ToUpper(strings.ReplaceAll(match, "_", "-"))
	}

	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()
	globalRecorder.currentTest = &testResult{TestName: testName, Result: "PASS", Steps: []stepResult{}}
}

func recordStep(requirement, actual, status string) {
	if requirement == "" {
		return // Skip staging steps
	}

	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	if globalRecorder.currentTest == nil {
		panic("No active test - call StartTest() first")
	}

	// Get caller's line number
	_, _, line, _ := runtime.Caller(2)

	step := stepResult{Actual: actual, Status: status, Line: line}
	if requirement != "" {
		step.Requirement = requirement
	}
	globalRecorder.currentTest.Steps = append(globalRecorder.currentTest.Steps, step)

	// Update overall test result if any step fails
	if status == "FAIL" {
		globalRecorder.currentTest.Result = "FAIL"
	}
}

// EndTest finishes recording the current test and writes results to file.
// Test completes fully, results are written, then test fails if any requirements failed.
func EndTest(t *testing.T) {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	if globalRecorder.currentTest == nil || len(globalRecorder.currentTest.Steps) == 0 {
		globalRecorder.currentTest = nil
		return
	}

	// Determine output file
	outputFile := globalRecorder.outputFile
	if outputFile == "" {
		outputFile = defaultOutputFile
	}

	// Step 1: Write results to file (always happens first)
	writeTestToFile(outputFile, globalRecorder.currentTest, globalRecorder.tester, globalRecorder.version)

	// Step 2: After writing results, fail the test if any requirements failed
	if globalRecorder.currentTest.Result == "FAIL" {
		var failedReqs []string
		for _, step := range globalRecorder.currentTest.Steps {
			if step.Status == "FAIL" && step.Requirement != "" {
				failedReqs = append(failedReqs, fmt.Sprintf("%s (line %d)", step.Requirement, step.Line))
			}
		}
		globalRecorder.currentTest = nil
		t.Errorf("Failed requirements: %s", strings.Join(failedReqs, ", "))
		return
	}

	globalRecorder.currentTest = nil
}

// writeTestToFile writes a single test to file, appending if file exists.
func writeTestToFile(outputFile string, test *testResult, tester, version string) error {
	// Check if file exists
	if _, err := os.Stat(outputFile); err == nil {
		// File exists - read, remove trailing ---, append test, write back
		content, err := os.ReadFile(outputFile)
		if err != nil {
			return err
		}

		contentStr := strings.TrimSuffix(string(content), "---\n")

		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString(contentStr)
		writeTestEntry(f, test)
		fmt.Fprintln(f, "---")
	} else {
		// File doesn't exist - create with header
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()

		fmt.Fprintln(f, "---")
		fmt.Fprintf(f, "execution_date: '%s'\n", time.Now().Format("2006-01-02"))
		fmt.Fprintf(f, "tester: %s\n", tester)
		if version != "" {
			fmt.Fprintf(f, "sut_version: %s\n", version)
		}
		fmt.Fprintln(f, "results:")
		writeTestEntry(f, test)
		fmt.Fprintln(f, "---")
	}

	return nil
}

// writeTestEntry writes a single test entry to the file.
func writeTestEntry(f io.Writer, test *testResult) {
	fmt.Fprintf(f, "- test_name: %s\n", test.TestName)
	fmt.Fprintf(f, "  result: %s\n", test.Result)
	fmt.Fprintln(f, "  steps:")
	for _, step := range test.Steps {
		actualJSON, _ := json.Marshal(step.Actual)
		fmt.Fprintf(f, "  - actual: %s\n", actualJSON)
		fmt.Fprintf(f, "    status: %s\n", step.Status)
		if step.Requirement != "" {
			fmt.Fprintf(f, "    requirement: %s\n", step.Requirement)
		}
	}
}

// Clear removes all recorded results.
func Clear() {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()
	globalRecorder.currentTest = nil
}

// verifyStep is an internal helper to verify a test step and record the result.
// Only steps with a non-empty requirement are recorded.
// If fatal is true and condition is false, the function panics.
func verifyStep(requirement string, condition bool, actual string, fatal bool) bool {
	status := "PASS"
	if !condition {
		status = "FAIL"
	}
	recordStep(requirement, actual, status)

	if !condition && fatal {
		panic(fmt.Sprintf("Fatal verification failed: %s", actual))
	}
	return condition
}

// VerifyEqual verifies two values are equal.
func VerifyEqual(requirement string, actual, expected interface{}, fatal bool) bool {
	condition := actual == expected
	msg := fmt.Sprintf("Expected %v, got %v", expected, actual)
	return verifyStep(requirement, condition, msg, fatal)
}

// VerifyTrue verifies a value is true.
func VerifyTrue(requirement string, value bool, message string, fatal bool) bool {
	if message == "" {
		message = fmt.Sprintf("Value is %v", value)
	}
	return verifyStep(requirement, value, message, fatal)
}

// VerifyFalse verifies a value is false.
func VerifyFalse(requirement string, value bool, message string, fatal bool) bool {
	if message == "" {
		message = fmt.Sprintf("Value is %v", value)
	}
	return verifyStep(requirement, !value, message, fatal)
}

// VerifyNotEqual verifies two values are not equal.
func VerifyNotEqual(requirement string, actual, expected interface{}, fatal bool) bool {
	condition := actual != expected
	msg := fmt.Sprintf("Expected not %v, got %v", expected, actual)
	return verifyStep(requirement, condition, msg, fatal)
}
