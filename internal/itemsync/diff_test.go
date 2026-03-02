package itemsync

import (
	"testing"
)

func TestFieldsEqualPlainStrings(t *testing.T) {
	tests := []struct {
		name  string
		local string
		srv   string
		want  bool
	}{
		{"identical", "hello", "hello", true},
		{"whitespace", "hello ", " hello", true},
		{"different", "hello", "world", false},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fieldsEqual(tt.local, tt.srv); got != tt.want {
				t.Errorf("fieldsEqual(%q, %q) = %v, want %v", tt.local, tt.srv, got, tt.want)
			}
		})
	}
}

func TestFieldsEqualJSON(t *testing.T) {
	// Same JSON with different formatting
	local := `[{"action":"Do X","expected":"Y","RequirementLink":"SOFT-1"}]`
	server := `[{"action":"Do X","RequirementLink":"SOFT-1","expected":"Y"}]`

	if !fieldsEqual(local, server) {
		t.Error("identical JSON with different key order should be equal")
	}

	// Different JSON
	local = `[{"action":"Do X"}]`
	server = `[{"action":"Do Y"}]`
	if fieldsEqual(local, server) {
		t.Error("different JSON should not be equal")
	}
}

func TestToSet(t *testing.T) {
	set := toSet([]string{"a", "b", "c"})
	if len(set) != 3 {
		t.Errorf("expected 3 items, got %d", len(set))
	}
	if !set["a"] || !set["b"] || !set["c"] {
		t.Error("set should contain a, b, c")
	}

	empty := toSet(nil)
	if len(empty) != 0 {
		t.Errorf("expected 0 items for nil, got %d", len(empty))
	}
}

func TestCompareSets(t *testing.T) {
	tests := []struct {
		name        string
		local       map[string]bool
		server      map[string]bool
		wantAdded   int
		wantRemoved int
	}{
		{
			name:        "identical",
			local:       map[string]bool{"a": true, "b": true},
			server:      map[string]bool{"a": true, "b": true},
			wantAdded:   0,
			wantRemoved: 0,
		},
		{
			name:        "added locally",
			local:       map[string]bool{"a": true, "b": true, "c": true},
			server:      map[string]bool{"a": true},
			wantAdded:   2,
			wantRemoved: 0,
		},
		{
			name:        "removed locally",
			local:       map[string]bool{"a": true},
			server:      map[string]bool{"a": true, "b": true, "c": true},
			wantAdded:   0,
			wantRemoved: 2,
		},
		{
			name:        "mixed changes",
			local:       map[string]bool{"a": true, "c": true},
			server:      map[string]bool{"a": true, "b": true},
			wantAdded:   1, // c
			wantRemoved: 1, // b
		},
		{
			name:        "both empty",
			local:       map[string]bool{},
			server:      map[string]bool{},
			wantAdded:   0,
			wantRemoved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := compareSets(tt.local, tt.server)
			if len(diff.Added) != tt.wantAdded {
				t.Errorf("added: got %d, want %d", len(diff.Added), tt.wantAdded)
			}
			if len(diff.Removed) != tt.wantRemoved {
				t.Errorf("removed: got %d, want %d", len(diff.Removed), tt.wantRemoved)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	short := "short string"
	if got := summarize(short); got != short {
		t.Errorf("short string should not be truncated: got %q", got)
	}

	long := "This is a very long string that exceeds eighty characters and should be truncated with an ellipsis at the end"
	got := summarize(long)
	if len(got) != 80 {
		t.Errorf("long string should be truncated to 80 chars, got %d", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Error("truncated string should end with ...")
	}

	whitespace := "  trimmed  "
	if got := summarize(whitespace); got != "trimmed" {
		t.Errorf("should trim whitespace: got %q", got)
	}
}

func TestCompareItemToServerIdentical(t *testing.T) {
	// We can't easily test compareItemToServer directly since it requires
	// a *fieldmap.FieldMap which has unexported fields. Instead we test
	// the component functions that it uses.

	// Test that identical sets produce no diff
	local := toSet([]string{"Automated", "Draft"})
	server := toSet([]string{"Automated", "Draft"})
	labelDiff := compareSets(local, server)
	if len(labelDiff.Added) != 0 || len(labelDiff.Removed) != 0 {
		t.Error("identical labels should produce no diff")
	}
}

func TestCompareItemToServerDifferentLabels(t *testing.T) {
	local := toSet([]string{"Automated", "New"})
	server := toSet([]string{"Automated", "Draft"})
	labelDiff := compareSets(local, server)

	if len(labelDiff.Added) != 1 {
		t.Errorf("should have 1 added label, got %d", len(labelDiff.Added))
	}
	if len(labelDiff.Removed) != 1 {
		t.Errorf("should have 1 removed label, got %d", len(labelDiff.Removed))
	}

	// Verify the actual values
	addedSet := toSet(labelDiff.Added)
	if !addedSet["New"] {
		t.Error("'New' should be in added")
	}

	removedSet := toSet(labelDiff.Removed)
	if !removedSet["Draft"] {
		t.Error("'Draft' should be in removed")
	}
}

func TestCompareItemToServerLinks(t *testing.T) {
	local := toSet([]string{"SOFT-123", "SOFT-456", "SOFT-789"})
	server := toSet([]string{"SOFT-123", "SOFT-456"})
	linkDiff := compareSets(local, server)

	if len(linkDiff.Added) != 1 {
		t.Errorf("should have 1 added link, got %d", len(linkDiff.Added))
	}
	addedSet := toSet(linkDiff.Added)
	if !addedSet["SOFT-789"] {
		t.Error("SOFT-789 should be added")
	}
}
