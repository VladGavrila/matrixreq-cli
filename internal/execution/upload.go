package execution

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// BuildTCToXTCMapping builds a mapping from TC refs to XTC items.
// XTC titles are expected to contain the TC ref in parentheses, e.g., "Title (TC-1377)".
func BuildTCToXTCMapping(folderItems []api.TrimFolder) map[string]api.TrimFolder {
	mapping := make(map[string]api.TrimFolder)
	for _, item := range folderItems {
		if len(item.ItemList) > 0 {
			// Recurse into subfolders
			sub := BuildTCToXTCMapping(item.ItemList)
			for k, v := range sub {
				mapping[k] = v
			}
		} else if item.IsFolder == 0 {
			// Extract TC ref from title: "Title (TC-1377)" → "TC-1377"
			title := item.Title
			start := strings.LastIndex(title, "(")
			end := strings.LastIndex(title, ")")
			if start != -1 && end != -1 && end > start {
				tcRef := title[start+1 : end]
				mapping[tcRef] = item
			}
		}
	}
	return mapping
}

// UploadResults uploads execution results to XTCs in a folder.
func UploadResults(svc *service.MatrixService, project string, folderRef string, results *ExecutionResults, fm *fieldmap.FieldMap) (*UploadResult, error) {
	// Get folder items to build TC→XTC mapping
	folder, err := svc.Items.GetFolder(project, folderRef, false)
	if err != nil {
		return nil, fmt.Errorf("getting folder: %w", err)
	}

	tcToXTC := BuildTCToXTCMapping(folder.ItemList)
	if len(tcToXTC) == 0 {
		return nil, fmt.Errorf("no XTCs found in folder %s", folderRef)
	}

	uploadResult := &UploadResult{
		Successes: make(map[string]bool),
	}

	// Load XTC data for each test result
	type xtcData struct {
		ref   string
		steps []TestStep
		item  *api.TrimItem
	}
	updatedXTCs := make(map[string]*xtcData)

	for _, testResult := range results.Results {
		tcName := testResult.TestName
		xtcFolder, ok := tcToXTC[tcName]
		if !ok {
			uploadResult.Issues = append(uploadResult.Issues, fmt.Sprintf("test %q not found in folder", tcName))
			continue
		}

		xtcRef := xtcFolder.ItemRef

		// Get full XTC item if not already loaded
		if _, loaded := updatedXTCs[tcName]; !loaded {
			item, err := svc.Items.Get(project, xtcRef, false)
			if err != nil {
				uploadResult.Issues = append(uploadResult.Issues, fmt.Sprintf("failed to get %s: %v", xtcRef, err))
				continue
			}

			// Parse steps from field values
			steps := parseStepsFromItem(item, fm)

			// Clear previous results
			for i := range steps {
				steps[i].Result = ""
				steps[i].Human = ""
				steps[i].Render = ""
				steps[i].Comment = ""
			}

			updatedXTCs[tcName] = &xtcData{
				ref:   xtcRef,
				steps: steps,
				item:  item,
			}
		}

		xtc := updatedXTCs[tcName]

		// Match step results to XTC steps by requirement
		for _, stepResult := range testResult.Steps {
			if stepResult.Requirement == "" {
				continue
			}

			stepFound := false
			for i := range xtc.steps {
				if xtc.steps[i].RequirementLink == stepResult.Requirement && xtc.steps[i].Human == "" {
					if stepResult.Status == "PASS" {
						xtc.steps[i].Result = "p"
						xtc.steps[i].Human = "passed"
						xtc.steps[i].Render = "ok"
					} else {
						xtc.steps[i].Result = "f"
						xtc.steps[i].Human = "failed"
						xtc.steps[i].Render = "error"
					}
					xtc.steps[i].Comment = stepResult.Actual
					stepFound = true
					break
				}
			}

			if !stepFound {
				uploadResult.Issues = append(uploadResult.Issues,
					fmt.Sprintf("%s: requirement %s not found in XTC steps", tcName, stepResult.Requirement))
			}
		}
	}

	// Upload results for each XTC
	for tcName, xtc := range updatedXTCs {
		// Check for incomplete executions
		hasIssue := false
		runResult := "p"
		for _, step := range xtc.steps {
			if step.RequirementLink != "" && step.Result == "" {
				hasIssue = true
				uploadResult.Issues = append(uploadResult.Issues,
					fmt.Sprintf("%s: step with requirement %s was not executed", tcName, step.RequirementLink))
				break
			}
			if step.Result == "f" {
				runResult = "f"
			}
		}
		if hasIssue {
			continue
		}

		// Upload
		err := updateXTCResults(svc, project, fm, xtc.ref, xtc.steps, xtc.item, runResult, results)
		uploadResult.Successes[xtc.ref] = err == nil
		if err != nil {
			uploadResult.Issues = append(uploadResult.Issues,
				fmt.Sprintf("failed to update %s: %v", xtc.ref, err))
		}
	}

	return uploadResult, nil
}

