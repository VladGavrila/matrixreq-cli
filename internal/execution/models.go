package execution

import (
	"encoding/json"
	"reflect"
	"strings"
)

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
//
// Matrix stores a step as a free-form JSON object whose keys depend on the
// project's xtc_config, so this struct is only a partial view of one. Keys it
// does not model are kept verbatim in extra and written back untouched by
// MarshalJSON, because an upload rewrites the whole Steps field: anything
// dropped here is deleted from the item. That is how the "Requirement/Tspec
// Link" column (the "Ref" key) used to disappear from every XTC an
// upload touched, even though the upload never means to change step
// definitions at all.
type TestStep struct {
	Action   string `json:"action"`
	Expected string `json:"expected"`
	// omitempty because Matrix omits the key entirely on steps that verify no
	// requirement. Writing it back as "" would add a key the step never had,
	// which is still an upload editing step definitions.
	RequirementLink string `json:"RequirementLink,omitempty"`
	Result          string `json:"result,omitempty"`  // "p", "f"
	Human           string `json:"human,omitempty"`   // "passed", "failed"
	Render          string `json:"render,omitempty"`  // "ok", "error"
	Comment         string `json:"comment,omitempty"` // Execution comment

	// extra holds every key above that this struct does not model, exactly as
	// Matrix sent it.
	extra map[string]json.RawMessage

	// emptyLink records that Matrix sent RequirementLink explicitly as "".
	// omitempty would otherwise drop the key on the way back, so MarshalJSON
	// restores it. A step that never had the key still gains none.
	emptyLink bool
}

// modeledStepKeys are the JSON keys TestStep itself round-trips. Derived from
// the struct tags so adding a field cannot leave a key in both places.
var modeledStepKeys = func() map[string]struct{} {
	t := reflect.TypeOf(TestStep{})
	keys := make(map[string]struct{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		keys[strings.Split(tag, ",")[0]] = struct{}{}
	}
	return keys
}()

// stepAlias drops the custom marshaler so the modeled fields can be handled by
// encoding/json without recursing.
type stepAlias TestStep

func (s *TestStep) UnmarshalJSON(data []byte) error {
	var modeled stepAlias
	if err := json.Unmarshal(data, &modeled); err != nil {
		return err
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// encoding/json fills a typed field from a key of any letter case, but the
	// raw map is case-sensitive. Strip with the same rule the decoder used, or
	// a key such as "requirementlink" would sit in extra as well as in the typed
	// field and be written back twice.
	modeled.emptyLink = false
	for key := range modeledStepKeys {
		consumed := consumedKeys(all, key)
		if key == "RequirementLink" && len(consumed) > 0 && modeled.RequirementLink == "" {
			modeled.emptyLink = true
		}
		for _, k := range consumed {
			delete(all, k)
		}
	}
	if len(all) > 0 {
		modeled.extra = all
	}
	*s = TestStep(modeled)
	return nil
}

// consumedKeys returns every key of raw that encoding/json matched to the typed
// field tagged key: the exact key and any key differing only by letter case. If
// several spellings are present they all decode into the one field, so all of
// them are consumed.
func consumedKeys(raw map[string]json.RawMessage, key string) []string {
	var out []string
	for k := range raw {
		if strings.EqualFold(k, key) {
			out = append(out, k)
		}
	}
	return out
}

func (s TestStep) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(stepAlias(s))
	if err != nil {
		return nil, err
	}
	if len(s.extra) == 0 && !(s.emptyLink && s.RequirementLink == "") {
		return data, nil
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal(data, &merged); err != nil {
		return nil, err
	}
	for key, value := range s.extra {
		if _, modeled := merged[key]; !modeled {
			merged[key] = value
		}
	}
	if s.emptyLink && s.RequirementLink == "" {
		merged["RequirementLink"] = json.RawMessage(`""`)
	}
	return json.Marshal(merged)
}

// UploadResult holds the outcome of an execution results upload.
//
// Every test in the results file ends up in exactly one place: Successes
// (written), Skipped (refused as out of sync), Failed (could not be read or
// written), or Missing (no XTC in the folder). --strict fails on all but the
// first.
type UploadResult struct {
	Successes map[string]bool   // XTC ref → success
	Skipped   map[string]string // XTC ref → why nothing was written to it
	Failed    map[string]string // XTC ref → error reading or updating it
	Missing   []string          // test names with no XTC in the folder
	Issues    []string          // operator-facing messages for all of the above
}

// Unrecorded counts the results that did not make it into Matrix.
func (u *UploadResult) Unrecorded() int {
	return len(u.Skipped) + len(u.Failed) + len(u.Missing)
}

// StepMatch is the outcome of matching one test's step results against the
// steps an XTC holds.
type StepMatch struct {
	// Steps is the XTC's step list with a result on every step. It is nil when
	// OutOfSync is set, so a caller that forgets to check fails loudly.
	Steps []TestStep
	// RunResult is the XTC's overall result, "p" or "f". Empty when OutOfSync.
	RunResult string
	// OutOfSync reports that the local test and the XTC no longer describe the
	// same test case, so nothing may be written to the XTC.
	OutOfSync bool
	// Reason is a short summary of the first drift found, for the skip report.
	Reason string
	// Issues are the operator-facing messages describing every drift found.
	Issues []string
}
