package execution

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// XTCStats holds statistics for a single executed test case.
type XTCStats struct {
	ItemRef                   string `json:"item_ref"`
	TestStatus                string `json:"test_status"` // "Done", "In Progress", "Not Executed"
	NumSteps                  int    `json:"num_steps"`
	NumPassed                 int    `json:"num_passed"`
	NumFailed                 int    `json:"num_failed"`
	NumPassWithIssue          int    `json:"num_pass_with_issue"`
	NumNotExecutedWithReq     int    `json:"num_not_executed_with_req"`
	NumNotExecutedWithoutReq  int    `json:"num_not_executed_without_req"`
}

// ExecutionStats holds aggregated execution statistics for a folder.
type ExecutionStats struct {
	XTCStats             map[string]*XTCStats  `json:"xtc_stats"`
	SOFTCoverage         map[string]map[string]string `json:"soft_coverage"`         // SOFT → {XTC: result}
	OverallSOFTCoverage  map[string]string            `json:"overall_soft_coverage"` // SOFT → worst result
	ExecutedSOFTTotals   map[string]int               `json:"executed_soft_totals"`
	TotalTestsExecuted   int `json:"total_tests_executed"`
	TotalTestsInProgress int `json:"total_tests_in_progress"`
	TotalTestsNotExecuted int `json:"total_tests_not_executed"`
	TotalSteps           int `json:"total_steps"`
	TotalPassed          int `json:"total_passed"`
	TotalFailed          int `json:"total_failed"`
	TotalPassWithIssue   int `json:"total_pass_with_issue"`
	TotalNotExecutedWithReq int `json:"total_not_executed_with_req"`
	TotalWithoutReq      int `json:"total_without_req"`
	TotalWithoutReqComplete int `json:"total_without_req_complete"`
}

// priorityOrder defines the priority for determining worst result.
// Higher index = worse result.
var priorityOrder = []string{"passed", "pass with issue", "Not Executed", "failed"}

func priorityIndex(result string) int {
	for i, r := range priorityOrder {
		if r == result {
			return i
		}
	}
	return -1
}

// ToDict converts ExecutionStats to a map for JSON serialization.
func (s *ExecutionStats) ToDict() map[string]interface{} {
	xtcIndividual := make(map[string]interface{})
	for ref, stats := range s.XTCStats {
		xtcIndividual[ref] = map[string]interface{}{
			"TestStatus":              stats.TestStatus,
			"NumSteps":                stats.NumSteps,
			"NumPassed":               stats.NumPassed,
			"NumFailed":               stats.NumFailed,
			"NumPassWithIssue":        stats.NumPassWithIssue,
			"NumNotExecutedWithReq":   stats.NumNotExecutedWithReq,
			"NumNotExecutedWithoutReq": stats.NumNotExecutedWithoutReq,
		}
	}

	// Sort overall SOFT coverage
	sortedSOFT := make(map[string]string)
	var keys []string
	for k := range s.OverallSOFTCoverage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedSOFT[k] = s.OverallSOFTCoverage[k]
	}

	return map[string]interface{}{
		"SOFT Coverage":           s.SOFTCoverage,
		"XTC Individual Stats":   xtcIndividual,
		"XTC Total Stats": map[string]interface{}{
			"TotalTestsExecuted":     s.TotalTestsExecuted,
			"TotalTestsInProgress":   s.TotalTestsInProgress,
			"TotalTestsNotExecuted":  s.TotalTestsNotExecuted,
			"TotalNumSteps":          s.TotalSteps,
			"TotalNumPassed":         s.TotalPassed,
			"TotalNumFailed":         s.TotalFailed,
			"TotalNumPassWithIssue":  s.TotalPassWithIssue,
			"TotalNumNotExecutedWithReq": s.TotalNotExecutedWithReq,
			"TotalNumWithoutReq":     s.TotalWithoutReq,
			"TotalNumWithoutReqComplete": s.TotalWithoutReqComplete,
			"TotalStepsLeft":         s.TotalNotExecutedWithReq + s.TotalWithoutReq - s.TotalWithoutReqComplete,
		},
		"Executed SOFT Total Stats": s.ExecutedSOFTTotals,
		"Overall Soft Coverage":     sortedSOFT,
	}
}

