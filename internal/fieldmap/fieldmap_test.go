package fieldmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
)

func TestBuildFieldMap(t *testing.T) {
	info := &api.ProjectInfo{
		CategoryList: api.CategoryExtendedListWrapper{
			CategoryExtended: []api.CategoryExtendedType{
				{
					Category: api.CategoryType{ShortLabel: "TC"},
					FieldList: api.FieldListType{
						Field: []api.FieldType{
							{ID: 100, Label: "Description"},
							{ID: 101, Label: "Steps"},
							{ID: 102, Label: "Assumptions"},
						},
					},
				},
				{
					Category: api.CategoryType{ShortLabel: "REQ"},
					FieldList: api.FieldListType{
						Field: []api.FieldType{
							{ID: 200, Label: "Description"},
							{ID: 201, Label: "Rationale"},
						},
					},
				},
				{
					Category: api.CategoryType{ShortLabel: "XTC"},
					FieldList: api.FieldListType{
						Field: []api.FieldType{
							{ID: 300, Label: "Test Case Steps"},
							{ID: 301, Label: "Tester"},
							{ID: 302, Label: "Test Date"},
							{ID: 303, Label: "Test Run Result"},
						},
					},
				},
			},
		},
	}

	fields := buildFieldMap(info)

	tests := []struct {
		key    string
		wantID int
	}{
		{"TC.Description", 100},
		{"TC.Steps", 101},
		{"TC.Assumptions", 102},
		{"REQ.Description", 200},
		{"REQ.Rationale", 201},
		{"XTC.Test Case Steps", 300},
		{"XTC.Tester", 301},
		{"XTC.Test Date", 302},
		{"XTC.Test Run Result", 303},
	}

	for _, tt := range tests {
		id, ok := fields[tt.key]
		if !ok {
			t.Errorf("field %q not found in map", tt.key)
			continue
		}
		if id != tt.wantID {
			t.Errorf("field %q: got ID %d, want %d", tt.key, id, tt.wantID)
		}
	}

	if len(fields) != 9 {
		t.Errorf("expected 9 fields, got %d", len(fields))
	}
}

func TestBuildFieldMapSkipsEmptyLabels(t *testing.T) {
	info := &api.ProjectInfo{
		CategoryList: api.CategoryExtendedListWrapper{
			CategoryExtended: []api.CategoryExtendedType{
				{
					Category: api.CategoryType{ShortLabel: "TC"},
					FieldList: api.FieldListType{
						Field: []api.FieldType{
							{ID: 100, Label: "Description"},
							{ID: 101, Label: ""},
						},
					},
				},
				{
					Category: api.CategoryType{ShortLabel: ""},
					FieldList: api.FieldListType{
						Field: []api.FieldType{
							{ID: 200, Label: "Ignored"},
						},
					},
				},
			},
		},
	}

	fields := buildFieldMap(info)
	if len(fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(fields))
	}
	if fields["TC.Description"] != 100 {
		t.Error("expected TC.Description=100")
	}
}

func TestFieldMapResolve(t *testing.T) {
	fm := &FieldMap{fields: map[string]int{
		"TC.Description": 100,
		"TC.Steps":       101,
		"REQ.Description": 200,
	}}

	id, err := fm.Resolve("TC", "Description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 100 {
		t.Errorf("got %d, want 100", id)
	}

	id, err = fm.Resolve("REQ", "Description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 200 {
		t.Errorf("got %d, want 200", id)
	}

	_, err = fm.Resolve("TC", "NonExistent")
	if err == nil {
		t.Error("expected error for non-existent field")
	}
}

func TestFieldMapFieldsForCategory(t *testing.T) {
	fm := &FieldMap{fields: map[string]int{
		"TC.Description":  100,
		"TC.Steps":        101,
		"REQ.Description": 200,
	}}

	tcFields := fm.FieldsForCategory("TC")
	if len(tcFields) != 2 {
		t.Errorf("expected 2 TC fields, got %d", len(tcFields))
	}
	if tcFields["Description"] != 100 {
		t.Error("expected Description=100")
	}
	if tcFields["Steps"] != 101 {
		t.Error("expected Steps=101")
	}

	reqFields := fm.FieldsForCategory("REQ")
	if len(reqFields) != 1 {
		t.Errorf("expected 1 REQ field, got %d", len(reqFields))
	}

	empty := fm.FieldsForCategory("NONEXISTENT")
	if len(empty) != 0 {
		t.Errorf("expected 0 fields for nonexistent category, got %d", len(empty))
	}
}

