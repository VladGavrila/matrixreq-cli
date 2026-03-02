package itemsync

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// Diff compares items parsed from a file against their server state.
func Diff(svc *service.MatrixService, project string, filePath string, fm *fieldmap.FieldMap) ([]ItemDiff, error) {
	entries, err := ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	var diffs []ItemDiff
	for _, entry := range entries {
		if entry.Action == ActionCreate {
			// New items have no server state to compare
			diffs = append(diffs, ItemDiff{
				ItemRef: entry.Item.Title + " (new)",
				IsEqual: false,
				Differences: map[string][2]string{
					"status": {"new item", "does not exist on server"},
				},
			})
			continue
		}

		// Fetch server state
		serverItem, err := svc.Items.Get(project, entry.Item.ItemRef, false)
		if err != nil {
			diffs = append(diffs, ItemDiff{
				ItemRef: entry.Item.ItemRef,
				IsEqual: false,
				Differences: map[string][2]string{
					"error": {"", fmt.Sprintf("failed to fetch: %v", err)},
				},
			})
			continue
		}

		diff := compareItemToServer(entry.Item, serverItem, fm)
		diffs = append(diffs, diff)
	}

	return diffs, nil
}

// compareItemToServer compares a local ItemDef against a server TrimItem.
func compareItemToServer(local ItemDef, server *api.TrimItem, fm *fieldmap.FieldMap) ItemDiff {
	diff := ItemDiff{
		ItemRef:     local.ItemRef,
		IsEqual:     true,
		Differences: make(map[string][2]string),
	}

	// Compare title
	if local.Title != "" && local.Title != server.Title {
		diff.Differences["title"] = [2]string{local.Title, server.Title}
		diff.IsEqual = false
	}

	// Compare fields against server field values
	if server.FieldValList != nil {
		category := CategoryFromItemRef(local.ItemRef)
		serverFields := make(map[string]string)
		for _, fv := range server.FieldValList.FieldVal {
			// Try to reverse-lookup field label from ID
			if fm != nil {
				catFields := fm.FieldsForCategory(category)
				for label, id := range catFields {
					if id == fv.ID {
						serverFields[label] = fv.Value
						break
					}
				}
			}
		}

		for label, localVal := range local.Fields {
			serverVal, exists := serverFields[label]
			if !exists {
				diff.Differences[label] = [2]string{summarize(localVal), "(not set)"}
				diff.IsEqual = false
			} else if !fieldsEqual(localVal, serverVal) {
				diff.Differences[label] = [2]string{summarize(localVal), summarize(serverVal)}
				diff.IsEqual = false
			}
		}
	}

	// Compare labels (set-based)
	localLabels := toSet(local.Labels)
	serverLabels := toSet(server.Labels)
	labelDiff := compareSets(localLabels, serverLabels)
	if len(labelDiff.Added) > 0 || len(labelDiff.Removed) > 0 {
		diff.LabelDiff = labelDiff
		diff.IsEqual = false
	}

	// Compare up_links (set-based)
	localLinks := toSet(local.UpLinks)
	var serverLinkRefs []string
	for _, link := range server.UpLinkList {
		serverLinkRefs = append(serverLinkRefs, link.ItemRef)
	}
	serverLinks := toSet(serverLinkRefs)
	linkDiff := compareSets(localLinks, serverLinks)
	if len(linkDiff.Added) > 0 || len(linkDiff.Removed) > 0 {
		diff.LinkDiff = linkDiff
		diff.IsEqual = false
	}

	return diff
}

// fieldsEqual compares two field values, handling JSON normalization for step fields.
func fieldsEqual(local, server string) bool {
	// Try JSON normalization (for step fields)
	var localJSON, serverJSON interface{}
	if json.Unmarshal([]byte(local), &localJSON) == nil && json.Unmarshal([]byte(server), &serverJSON) == nil {
		localNorm, _ := json.Marshal(localJSON)
		serverNorm, _ := json.Marshal(serverJSON)
		return string(localNorm) == string(serverNorm)
	}
	// Plain string comparison with whitespace normalization
	return strings.TrimSpace(local) == strings.TrimSpace(server)
}

// toSet converts a string slice to a map for set operations.
func toSet(items []string) map[string]bool {
	s := make(map[string]bool)
	for _, item := range items {
		s[item] = true
	}
	return s
}

// compareSets computes the added/removed differences between two sets.
func compareSets(local, server map[string]bool) *SetDiff {
	diff := &SetDiff{}
	for item := range local {
		if !server[item] {
			diff.Added = append(diff.Added, item)
		}
	}
	for item := range server {
		if !local[item] {
			diff.Removed = append(diff.Removed, item)
		}
	}
	return diff
}

// summarize truncates a string for display, handling long HTML/JSON content.
func summarize(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}
