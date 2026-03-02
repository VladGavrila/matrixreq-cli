package execution

import (
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
)

func TestBuildTCToXTCMapping(t *testing.T) {
	items := []api.TrimFolder{
		{ItemRef: "XTC-1", Title: "Test Login (TC-100)", IsFolder: 0},
		{ItemRef: "XTC-2", Title: "Test Logout (TC-200)", IsFolder: 0},
		{ItemRef: "XTC-3", Title: "Test Registration (TC-300)", IsFolder: 0},
	}

	mapping := BuildTCToXTCMapping(items)

	if len(mapping) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mapping))
	}

	if m, ok := mapping["TC-100"]; !ok {
		t.Error("TC-100 not found in mapping")
	} else if m.ItemRef != "XTC-1" {
		t.Errorf("TC-100 should map to XTC-1, got %s", m.ItemRef)
	}

	if m, ok := mapping["TC-200"]; !ok {
		t.Error("TC-200 not found in mapping")
	} else if m.ItemRef != "XTC-2" {
		t.Errorf("TC-200 should map to XTC-2, got %s", m.ItemRef)
	}

	if m, ok := mapping["TC-300"]; !ok {
		t.Error("TC-300 not found in mapping")
	} else if m.ItemRef != "XTC-3" {
		t.Errorf("TC-300 should map to XTC-3, got %s", m.ItemRef)
	}
}

func TestBuildTCToXTCMappingRecursive(t *testing.T) {
	items := []api.TrimFolder{
		{
			ItemRef:  "F-XTC-1",
			Title:    "Subfolder",
			IsFolder: 1,
			ItemList: []api.TrimFolder{
				{ItemRef: "XTC-10", Title: "Nested Test (TC-500)", IsFolder: 0},
				{ItemRef: "XTC-11", Title: "Another Nested (TC-600)", IsFolder: 0},
			},
		},
		{ItemRef: "XTC-1", Title: "Top Level (TC-100)", IsFolder: 0},
	}

	mapping := BuildTCToXTCMapping(items)

	if len(mapping) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mapping))
	}

	if _, ok := mapping["TC-500"]; !ok {
		t.Error("TC-500 not found in recursive mapping")
	}
	if _, ok := mapping["TC-600"]; !ok {
		t.Error("TC-600 not found in recursive mapping")
	}
	if _, ok := mapping["TC-100"]; !ok {
		t.Error("TC-100 not found in mapping")
	}
}

func TestBuildTCToXTCMappingNoParentheses(t *testing.T) {
	items := []api.TrimFolder{
		{ItemRef: "XTC-1", Title: "No TC Ref Here", IsFolder: 0},
		{ItemRef: "XTC-2", Title: "Also No Ref", IsFolder: 0},
	}

	mapping := BuildTCToXTCMapping(items)

	if len(mapping) != 0 {
		t.Errorf("expected 0 mappings for titles without parentheses, got %d", len(mapping))
	}
}

func TestBuildTCToXTCMappingEmpty(t *testing.T) {
	mapping := BuildTCToXTCMapping(nil)
	if len(mapping) != 0 {
		t.Errorf("expected 0 mappings for nil input, got %d", len(mapping))
	}

	mapping = BuildTCToXTCMapping([]api.TrimFolder{})
	if len(mapping) != 0 {
		t.Errorf("expected 0 mappings for empty input, got %d", len(mapping))
	}
}

func TestBuildTCToXTCMappingSkipsFolders(t *testing.T) {
	items := []api.TrimFolder{
		{ItemRef: "F-XTC-1", Title: "Folder (TC-999)", IsFolder: 1, ItemList: []api.TrimFolder{}},
	}

	mapping := BuildTCToXTCMapping(items)

	// Folder with no children - should produce no mapping
	if len(mapping) != 0 {
		t.Errorf("expected 0 mappings for folder-only items, got %d", len(mapping))
	}
}

