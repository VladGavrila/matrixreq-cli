package fieldmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
)

// cacheFileName is the name of the field cache file within the mxreq config directory.
const cacheFileName = "fieldcache.json"

// FieldMap holds resolved field IDs for all categories in a project.
// Keys are "Category.FieldLabel" (e.g., "TC.Steps", "REQ.Description").
type FieldMap struct {
	fields map[string]int
}

// Entry represents a single field mapping for display purposes.
type Entry struct {
	Category string
	Label    string
	FieldID  int
}

// Resolve returns the field ID for a given category and field label.
func (fm *FieldMap) Resolve(category, label string) (int, error) {
	key := category + "." + label
	id, ok := fm.fields[key]
	if !ok {
		return 0, fmt.Errorf("field %q not found in field map", key)
	}
	return id, nil
}

// DescriptionField returns the field ID for the content/description field of a category.
// It looks for "Description" first, then "Contents", then falls back to the only field if exactly one exists.
func (fm *FieldMap) DescriptionField(category string) (int, error) {
	fields := fm.FieldsForCategory(category)
	if len(fields) == 0 {
		return 0, fmt.Errorf("category %q has no fields", category)
	}
	for _, name := range []string{"Description", "Contents"} {
		if id, ok := fields[name]; ok {
			return id, nil
		}
	}
	if len(fields) == 1 {
		for _, id := range fields {
			return id, nil
		}
	}
	var names []string
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return 0, fmt.Errorf("category %q has no Description/Contents field; available fields: %s — use --field name=value", category, strings.Join(names, ", "))
}

// FieldsForCategory returns all field label→ID mappings for a given category.
func (fm *FieldMap) FieldsForCategory(category string) map[string]int {
	result := make(map[string]int)
	prefix := category + "."
	for key, id := range fm.fields {
		if strings.HasPrefix(key, prefix) {
			label := strings.TrimPrefix(key, prefix)
			result[label] = id
		}
	}
	return result
}

// Entries returns all field mappings as a sorted slice for display.
func (fm *FieldMap) Entries() []Entry {
	var entries []Entry
	for key, id := range fm.fields {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) == 2 {
			entries = append(entries, Entry{
				Category: parts[0],
				Label:    parts[1],
				FieldID:  id,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].Label < entries[j].Label
	})
	return entries
}

// Categories returns all unique category short labels in the field map.
func (fm *FieldMap) Categories() []string {
	seen := make(map[string]bool)
	for key := range fm.fields {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) == 2 {
			seen[parts[0]] = true
		}
	}
	var cats []string
	for c := range seen {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

// LoadOrFetch loads field mappings from cache or fetches from the API.
func LoadOrFetch(svc *service.MatrixService, project string) (*FieldMap, error) {
	// Try cache first
	cache, err := loadCache()
	if err == nil {
		if fields, ok := cache[project]; ok {
			return &FieldMap{fields: fields}, nil
		}
	}

	// Cache miss - fetch from API
	info, err := svc.Projects.Get(project)
	if err != nil {
		return nil, fmt.Errorf("fetching project info: %w", err)
	}

	fields := buildFieldMap(info)
	fm := &FieldMap{fields: fields}

	// Save to cache
	if cache == nil {
		cache = make(map[string]map[string]int)
	}
	cache[project] = fields
	_ = saveCache(cache)

	return fm, nil
}

// buildFieldMap extracts field ID mappings from a ProjectInfo response.
func buildFieldMap(info *api.ProjectInfo) map[string]int {
	fields := make(map[string]int)
	for _, catExt := range info.CategoryList.CategoryExtended {
		shortLabel := catExt.Category.ShortLabel
		if shortLabel == "" {
			continue
		}
		for _, f := range catExt.FieldList.Field {
			if f.Label == "" {
				continue
			}
			key := shortLabel + "." + f.Label
			fields[key] = f.ID
		}
	}
	return fields
}

// Clear removes the cached field map for a specific project.
func Clear(project string) error {
	cache, err := loadCache()
	if err != nil {
		return nil // No cache file is fine
	}
	delete(cache, project)
	return saveCache(cache)
}

// ClearAll removes all cached field maps.
func ClearAll() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing cache: %w", err)
	}
	return nil
}

// cachePath returns the full path to the field cache file.
func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mxreq", cacheFileName), nil
}

// loadCache reads the per-project field cache from disk.
func loadCache() (map[string]map[string]int, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cache map[string]map[string]int
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parsing field cache: %w", err)
	}
	return cache, nil
}

// saveCache writes the per-project field cache to disk.
func saveCache(cache map[string]map[string]int) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling field cache: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}