func TestFieldMapCategories(t *testing.T) {
	fm := &FieldMap{fields: map[string]int{
		"TC.Description":  100,
		"TC.Steps":        101,
		"REQ.Description": 200,
		"XTC.Tester":      301,
	}}

	cats := fm.Categories()
	if len(cats) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cats))
	}
	// Should be sorted
	expected := []string{"REQ", "TC", "XTC"}
	for i, cat := range cats {
		if cat != expected[i] {
			t.Errorf("category %d: got %q, want %q", i, cat, expected[i])
		}
	}
}

func TestFieldMapEntries(t *testing.T) {
	fm := &FieldMap{fields: map[string]int{
		"TC.Steps":        101,
		"TC.Description":  100,
		"REQ.Description": 200,
	}}

	entries := fm.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Should be sorted by category, then by label
	if entries[0].Category != "REQ" || entries[0].Label != "Description" {
		t.Errorf("entry 0: got %s.%s, want REQ.Description", entries[0].Category, entries[0].Label)
	}
	if entries[1].Category != "TC" || entries[1].Label != "Description" {
		t.Errorf("entry 1: got %s.%s, want TC.Description", entries[1].Category, entries[1].Label)
	}
	if entries[2].Category != "TC" || entries[2].Label != "Steps" {
		t.Errorf("entry 2: got %s.%s, want TC.Steps", entries[2].Category, entries[2].Label)
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	// Use a temp directory for the cache
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() {
		// Restore (t.Setenv handles cleanup automatically)
		_ = origHome
	}()

	// Create the config directory
	cacheDir := filepath.Join(tmpDir, ".config", "mxreq")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}

	// Save a cache
	cache := map[string]map[string]int{
		"PROJECT_A": {"TC.Steps": 100, "REQ.Description": 200},
		"PROJECT_B": {"TC.Steps": 300},
	}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	// Verify the file exists
	path, _ := cachePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("cache file does not exist after save")
	}

	// Load it back
	loaded, err := loadCache()
	if err != nil {
		t.Fatalf("loading cache: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 projects in cache, got %d", len(loaded))
	}
	if loaded["PROJECT_A"]["TC.Steps"] != 100 {
		t.Error("PROJECT_A.TC.Steps should be 100")
	}
	if loaded["PROJECT_B"]["TC.Steps"] != 300 {
		t.Error("PROJECT_B.TC.Steps should be 300")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Set up cache with two projects
	cacheDir := filepath.Join(tmpDir, ".config", "mxreq")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}

	cache := map[string]map[string]int{
		"PROJECT_A": {"TC.Steps": 100},
		"PROJECT_B": {"TC.Steps": 200},
	}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	// Clear one project
	if err := Clear("PROJECT_A"); err != nil {
		t.Fatalf("clearing project: %v", err)
	}

	loaded, err := loadCache()
	if err != nil {
		t.Fatalf("loading cache after clear: %v", err)
	}
	if _, ok := loaded["PROJECT_A"]; ok {
		t.Error("PROJECT_A should be removed from cache")
	}
	if _, ok := loaded["PROJECT_B"]; !ok {
		t.Error("PROJECT_B should still be in cache")
	}
}

func TestCacheClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cacheDir := filepath.Join(tmpDir, ".config", "mxreq")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}

	cache := map[string]map[string]int{
		"PROJECT_A": {"TC.Steps": 100},
	}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	if err := ClearAll(); err != nil {
		t.Fatalf("clearing all: %v", err)
	}

	path, _ := cachePath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("cache file should not exist after ClearAll")
	}
}

func TestCacheClearNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Should not error when cache doesn't exist
	if err := Clear("NONEXISTENT"); err != nil {
		t.Errorf("unexpected error clearing non-existent cache: %v", err)
	}

	if err := ClearAll(); err != nil {
		t.Errorf("unexpected error clearing all on non-existent cache: %v", err)
	}
}

func TestCacheFileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cacheDir := filepath.Join(tmpDir, ".config", "mxreq")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("creating cache dir: %v", err)
	}

	cache := map[string]map[string]int{
		"MY_PROJECT": {"TC.Steps": 762, "TC.Description": 761},
	}
	if err := saveCache(cache); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	// Read raw file and verify it's valid JSON
	path, _ := cachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading cache file: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("cache file is not valid JSON: %v", err)
	}

	proj, ok := raw["MY_PROJECT"]
	if !ok {
		t.Fatal("MY_PROJECT not found in cache file")
	}

	projMap := proj.(map[string]interface{})
	if v := projMap["TC.Steps"]; v.(float64) != 762 {
		t.Errorf("TC.Steps: got %v, want 762", v)
	}
}
