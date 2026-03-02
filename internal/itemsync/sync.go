package itemsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// SyncOptions controls sync behavior.
type SyncOptions struct {
	DryRun     bool // Show what would happen without making changes
	CreateOnly bool // Only process new items
	UpdateOnly bool // Only process existing items
	Yes        bool // Skip confirmation prompts
}

// SyncResult holds the outcome of a sync operation.
type SyncResult struct {
	Created []SyncResultEntry
	Updated []SyncResultEntry
	Skipped []SyncResultEntry
	Errors  []SyncResultEntry
}

// SyncResultEntry describes what happened to a single item.
type SyncResultEntry struct {
	Title   string
	ItemRef string
	Folder  string
	Action  string // "created", "updated", "skipped", "error"
	Error   string
}

// Sync parses a file and creates/updates items in Matrix.
func Sync(svc *service.MatrixService, project string, filePath string, fm *fieldmap.FieldMap, opts SyncOptions) (*SyncResult, error) {
	entries, err := ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	if len(entries) == 0 {
		return &SyncResult{}, nil
	}

	result := &SyncResult{}
	ext := strings.ToLower(filepath.Ext(filePath))
	isSourceFile := ext == ".py" || ext == ".go" || ext == ".ts"

	// Read source file content for potential modification
	var fileContent string
	if isSourceFile {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading source file: %w", err)
		}
		fileContent = string(data)
	}

	fileModified := false

	for _, entry := range entries {
		// Apply filters
		if opts.CreateOnly && entry.Action != ActionCreate {
			result.Skipped = append(result.Skipped, SyncResultEntry{
				Title:   entry.Item.Title,
				ItemRef: entry.Item.ItemRef,
				Action:  "skipped",
			})
			continue
		}
		if opts.UpdateOnly && entry.Action != ActionUpdate {
			result.Skipped = append(result.Skipped, SyncResultEntry{
				Title:   entry.Item.Title,
				ItemRef: entry.Item.ItemRef,
				Action:  "skipped",
			})
			continue
		}

		// Determine category from folder or item ref
		category := CategoryFromFolderRef(entry.Item.Folder)
		if category == "" && entry.Item.ItemRef != "" {
			category = CategoryFromItemRef(entry.Item.ItemRef)
		}

		// Resolve field labels to IDs
		fieldVals, err := resolveFields(fm, category, entry.Item.Fields)
		if err != nil {
			result.Errors = append(result.Errors, SyncResultEntry{
				Title:  entry.Item.Title,
				Action: "error",
				Error:  fmt.Sprintf("resolving fields: %v", err),
			})
			continue
		}

		upLinks := strings.Join(entry.Item.UpLinks, ",")
		labelsStr := strings.Join(entry.Item.Labels, ",")

		switch entry.Action {
		case ActionCreate:
			if opts.DryRun {
				result.Created = append(result.Created, SyncResultEntry{
					Title:  entry.Item.Title,
					Folder: entry.Item.Folder,
					Action: "would create",
				})
				continue
			}

			req := &api.CreateItemRequest{
				Title:  entry.Item.Title,
				Folder: entry.Item.Folder,
				Reason: "synced by mxreq",
				Fields: fieldVals,
				Labels: entry.Item.Labels,
			}

			ack, err := svc.Items.Create(project, req)
			if err != nil {
				result.Errors = append(result.Errors, SyncResultEntry{
					Title:  entry.Item.Title,
					Action: "error",
					Error:  fmt.Sprintf("creating item: %v", err),
				})
				continue
			}

			// Create uplinks
			if upLinks != "" {
				newRef := fmt.Sprintf("%s-%d", category, ack.Serial)
				for _, ref := range entry.Item.UpLinks {
					_ = svc.Items.CreateLink(project, ref, newRef, "synced by mxreq")
				}
			}

			newRef := fmt.Sprintf("%s-%d", category, ack.Serial)
			result.Created = append(result.Created, SyncResultEntry{
				Title:   entry.Item.Title,
				ItemRef: newRef,
				Folder:  entry.Item.Folder,
				Action:  "created",
			})

			// Rename function in source file
			if isSourceFile && entry.Location != nil {
				fileContent, _ = UpdateFunctionName(
					fileContent,
					entry.Location.LineNumber,
					entry.Location.FunctionLine,
					newRef,
					ext,
				)
				fileModified = true
			}

		case ActionUpdate:
			if opts.DryRun {
				result.Updated = append(result.Updated, SyncResultEntry{
					Title:   entry.Item.Title,
					ItemRef: entry.Item.ItemRef,
					Folder:  entry.Item.Folder,
					Action:  "would update",
				})
				continue
			}

			updateReq := &api.UpdateItemRequest{
				Title:      entry.Item.Title,
				Reason:     "synced by mxreq",
				Fields:     fieldVals,
				Labels:     entry.Item.Labels,
				OnlyThose:  true,
				OnlyLabels: true,
			}

			_, err := svc.Items.Update(project, entry.Item.ItemRef, updateReq)
			if err != nil {
				result.Errors = append(result.Errors, SyncResultEntry{
					Title:   entry.Item.Title,
					ItemRef: entry.Item.ItemRef,
					Action:  "error",
					Error:   fmt.Sprintf("updating item: %v", err),
				})
				continue
			}

			// Update uplinks if specified
			if upLinks != "" {
				for _, ref := range entry.Item.UpLinks {
					_ = svc.Items.CreateLink(project, ref, entry.Item.ItemRef, "synced by mxreq")
				}
			}

			// Update labels if specified
			_ = labelsStr // labels are included in the update request

			result.Updated = append(result.Updated, SyncResultEntry{
				Title:   entry.Item.Title,
				ItemRef: entry.Item.ItemRef,
				Folder:  entry.Item.Folder,
				Action:  "updated",
			})
		}
	}

	// Write back modified source file
	if isSourceFile && fileModified && !opts.DryRun {
		if err := os.WriteFile(filePath, []byte(fileContent), 0o644); err != nil {
			return result, fmt.Errorf("writing modified source file: %w", err)
		}
	}

	return result, nil
}

// resolveFields converts field label→value pairs to FieldValSetType using the field map.
func resolveFields(fm *fieldmap.FieldMap, category string, fields map[string]string) ([]api.FieldValSetType, error) {
	var vals []api.FieldValSetType
	for label, value := range fields {
		id, err := fm.Resolve(category, label)
		if err != nil {
			return nil, fmt.Errorf("field %s.%s: %w", category, label, err)
		}
		vals = append(vals, api.FieldValSetType{
			ID:    id,
			Value: value,
		})
	}
	return vals, nil
}
