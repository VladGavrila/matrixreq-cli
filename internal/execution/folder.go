package execution

import (
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// resolveFolderItems returns the child items of a folder.
//
// It first tries GET /item/{folder} (svc.Items.GetFolder). Some Matrix
// instances return that folder as "partial" with an empty itemList, which
// leaves xtc upload/stats unable to find any XTCs ("no XTCs found in folder").
// When the itemList is empty, it falls back to the category tree
// (GET /cat/{category}, svc.Categories.Get) and locates the folder's subtree
// there, which is returned fully populated.
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
