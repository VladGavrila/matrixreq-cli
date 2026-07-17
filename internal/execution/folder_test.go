package execution

import (
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
)

func TestCategoryFromRef(t *testing.T) {
	cases := map[string]string{
		"F-XTC-123": "XTC",
		"XTC-308":   "XTC",
		"F-TC-2":    "TC",
		"TC-100":    "TC",
		"F-SOFT-4":  "SOFT",
		"":          "",
		"XTC":       "",
	}
	for ref, want := range cases {
		if got := categoryFromRef(ref); got != want {
			t.Errorf("categoryFromRef(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestFindFolderSubtree(t *testing.T) {
	tree := []api.TrimFolder{
		{
			ItemRef:  "F-XTC-1",
			Title:    "Testing",
			IsFolder: 1,
			ItemList: []api.TrimFolder{
				{
					ItemRef:  "F-XTC-119",
					Title:    "CytoCanvas Studio v1.2.0-customerRelease",
					IsFolder: 1,
					ItemList: []api.TrimFolder{
						{
							ItemRef:  "F-XTC-123",
							Title:    "CytoCanvasStudio - Automation",
							IsFolder: 1,
							ItemList: []api.TrimFolder{
								{ItemRef: "XTC-308", Title: "Flow Cell View (TC-2)", IsFolder: 0},
								{ItemRef: "XTC-309", Title: "Login and Logout (TC-3)", IsFolder: 0},
							},
						},
					},
				},
			},
		},
	}

	sub := findFolderSubtree(tree, "F-XTC-123")
	if sub == nil {
		t.Fatal("expected to find F-XTC-123 subtree, got nil")
	}
	if len(sub) != 2 {
		t.Fatalf("expected 2 children under F-XTC-123, got %d", len(sub))
	}
	if sub[0].ItemRef != "XTC-308" || sub[1].ItemRef != "XTC-309" {
		t.Errorf("unexpected children: %s, %s", sub[0].ItemRef, sub[1].ItemRef)
	}

	if got := findFolderSubtree(tree, "F-XTC-999"); got != nil {
		t.Errorf("expected nil for missing folder, got %v", got)
	}
}
