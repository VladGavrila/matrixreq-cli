package itemsync

// ItemDef represents any Matrix item parsed from a YAML docstring or definition file.
// Works for TC, REQ, SPEC, RISK, or any category.
type ItemDef struct {
	Title   string            // Item title
	ItemRef string            // e.g., "TC-1377", "REQ-42" (empty for new items)
	Folder  string            // Target folder, e.g., "F-TC-186", "F-REQ-1"
	Fields  map[string]string // Field label → value (e.g., "Description" → "<p>...</p>")
	Labels  []string          // Item labels
	UpLinks []string          // Uplink references (e.g., "SOFT-123", "SPEC-45")
}

// ItemDiff represents the result of comparing a local ItemDef vs. a server item.
type ItemDiff struct {
	ItemRef     string                    // Item reference (or title for new items)
	IsEqual     bool                      // True if no differences
	Differences map[string][2]string      // field → [local, server]
	LabelDiff   *SetDiff                  // Set-based label comparison
	LinkDiff    *SetDiff                  // Set-based link comparison
}

// SetDiff represents a set-based comparison result.
type SetDiff struct {
	Added   []string // Present locally but not on server
	Removed []string // Present on server but not locally
}

// ParsedLocation holds a parsed item along with its source file location metadata.
// Used for source file parsing (.py, .go, .ts) where we need to track line numbers.
type ParsedLocation struct {
	LineNumber   int     // Line number of the function definition (0-based)
	FunctionLine string  // The full function definition line
	Item         ItemDef // The parsed item definition
}

// SyncAction indicates what operation should be performed for an item.
type SyncAction int

const (
	ActionCreate SyncAction = iota
	ActionUpdate
)

// SyncEntry pairs a parsed item with its intended action and source location.
type SyncEntry struct {
	Action   SyncAction
	Item     ItemDef
	Location *ParsedLocation // Non-nil for source file items
}