// parseStepsFromItem extracts test steps from an item's field values.
func parseStepsFromItem(item *api.TrimItem, fm *fieldmap.FieldMap) []TestStep {
	if item.FieldValList == nil {
		return nil
	}

	// Resolve the Steps field ID for XTC category
	stepsFieldID, err := fm.Resolve("XTC", "Test Case Steps")
	if err != nil {
		// Try alternative field name
		stepsFieldID, err = fm.Resolve("XTC", "Steps")
		if err != nil {
			return nil
		}
	}

	for _, fv := range item.FieldValList.FieldVal {
		if fv.ID == stepsFieldID && fv.Value != "" {
			var steps []TestStep
			if err := json.Unmarshal([]byte(fv.Value), &steps); err != nil {
				return nil
			}
			return steps
		}
	}
	return nil
}

// updateXTCResults updates an XTC with execution results via the API.
func updateXTCResults(svc *service.MatrixService, project string, fm *fieldmap.FieldMap, xtcRef string, steps []TestStep, item *api.TrimItem, runResult string, results *ExecutionResults) error {
	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return fmt.Errorf("marshaling steps: %w", err)
	}

	// Convert date format from YYYY-MM-DD to YYYY/MM/DD for Matrix
	testDate := strings.ReplaceAll(results.ExecutionDate, "-", "/")

	// Resolve XTC field IDs
	testerID, err := fm.Resolve("XTC", "Tester")
	if err != nil {
		return fmt.Errorf("resolving Tester field: %w", err)
	}
	dateID, err := fm.Resolve("XTC", "Test Date")
	if err != nil {
		return fmt.Errorf("resolving Test Date field: %w", err)
	}
	runResultID, err := fm.Resolve("XTC", "Test Run Result")
	if err != nil {
		return fmt.Errorf("resolving Test Run Result field: %w", err)
	}
	stepsID, err := fm.Resolve("XTC", "Test Case Steps")
	if err != nil {
		// Try alternative
		stepsID, err = fm.Resolve("XTC", "Steps")
		if err != nil {
			return fmt.Errorf("resolving Steps field: %w", err)
		}
	}

	fields := []api.FieldValSetType{
		{ID: testerID, Value: results.Tester},
		{ID: dateID, Value: testDate},
		{ID: runResultID, Value: runResult},
		{ID: stepsID, Value: string(stepsJSON)},
	}

	// Optionally set version field
	if results.SUTVersion != "" {
		versionID, err := fm.Resolve("XTC", "Version")
		if err == nil {
			fields = append(fields, api.FieldValSetType{ID: versionID, Value: results.SUTVersion})
		}
	}

	updateReq := &api.UpdateItemRequest{
		Title:     item.Title,
		Reason:    "synced by mxreq",
		Fields:    fields,
		OnlyThose: true,
	}

	_, err = svc.Items.Update(project, xtcRef, updateReq)
	return err
}
