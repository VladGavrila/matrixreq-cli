/*
mxreq YAML Docstring Generator for Go.
Copy this file into your test repository and import it.

Generates YAML-formatted docstrings and inserts them directly into source files.
Use the DS=1 environment variable to enable docstring generation mode.

Usage:

	# Generate docstrings (test code is skipped):
	DS=1 go test ./...

	# Normal test execution (Ds* calls are no-ops):
	go test ./...

Example:

	func Test_NEW_short_test_name(t *testing.T) {
	    mxreq.DsHeader("Long Test Name",
	        "Test description",
	        "F-TC-123",
	        "* Assumption 1",
	        "* Assumption 2")

	    mxreq.StartTest(t)

	    // Step 1: Setup
	    mxreq.DsStep("Description of step", "", "")
	    if !mxreq.Ds {
	        run := setup("arg1", "arg2")
	        execute(run)
	    }

	    // Step 2: User intervention required
	    mxreq.DsUser("User will perform an action", "Action has been performed", "SOFT-123")

	    // Step 3: Verify (automated)
	    mxreq.DsStep("Verify action is as per TSPEC-82", "Action is as per TSPEC", "SOFT-124")
	    if !mxreq.Ds {
	        mxreq.VerifyEqual(t, "SOFT-124", action(), "argToCompare")
	    }

	    // Step 4: Manual verification
	    mxreq.DsManual("Visually inspect the display for correct formatting", "Time is displayed in MM:SS format", "SOFT-125")

	    // Step 5: Cleanup
	    mxreq.DsStep("Discard process", "", "")
	    if !mxreq.Ds {
	        discard()
	    }

	    mxreq.DsFooter("Automated", "VM")
	    if !mxreq.Ds { mxreq.EndTest(t) }
	}
*/
package mxreq

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Ds is true when DS=1 environment variable is set.
// Use this to conditionally skip test code during docstring generation.
var Ds = os.Getenv("DS") == "1"

type testStep struct {
	Action          string
	Expected        string
	RequirementLink string
}

type docstringBuilder struct {
	mu          sync.Mutex
	title       string
	description string
	folder      string
	assumptions []string
	steps       []testStep
	labels      []string
	upLinks     map[string]bool
	sourceFile  string
	funcName    string
}

var currentBuilder *docstringBuilder

// DsHeader starts building a test case docstring.
//
// Parameters:
//   - title: Test case title as it will appear in Matrix.
//   - description: Description of the test.
//   - folder: Parent folder reference (e.g., "F-TC-521").
//   - assumptions: Variadic list of assumption/precondition strings.
func DsHeader(title, description, folder string, assumptions ...string) {
	if !Ds {
		return
	}

	// Get caller info
	pc, file, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Could not determine caller info")
		return
	}

	funcName := runtime.FuncForPC(pc).Name()
	// Extract just the function name (remove package path)
	parts := strings.Split(funcName, ".")
	funcName = parts[len(parts)-1]

	currentBuilder = &docstringBuilder{
		title:       title,
		description: description,
		folder:      folder,
		assumptions: assumptions,
		steps:       []testStep{},
		labels:      []string{},
		upLinks:     make(map[string]bool),
		sourceFile:  file,
		funcName:    funcName,
	}
}

// DsStep adds a test step to the docstring.
//
// Parameters:
//   - action: Description of the test action.
//   - expected: Expected result. Use "" for "N/A" (setup/cleanup steps).
//   - req: Requirement link(s), comma-separated (e.g., "SOFT-387").
func DsStep(action, expected, req string) {
	if !Ds || currentBuilder == nil {
		return
	}

	currentBuilder.mu.Lock()
	defer currentBuilder.mu.Unlock()

	if expected == "" {
		expected = "N/A"
	}

	currentBuilder.steps = append(currentBuilder.steps, testStep{
		Action:          action,
		Expected:        expected,
		RequirementLink: req,
	})

	// Collect requirements for up_links
	if req != "" {
		for _, r := range strings.Split(req, ",") {
			currentBuilder.upLinks[strings.TrimSpace(r)] = true
		}
	}
}

// DsManual is a convenience wrapper around DsStep that prepends '<b>Manual:</b> ' to the action.
//
// Equivalent to: DsStep("<b>Manual:</b> "+action, expected, req)
//
// Parameters:
//   - action: Description of the test action.
//   - expected: Expected result. Use "" for "N/A" (setup/cleanup steps).
//   - req: Requirement link(s), comma-separated (e.g., "SOFT-387").
func DsManual(action, expected, req string) {
	DsStep("<b>Manual:</b> "+action, expected, req)
}

