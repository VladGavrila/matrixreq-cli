package execution

import (
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
)

func TestPriorityIndex(t *testing.T) {
	tests := []struct {
		result string
		want   int
	}{
		{"passed", 0},
		{"pass with issue", 1},
		{"Not Executed", 2},
		{"failed", 3},
		{"unknown", -1},
	}

	for _, tt := range tests {
		got := priorityIndex(tt.result)
		if got != tt.want {
			t.Errorf("priorityIndex(%q) = %d, want %d", tt.result, got, tt.want)
		}
	}
}

func TestPriorityOrder(t *testing.T) {
	// Verify that higher index = worse result
	if priorityIndex("passed") >= priorityIndex("failed") {
		t.Error("passed should have lower priority than failed")
	}
	if priorityIndex("pass with issue") >= priorityIndex("Not Executed") {
		t.Error("pass with issue should have lower priority than Not Executed")
	}
	if priorityIndex("Not Executed") >= priorityIndex("failed") {
		t.Error("Not Executed should have lower priority than failed")
	}
}

func TestUpdateSOFTCoverageNewReq(t *testing.T) {
	stats := &ExecutionStats{
		SOFTCoverage:        make(map[string]map[string]string),
		OverallSOFTCoverage: make(map[string]string),
		ExecutedSOFTTotals: map[string]int{
			"passed":          0,
			"pass with issue": 0,
			"Not Executed":    0,
			"failed":          0,
		},
	}

	updateSOFTCoverage(stats, "SOFT-1", "XTC-10", "passed")

	if stats.OverallSOFTCoverage["SOFT-1"] != "passed" {
		t.Errorf("overall coverage: got %q, want %q", stats.OverallSOFTCoverage["SOFT-1"], "passed")
	}
	if stats.ExecutedSOFTTotals["passed"] != 1 {
		t.Errorf("passed total: got %d, want 1", stats.ExecutedSOFTTotals["passed"])
	}
	if stats.SOFTCoverage["SOFT-1"]["XTC-10"] != "passed" {
		t.Errorf("per-XTC coverage: got %q", stats.SOFTCoverage["SOFT-1"]["XTC-10"])
	}
}

func TestUpdateSOFTCoverageWorseResult(t *testing.T) {
	stats := &ExecutionStats{
		SOFTCoverage: map[string]map[string]string{
			"SOFT-1": {"XTC-10": "passed"},
		},
		OverallSOFTCoverage: map[string]string{
			"SOFT-1": "passed",
		},
		ExecutedSOFTTotals: map[string]int{
			"passed":          1,
			"pass with issue": 0,
			"Not Executed":    0,
			"failed":          0,
		},
	}

	// Add a worse result for the same requirement from a different XTC
	updateSOFTCoverage(stats, "SOFT-1", "XTC-20", "failed")

	if stats.OverallSOFTCoverage["SOFT-1"] != "failed" {
		t.Errorf("overall should be worst-case: got %q, want %q", stats.OverallSOFTCoverage["SOFT-1"], "failed")
	}
	if stats.ExecutedSOFTTotals["passed"] != 0 {
		t.Errorf("passed total should decrease: got %d", stats.ExecutedSOFTTotals["passed"])
	}
	if stats.ExecutedSOFTTotals["failed"] != 1 {
		t.Errorf("failed total should increase: got %d", stats.ExecutedSOFTTotals["failed"])
	}
}

func TestUpdateSOFTCoverageBetterResult(t *testing.T) {
	stats := &ExecutionStats{
		SOFTCoverage: map[string]map[string]string{
			"SOFT-1": {"XTC-10": "failed"},
		},
		OverallSOFTCoverage: map[string]string{
			"SOFT-1": "failed",
		},
		ExecutedSOFTTotals: map[string]int{
			"passed":          0,
			"pass with issue": 0,
			"Not Executed":    0,
			"failed":          1,
		},
	}

	// Add a better result - should NOT change overall (overall = worst-case)
	updateSOFTCoverage(stats, "SOFT-1", "XTC-20", "passed")

	if stats.OverallSOFTCoverage["SOFT-1"] != "failed" {
		t.Errorf("overall should remain worst-case: got %q, want %q", stats.OverallSOFTCoverage["SOFT-1"], "failed")
	}
	if stats.ExecutedSOFTTotals["failed"] != 1 {
		t.Errorf("failed total should remain 1: got %d", stats.ExecutedSOFTTotals["failed"])
	}
}

func TestCollectItems(t *testing.T) {
	items := []api.TrimFolder{
		{ItemRef: "XTC-1", IsFolder: 0},
		{
			ItemRef:  "F-XTC-1",
			IsFolder: 1,
			ItemList: []api.TrimFolder{
				{ItemRef: "XTC-2", IsFolder: 0},
				{ItemRef: "XTC-3", IsFolder: 0},
				{
					ItemRef:  "F-XTC-2",
					IsFolder: 1,
					ItemList: []api.TrimFolder{
						{ItemRef: "XTC-4", IsFolder: 0},
					},
				},
			},
		},
	}

	collected := collectItems(items)
	if len(collected) != 4 {
		t.Fatalf("expected 4 items, got %d", len(collected))
	}

	refs := make(map[string]bool)
	for _, item := range collected {
		refs[item.ItemRef] = true
	}
	for _, expected := range []string{"XTC-1", "XTC-2", "XTC-3", "XTC-4"} {
		if !refs[expected] {
			t.Errorf("missing item %s in collected items", expected)
		}
	}
}

