package execution

import (
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// resolveFolderItems returns the child items of a folder.
//
// It first tries GET /item/{folder} (svc.Items.GetFolder). Matrix only expands
// children for a category-root folder there; every *nested* folder is returned
// as a "partial" stub with an empty itemList (non-recursive, lazy tree loading).
// Since xtc upload/stats target a nested folder (e.g. F-XTC-123), that call
// yields no XTCs and the operation fails with "no XTCs found in folder".
//
// When the itemList is empty, it falls back to the category tree
// (GET /cat/{category}, svc.Categories.Get) — the authoritative source that
// returns the full recursive tree — and locates the folder's subtree within it.
// The fast path is kept because it is a single small request when the target is
// a category root, and it stays correct on instances where /item does expand.
func resolveFolderItems(svc *service.MatrixService, project, folderRef string) ([]api.TrimFolder, error) {
	folder, err := svc.Items.GetFolder(project, folderRef, false)
	if err != nil {
		return nil, err
	}
	if len(folder.ItemList) > 0 {
		return folder.ItemList, nil
	}

	// Fallback: resolve the subtree from the category tree.
	cat := categoryFromRef(folderRef)
	if cat == "" {
		return folder.ItemList, nil
	}
	catFull, cerr := svc.Categories.Get(project, cat)
	if cerr != nil {
		// Return the (empty) itemList and let the caller report "no XTCs found".
		return folder.ItemList, nil
	}
	if sub := findFolderSubtree(catFull.Folder.ItemList, folderRef); sub != nil {
		return sub, nil
	}
	return folder.ItemList, nil
}

// categoryFromRef derives a category short label from a folder or item ref,
// e.g. "F-XTC-123" -> "XTC", "XTC-308" -> "XTC".
func categoryFromRef(ref string) string {
	parts := strings.Split(ref, "-")
	if len(parts) >= 2 {
		// "F-XTC-123" -> parts[1], "XTC-308" -> parts[0].
		if strings.EqualFold(parts[0], "F") {
			return parts[1]
		}
		return parts[0]
	}
	return ""
}

// findFolderSubtree recursively locates the folder with the given ref within a
// category tree and returns its child itemList, or nil if not found.
func findFolderSubtree(items []api.TrimFolder, folderRef string) []api.TrimFolder {
	for i := range items {
		if items[i].ItemRef == folderRef {
			return items[i].ItemList
		}
		if len(items[i].ItemList) > 0 {
			if sub := findFolderSubtree(items[i].ItemList, folderRef); sub != nil {
				return sub
			}
		}
	}
	return nil
}