// DsUser is a convenience wrapper around DsStep that prepends '<b><i>USER INTERVENTION</b></i> ' to the action.
//
// Equivalent to: DsStep("<b><i>USER INTERVENTION</b></i> "+action, expected, req)
//
// Parameters:
//   - action: Description of the test action.
//   - expected: Expected result. Use "" for "N/A" (setup/cleanup steps).
//   - req: Requirement link(s), comma-separated (e.g., "SOFT-387").
func DsUser(action, expected, req string) {
	DsStep("<b><i>USER INTERVENTION</b></i> "+action, expected, req)
}

// DsFooter finishes building and inserts the docstring into the source file.
//
// Parameters:
//   - labels: Variadic list of label strings (e.g., "Automated", "VM").
func DsFooter(labels ...string) {
	if !Ds || currentBuilder == nil {
		return
	}

	currentBuilder.mu.Lock()
	defer currentBuilder.mu.Unlock()

	currentBuilder.labels = labels

	yaml := generateYAML()
	insertDocstring(yaml)
	currentBuilder = nil
}

func generateYAML() string {
	b := currentBuilder
	var lines []string

	lines = append(lines, "/*")
	lines = append(lines, "---")

	// Title
	lines = append(lines, fmt.Sprintf(`title: "%s"`, b.title))

	// Folder
	if b.folder != "" {
		lines = append(lines, fmt.Sprintf("folder: %s", b.folder))
	}

	// Description (literal block scalar)
	if b.description != "" {
		lines = append(lines, "description: |")
		for _, line := range strings.Split(b.description, "\n") {
			lines = append(lines, fmt.Sprintf("  %s", line))
		}
	}

	// Assumptions (formatted as HTML list)
	if len(b.assumptions) > 0 {
		lines = append(lines, "assumptions: |")
		lines = append(lines, "  <ul>")
		for _, a := range b.assumptions {
			// Strip leading "* " if present and wrap in <li> tags
			text := a
			if strings.HasPrefix(a, "* ") {
				text = a[2:]
			}
			lines = append(lines, fmt.Sprintf("   <li>%s</li>", text))
		}
		lines = append(lines, "  </ul>")
	}

	// Steps
	lines = append(lines, "steps:")
	for _, step := range b.steps {
		lines = append(lines, fmt.Sprintf(`  - action: "%s"`, step.Action))
		if step.Expected != "N/A" {
			lines = append(lines, fmt.Sprintf(`    expected: "%s"`, step.Expected))
		}
		if step.RequirementLink != "" {
			lines = append(lines, fmt.Sprintf(`    RequirementLink: "%s"`, step.RequirementLink))
		}
	}

	// Labels
	if len(b.labels) > 0 {
		lines = append(lines, "labels:")
		for _, l := range b.labels {
			lines = append(lines, fmt.Sprintf("  - %s", l))
		}
	}

	// Up links (collected from steps)
	if len(b.upLinks) > 0 {
		var links []string
		for l := range b.upLinks {
			links = append(links, l)
		}
		sort.Strings(links)
		lines = append(lines, fmt.Sprintf(`up_links: "%s"`, strings.Join(links, ", ")))
	}

	lines = append(lines, "---")
	lines = append(lines, "*/")

	return strings.Join(lines, "\n")
}

func insertDocstring(yaml string) {
	b := currentBuilder

	content, err := os.ReadFile(b.sourceFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", b.sourceFile, err)
		return
	}

	// Pattern to find function definition and optional existing docstring
	// Matches: func FuncName(...) { followed by optional block comment
	funcPattern := regexp.MustCompile(
		fmt.Sprintf(`(func %s\([^)]*\)[^{]*\{)\s*\n(\s*/\*[\s\S]*?\*/\s*\n)?`,
			regexp.QuoteMeta(b.funcName)))

	// Indent the YAML with a tab (Go convention)
	indentedYAML := strings.ReplaceAll(yaml, "\n", "\n\t")
	replacement := fmt.Sprintf("$1\n\t%s\n", indentedYAML)

	newContent := funcPattern.ReplaceAllString(string(content), replacement)

	if newContent == string(content) {
		fmt.Printf("Could not find function '%s' in %s\n", b.funcName, b.sourceFile)
		return
	}

	err = os.WriteFile(b.sourceFile, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", b.sourceFile, err)
		return
	}

	fmt.Printf("Generated docstring for %s in %s\n", b.funcName, b.sourceFile)
}
