package execution

import (
	"encoding/json"
	"testing"
)

func TestTestStepPreservesUnmodeledKeys(t *testing.T) {
	// "Ref" is the xtc_config column id Matrix reads for "Requirement/Tspec
	// Link"; the trailing key stands in for any other project-specific column.
	const raw = `{"action":"verify","expected":"ok","Ref":"SOFT-5",` +
		`"RequirementLink":"SOFT-5","projectColumn":"keep me"}`

	var step TestStep
	if err := json.Unmarshal([]byte(raw), &step); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if step.RequirementLink != "SOFT-5" {
		t.Fatalf("RequirementLink: got %q", step.RequirementLink)
	}

	step.Result, step.Human, step.Render = "p", "passed", "ok"
	step.Comment = "defaults populated"

	out, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}

	if got["Ref"] != "SOFT-5" {
		t.Errorf("Ref was dropped (%v) — the Requirement/Tspec Link column would be emptied", got["Ref"])
	}
	if got["projectColumn"] != "keep me" {
		t.Errorf("unmodeled column was dropped: %v", got["projectColumn"])
	}
	if got["action"] != "verify" || got["expected"] != "ok" {
		t.Errorf("step definition changed: %v", got)
	}
	if got["result"] != "p" || got["human"] != "passed" || got["comment"] != "defaults populated" {
		t.Errorf("result not applied: %v", got)
	}
}

// Matrix omits RequirementLink on an action step, so the round-trip must not
// invent it: an upload may add results, never keys.
func TestTestStepActionStepGainsNoKeys(t *testing.T) {
	const raw = `{"action":"navigate","expected":"N/A"}`

	var step TestStep
	if err := json.Unmarshal([]byte(raw), &step); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if string(out) != raw {
		t.Errorf("action step changed:\n before %s\n after  %s", raw, out)
	}
}

// The whole upload path: read an XTC's Steps field, clear it, apply results,
// write it back. The step definitions must come out byte-equivalent apart from
// the result keys.
func TestUploadRoundTripKeepsRequirementColumn(t *testing.T) {
	const stepsField = `[` +
		`{"action":"navigate","expected":"N/A"},` +
		`{"action":"verify defaults","expected":"populated","Ref":"SOFT-5","RequirementLink":"SOFT-5"},` +
		`{"action":"discard","expected":"N/A"}` +
		`]`

	var upstream []TestStep
	if err := json.Unmarshal([]byte(stepsField), &upstream); err != nil {
		t.Fatalf("unmarshal steps field: %v", err)
	}

	local := []ExecutionResultStep{
		{Actual: "navigate", Status: "PASS"},
		pass("SOFT-5", "defaults populated"),
		{Actual: "discard", Status: "PASS"},
	}

	m := MatchSteps("TC-38", "XTC-686", local, upstream)
	if m.OutOfSync {
		t.Fatalf("unexpected out of sync: %v", m.Issues)
	}

	out, err := json.Marshal(m.Steps)
	if err != nil {
		t.Fatalf("marshal steps field: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}

	if got[1]["Ref"] != "SOFT-5" {
		t.Errorf("Ref lost on the verification step: %v", got[1])
	}
	if got[1]["result"] != "p" || got[1]["comment"] != "defaults populated" {
		t.Errorf("result not applied to the verification step: %v", got[1])
	}
	// A step that never had a Ref must not acquire one.
	if _, ok := got[0]["Ref"]; ok {
		t.Errorf("action step gained a Ref: %v", got[0])
	}
	// Every step is recorded, including those with no requirement.
	if got[0]["result"] != "p" {
		t.Errorf("action step has no result: %v", got[0])
	}
}

func TestClearResultsKeepsUnmodeledKeys(t *testing.T) {
	const raw = `[{"action":"verify","expected":"ok","Ref":"SOFT-5","RequirementLink":"SOFT-5",` +
		`"result":"p","human":"passed","render":"ok","comment":"previous run"}]`

	var steps []TestStep
	if err := json.Unmarshal([]byte(raw), &steps); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	out, err := json.Marshal(ClearResults(steps))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}

	if got[0]["Ref"] != "SOFT-5" {
		t.Errorf("Ref lost while clearing results: %v", got[0])
	}
	for _, key := range []string{"result", "human", "render", "comment"} {
		if _, ok := got[0][key]; ok {
			t.Errorf("%q survived the clear: %v", key, got[0])
		}
	}
}

// encoding/json fills a typed field from a key of any letter case, so the
// unmodeled-key stash has to strip with the same rule. Otherwise the value is
// written back under both spellings.
func TestTestStepCaseVariantKeyIsNotDuplicated(t *testing.T) {
	const raw = `{"action":"a","requirementlink":"SOFT-5","projectColumn":"keep"}`

	var step TestStep
	if err := json.Unmarshal([]byte(raw), &step); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if step.RequirementLink != "SOFT-5" {
		t.Fatalf("typed field not filled: %q", step.RequirementLink)
	}
	if _, dup := step.extra["requirementlink"]; dup {
		t.Errorf("case-variant key left in extra: %v", step.extra)
	}

	out, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if got["RequirementLink"] != "SOFT-5" {
		t.Errorf("RequirementLink lost: %v", got)
	}
	if _, dup := got["requirementlink"]; dup {
		t.Errorf("value written back twice: %s", out)
	}
	if got["projectColumn"] != "keep" {
		t.Errorf("unmodeled column dropped: %v", got)
	}
}

// A step stored with RequirementLink explicitly "" must round-trip that key; a
// step that never had it must still gain none.
func TestTestStepKeepsExplicitEmptyRequirementLink(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantKey bool
	}{
		{"explicit empty", `{"action":"a","RequirementLink":""}`, true},
		{"explicit empty, other case", `{"action":"a","requirementlink":""}`, true},
		{"absent", `{"action":"a"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var step TestStep
			if err := json.Unmarshal([]byte(tt.raw), &step); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			// Survive the upload path, which copies steps via ClearResults.
			out, err := json.Marshal(ClearResults([]TestStep{step})[0])
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			_ = json.Unmarshal(out, &got)
			v, has := got["RequirementLink"]
			if has != tt.wantKey {
				t.Fatalf("RequirementLink present=%v, want %v: %s", has, tt.wantKey, out)
			}
			if has && v != "" {
				t.Errorf("RequirementLink = %v, want empty", v)
			}
			if _, dup := got["requirementlink"]; dup {
				t.Errorf("case variant written back: %s", out)
			}
		})
	}
}