// ComputeStats calculates execution statistics for all XTCs in a folder.
func ComputeStats(svc *service.MatrixService, project string, folderRef string, fm *fieldmap.FieldMap) (*ExecutionStats, error) {
	folder, err := svc.Items.GetFolder(project, folderRef, false)
	if err != nil {
		return nil, fmt.Errorf("getting folder: %w", err)
	}

	// Collect all non-folder items recursively
	items := collectItems(folder.ItemList)

	stats := &ExecutionStats{
		XTCStats:            make(map[string]*XTCStats),
		SOFTCoverage:        make(map[string]map[string]string),
		OverallSOFTCoverage: make(map[string]string),
		ExecutedSOFTTotals: map[string]int{
			"passed":          0,
			"pass with issue": 0,
			"Not Executed":    0,
			"failed":          0,
		},
	}

	for _, item := range items {
		itemRef := item.ItemRef

		// Get full item details
		itemData, err := svc.Items.Get(project, itemRef, false)
		if err != nil {
			continue
		}

		steps := parseStepsFromItem(itemData, fm)
		testRunResult := getFieldValue(itemData, fm, "XTC", "Test Run Result")
		testComplete := testRunResult == "p"

		xtcStat := &XTCStats{
			ItemRef:    itemRef,
			TestStatus: "Not Executed",
		}

		for _, step := range steps {
			xtcStat.NumSteps++

			expected := step.Expected
			hasExpected := expected != "" && !strings.EqualFold(expected, "n/a")
			hasReq := step.RequirementLink != ""
			hasResult := step.Human != ""

			if hasExpected && hasReq {
				reqLinks := strings.Split(step.RequirementLink, ",")
				firstReq := true

				for _, reqLink := range reqLinks {
					reqLink = strings.TrimSpace(reqLink)
					if reqLink == "" {
						continue
					}

					if hasResult {
						result := step.Human
						updateSOFTCoverage(stats, reqLink, itemRef, result)

						if firstReq {
							switch result {
							case "passed":
								xtcStat.NumPassed++
							case "failed":
								xtcStat.NumFailed++
							default:
								xtcStat.NumPassWithIssue++
							}
						}
					} else {
						updateSOFTCoverage(stats, reqLink, itemRef, "Not Executed")
						xtcStat.NumNotExecutedWithReq++
					}
					firstReq = false
				}
			} else if !hasResult {
				xtcStat.NumNotExecutedWithoutReq++
				if testComplete {
					stats.TotalWithoutReqComplete++
				}
			}
		}

		// Determine test status
		if xtcStat.NumNotExecutedWithReq == 0 {
			xtcStat.TestStatus = "Done"
			stats.TotalTestsExecuted++
		} else if xtcStat.NumPassed == 0 && xtcStat.NumFailed == 0 && xtcStat.NumPassWithIssue == 0 {
			xtcStat.TestStatus = "Not Executed"
			stats.TotalTestsNotExecuted++
		} else {
			xtcStat.TestStatus = "In Progress"
			stats.TotalTestsInProgress++
		}

		stats.XTCStats[itemRef] = xtcStat
		stats.TotalSteps += xtcStat.NumSteps
		stats.TotalPassed += xtcStat.NumPassed
		stats.TotalFailed += xtcStat.NumFailed
		stats.TotalPassWithIssue += xtcStat.NumPassWithIssue
		stats.TotalNotExecutedWithReq += xtcStat.NumNotExecutedWithReq
		stats.TotalWithoutReq += xtcStat.NumNotExecutedWithoutReq
	}

	return stats, nil
}

// updateSOFTCoverage updates the SOFT coverage maps with a new result.
func updateSOFTCoverage(stats *ExecutionStats, reqLink, itemRef, result string) {
	if _, ok := stats.SOFTCoverage[reqLink]; !ok {
		stats.SOFTCoverage[reqLink] = map[string]string{itemRef: result}
		stats.OverallSOFTCoverage[reqLink] = result
		stats.ExecutedSOFTTotals[result]++
		return
	}

	// Update per-XTC coverage
	current, exists := stats.SOFTCoverage[reqLink][itemRef]
	if !exists || priorityIndex(current) < priorityIndex(result) {
		stats.SOFTCoverage[reqLink][itemRef] = result
	}

	// Update overall coverage
	overall := stats.OverallSOFTCoverage[reqLink]
	if priorityIndex(overall) < priorityIndex(result) {
		stats.ExecutedSOFTTotals[overall]--
		stats.OverallSOFTCoverage[reqLink] = result
		stats.ExecutedSOFTTotals[result]++
	}
}

// collectItems recursively collects all non-folder items from a folder tree.
func collectItems(items []api.TrimFolder) []api.TrimFolder {
	var result []api.TrimFolder
	for _, item := range items {
		if item.IsFolder != 0 || len(item.ItemList) > 0 {
			result = append(result, collectItems(item.ItemList)...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

// getFieldValue extracts a field value from an item by category and label.
func getFieldValue(item *api.TrimItem, fm *fieldmap.FieldMap, category, label string) string {
	if item.FieldValList == nil {
		return ""
	}
	fieldID, err := fm.Resolve(category, label)
	if err != nil {
		return ""
	}
	for _, fv := range item.FieldValList.FieldVal {
		if fv.ID == fieldID {
			return fv.Value
		}
	}
	return ""
}

// parseStepsFromTrimItem parses steps JSON from an item's field values.
func parseStepsFromTrimItem(item *api.TrimItem, stepsFieldID int) []TestStep {
	if item.FieldValList == nil {
		return nil
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