func TestCollectItemsEmpty(t *testing.T) {
	collected := collectItems(nil)
	if len(collected) != 0 {
		t.Errorf("expected 0 items for nil input, got %d", len(collected))
	}

	collected = collectItems([]api.TrimFolder{})
	if len(collected) != 0 {
		t.Errorf("expected 0 items for empty input, got %d", len(collected))
	}
}

func TestParseStepsFromTrimItem(t *testing.T) {
	stepsJSON := `[{"action":"Do X","expected":"Y","RequirementLink":"SOFT-1"},{"action":"Do Z","expected":"W","RequirementLink":""}]`
	item := &api.TrimItem{
		FieldValList: &api.FieldValListType{
			FieldVal: []api.FieldValType{
				{ID: 100, Value: "other field"},
				{ID: 762, Value: stepsJSON},
			},
		},
	}

	steps := parseStepsFromTrimItem(item, 762)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}

	if steps[0].Action != "Do X" {
		t.Errorf("step 0 action: got %q", steps[0].Action)
	}
	if steps[0].RequirementLink != "SOFT-1" {
		t.Errorf("step 0 requirement: got %q", steps[0].RequirementLink)
	}
	if steps[1].RequirementLink != "" {
		t.Errorf("step 1 requirement should be empty, got %q", steps[1].RequirementLink)
	}
}

func TestParseStepsFromTrimItemNoFieldValList(t *testing.T) {
	item := &api.TrimItem{}
	steps := parseStepsFromTrimItem(item, 762)
	if steps != nil {
		t.Error("expected nil steps for item without field values")
	}
}

func TestParseStepsFromTrimItemInvalidJSON(t *testing.T) {
	item := &api.TrimItem{
		FieldValList: &api.FieldValListType{
			FieldVal: []api.FieldValType{
				{ID: 762, Value: "not json"},
			},
		},
	}

	steps := parseStepsFromTrimItem(item, 762)
	if steps != nil {
		t.Error("expected nil steps for invalid JSON")
	}
}

func TestParseStepsFromTrimItemEmptyValue(t *testing.T) {
	item := &api.TrimItem{
		FieldValList: &api.FieldValListType{
			FieldVal: []api.FieldValType{
				{ID: 762, Value: ""},
			},
		},
	}

	steps := parseStepsFromTrimItem(item, 762)
	if steps != nil {
		t.Error("expected nil steps for empty value")
	}
}

func TestExecutionStatsToDictStructure(t *testing.T) {
	stats := &ExecutionStats{
		XTCStats: map[string]*XTCStats{
			"XTC-1": {
				ItemRef:    "XTC-1",
				TestStatus: "Done",
				NumSteps:   3,
				NumPassed:  2,
				NumFailed:  1,
			},
		},
		SOFTCoverage: map[string]map[string]string{
			"SOFT-1": {"XTC-1": "passed"},
		},
		OverallSOFTCoverage: map[string]string{
			"SOFT-1": "passed",
		},
		ExecutedSOFTTotals: map[string]int{
			"passed":          1,
			"pass with issue": 0,
			"Not Executed":    0,
			"failed":          0,
		},
		TotalTestsExecuted:    1,
		TotalTestsInProgress:  0,
		TotalTestsNotExecuted: 0,
		TotalSteps:            3,
		TotalPassed:           2,
		TotalFailed:           1,
	}

	dict := stats.ToDict()

	// Verify top-level keys
	if _, ok := dict["SOFT Coverage"]; !ok {
		t.Error("missing 'SOFT Coverage' key")
	}
	if _, ok := dict["XTC Individual Stats"]; !ok {
		t.Error("missing 'XTC Individual Stats' key")
	}
	if _, ok := dict["XTC Total Stats"]; !ok {
		t.Error("missing 'XTC Total Stats' key")
	}
	if _, ok := dict["Executed SOFT Total Stats"]; !ok {
		t.Error("missing 'Executed SOFT Total Stats' key")
	}
	if _, ok := dict["Overall Soft Coverage"]; !ok {
		t.Error("missing 'Overall Soft Coverage' key")
	}

	// Verify XTC Total Stats
	totalStats := dict["XTC Total Stats"].(map[string]interface{})
	if v, ok := totalStats["TotalTestsExecuted"]; !ok || v.(int) != 1 {
		t.Errorf("TotalTestsExecuted: got %v, want 1", v)
	}
	if v, ok := totalStats["TotalNumPassed"]; !ok || v.(int) != 2 {
		t.Errorf("TotalNumPassed: got %v, want 2", v)
	}
	if v, ok := totalStats["TotalNumFailed"]; !ok || v.(int) != 1 {
		t.Errorf("TotalNumFailed: got %v, want 1", v)
	}

	// Verify TotalStepsLeft calculation
	if v, ok := totalStats["TotalStepsLeft"]; !ok {
		t.Error("missing TotalStepsLeft")
	} else {
		// TotalStepsLeft = TotalNotExecutedWithReq + TotalWithoutReq - TotalWithoutReqComplete
		// 0 + 0 - 0 = 0
		if v.(int) != 0 {
			t.Errorf("TotalStepsLeft: got %v, want 0", v)
		}
	}

	// Verify XTC Individual Stats
	xtcStats := dict["XTC Individual Stats"].(map[string]interface{})
	xtc1 := xtcStats["XTC-1"].(map[string]interface{})
	if xtc1["TestStatus"] != "Done" {
		t.Errorf("XTC-1 TestStatus: got %v, want Done", xtc1["TestStatus"])
	}
	if xtc1["NumSteps"] != 3 {
		t.Errorf("XTC-1 NumSteps: got %v, want 3", xtc1["NumSteps"])
	}
}
