package itemsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// --- Source file parsing (.py, .go, .ts) ---

// FindTestFunctions finds all test function definitions matching a keyword in the content.
func FindTestFunctions(content string, keyword string) []string {
	var matches []string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(strings.TrimSpace(line), keyword) {
			matches = append(matches, line)
		}
	}
	return matches
}

// KeywordForFileType returns the appropriate search keyword for a file type and mode.
// mode is "new" (for creating) or "update" (for updating).
func KeywordForFileType(ext string, mode string) string {
	keywords := map[string]map[string]string{
		".py": {
			"new":    "def test_NEW_",
			"update": "def test_TC_",
		},
		".go": {
			"new":    "func Test_NEW_",
			"update": "func Test_TC_",
		},
		".ts": {
			"new":    "test('TC-NEW-",
			"update": "test('TC-",
		},
	}
	if m, ok := keywords[ext]; ok {
		return m[mode]
	}
	return ""
}

// ExtractTCNumber extracts a TC reference from a function definition line.
// Returns e.g. "TC-123" or empty string if not found.
func ExtractTCNumber(functionLine string) string {
	if strings.Contains(functionLine, "TC_") {
		// Python or Go format: TC_123_
		idx := strings.Index(functionLine, "TC_") + 3
		num := "TC-"
		for idx < len(functionLine) && functionLine[idx] != '_' {
			num += string(functionLine[idx])
			idx++
		}
		if len(num) > 3 {
			return num
		}
		return ""
	}

	if strings.Contains(functionLine, "test('TC-") {
		// TypeScript format: test('TC-123-title'
		idx := strings.Index(functionLine, "TC-") + 3
		num := "TC-"
		for idx < len(functionLine) {
			ch := functionLine[idx]
			if ch == '\'' || ch == ',' || ch == ')' {
				break
			}
			if ch == '-' && len(num) > 3 {
				// Past the number, into the title
				break
			}
			if ch >= '0' && ch <= '9' {
				num += string(ch)
			}
			idx++
		}
		if len(num) > 3 {
			return num
		}
		return ""
	}

	return ""
}

// ParseDocstring extracts a YAML docstring from source file content near the given function line.
// Supports Python triple-quote (""") and JavaScript/Go block comment (/* */) styles.
func ParseDocstring(content string, functionLine string) *ParsedLocation {
	lines := strings.Split(content, "\n")

	// Find the function line
	lineNumber := -1
	for idx, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(functionLine) {
			lineNumber = idx
			break
		}
	}
	if lineNumber < 0 {
		return nil
	}

	// Look for docstring start within 10 lines after function
	docStart := -1
	commentStyle := "" // "python" or "js"
	searchEnd := lineNumber + 10
	if searchEnd > len(lines) {
		searchEnd = len(lines)
	}

	for idx := lineNumber + 1; idx < searchEnd; idx++ {
		trimmed := strings.TrimSpace(lines[idx])
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
			docStart = idx
			commentStyle = "python"
			break
		}
		if strings.HasPrefix(trimmed, "/*") {
			docStart = idx
			commentStyle = "js"
			break
		}
	}
	if docStart < 0 {
		return nil
	}

	// Find docstring end
	docEnd := -1
	if commentStyle == "python" {
		quoteChar := strings.TrimSpace(lines[docStart])[:3]
		for idx := docStart + 1; idx < len(lines); idx++ {
			if strings.Contains(lines[idx], quoteChar) {
				docEnd = idx
				break
			}
		}
	} else {
		for idx := docStart + 1; idx < len(lines); idx++ {
			if strings.Contains(lines[idx], "*/") {
				docEnd = idx
				break
			}
		}
	}
	if docEnd < 0 {
		return nil
	}

	// Extract docstring content
	docContent := strings.Join(lines[docStart:docEnd+1], "\n")
	docContent = strings.TrimSpace(docContent)

	// Remove comment markers
	if commentStyle == "python" {
		docContent = strings.TrimPrefix(docContent, `"""`)
		docContent = strings.TrimPrefix(docContent, `'''`)
		docContent = strings.TrimSuffix(docContent, `"""`)
		docContent = strings.TrimSuffix(docContent, `'''`)
	} else {
		if strings.HasPrefix(docContent, "/**") {
			docContent = docContent[3:]
		} else if strings.HasPrefix(docContent, "/*") {
			docContent = docContent[2:]
		}
		if strings.HasSuffix(docContent, "*/") {
			docContent = docContent[:len(docContent)-2]
		}
		// Remove common leading whitespace for JS block comments
		docContent = dedentBlock(docContent)
	}

	// Extract YAML between --- markers
	yamlContent := extractYAML(docContent)
	if yamlContent == "" {
		return nil
	}

	item, err := parseYAMLDocstring(yamlContent)
	if err != nil {
		return nil
	}

	// If item_ref wasn't in the docstring, try to extract from function name
	if item.ItemRef == "" {
		item.ItemRef = ExtractTCNumber(functionLine)
	}

	return &ParsedLocation{
		LineNumber:   lineNumber,
		FunctionLine: functionLine,
		Item:         *item,
	}
}

// dedentBlock removes common leading whitespace from a multi-line block.
func dedentBlock(s string) string {
	rawLines := strings.Split(s, "\n")
	// Find minimum indent of non-empty lines
	minIndent := -1
	for _, line := range rawLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return s
	}
	var cleaned []string
	for _, line := range rawLines {
		if strings.TrimSpace(line) == "" {
			cleaned = append(cleaned, "")
		} else if len(line) >= minIndent {
			cleaned = append(cleaned, line[minIndent:])
		} else {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

// extractYAML extracts YAML content between --- markers.
func extractYAML(content string) string {
	re := regexp.MustCompile(`(?s)---\s*\n(.*?)\n\s*---`)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

// yamlDocstring is the intermediate YAML structure for parsing docstrings.
// Supports both the generic "fields" format and the legacy flat format.
type yamlDocstring struct {
	Title       string            `yaml:"title"`
	ItemRef     string            `yaml:"item_ref"`
	Folder      string            `yaml:"folder"`
	Fields      map[string]string `yaml:"fields"`
	Labels      []string          `yaml:"labels"`
	UpLinks     string            `yaml:"up_links"`
	// Legacy flat format fields (TC-specific, converted to Fields map)
	Description string        `yaml:"description"`
	Assumptions string        `yaml:"assumptions"`
	Steps       []yamlStep    `yaml:"steps"`
}

type yamlStep struct {
	Action          string `yaml:"action"`
	Expected        string `yaml:"expected"`
	RequirementLink string `yaml:"RequirementLink"`
}

// parseYAMLDocstring parses YAML content (without --- markers) into an ItemDef.
func parseYAMLDocstring(yamlContent string) (*ItemDef, error) {
	var doc yamlDocstring
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML docstring: %w", err)
	}

	fields := doc.Fields
	if fields == nil {
		fields = make(map[string]string)
	}

	// Handle legacy flat format: description, assumptions, steps at top level
	if doc.Description != "" {
		if _, ok := fields["Description"]; !ok {
			fields["Description"] = strings.TrimSpace(doc.Description)
		}
	}
	if doc.Assumptions != "" {
		if _, ok := fields["Assumptions"]; !ok {
			fields["Assumptions"] = strings.TrimSpace(doc.Assumptions)
		}
	}
	if len(doc.Steps) > 0 {
		if _, ok := fields["Steps"]; !ok {
			stepsJSON, err := stepsToJSON(doc.Steps)
			if err != nil {
				return nil, fmt.Errorf("serializing steps: %w", err)
			}
			fields["Steps"] = stepsJSON
		}
	}

	// Parse up_links from comma-separated string
	var upLinks []string
	if doc.UpLinks != "" {
		for _, ref := range strings.Split(doc.UpLinks, ",") {
			ref = strings.TrimSpace(ref)
			if ref != "" {
				upLinks = append(upLinks, ref)
			}
		}
	}

	return &ItemDef{
		Title:   doc.Title,
		ItemRef: doc.ItemRef,
		Folder:  doc.Folder,
		Fields:  fields,
		Labels:  doc.Labels,
		UpLinks: upLinks,
	}, nil
}

// stepJSON is the JSON structure for a test step (matches Matrix API format).
type stepJSON struct {
	Action          string `json:"action"`
	Expected        string `json:"expected"`
	RequirementLink string `json:"RequirementLink"`
}

// stepsToJSON serializes YAML steps to the compact JSON format used by Matrix.
func stepsToJSON(steps []yamlStep) (string, error) {
	var jsonSteps []stepJSON
	for _, s := range steps {
		expected := s.Expected
		if expected == "" {
			expected = "N/A"
		}
		jsonSteps = append(jsonSteps, stepJSON{
			Action:          s.Action,
			Expected:        expected,
			RequirementLink: s.RequirementLink,
		})
	}
	data, err := json.Marshal(jsonSteps)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateFunctionName replaces NEW_ or TC-NEW- in a function name with the actual TC number.
// Returns the updated content and the new function line.
func UpdateFunctionName(content string, lineNumber int, oldLine string, newRef string, ext string) (string, string) {
	lines := strings.Split(content, "\n")
	if lineNumber < 0 || lineNumber >= len(lines) {
		return content, oldLine
	}

	newLine := oldLine
	switch ext {
	case ".py":
		// def test_NEW_Name -> def test_TC_1234_Name
		tcUnderscore := strings.ReplaceAll(newRef, "-", "_")
		if len(oldLine) > 13 {
			newLine = "def test_" + tcUnderscore + "_" + oldLine[13:]
		}
	case ".go":
		// func Test_NEW_Name -> func Test_TC_1234_Name
		tcUnderscore := strings.ReplaceAll(newRef, "-", "_")
		if len(oldLine) > 14 {
			newLine = "func Test_" + tcUnderscore + "_" + oldLine[14:]
		}
	case ".ts":
		// test('TC-NEW-title', ...) -> test('TC-1234-title', ...)
		tcDash := strings.ReplaceAll(newRef, "_", "-")
		if strings.Contains(oldLine, "TC-NEW-") {
			idx := strings.Index(oldLine, "TC-NEW-")
			rest := oldLine[idx+7:] // Skip "TC-NEW-"
			newLine = oldLine[:idx] + tcDash + "-" + rest
		}
	}

	lines[lineNumber] = newLine
	return strings.Join(lines, "\n"), newLine
}

// --- YAML definition file parsing (.yaml) ---

// yamlDefinitionFile is the structure of a .yaml item definition file.
type yamlDefinitionFile struct {
	Items []yamlItemDef `yaml:"items"`
}

type yamlItemDef struct {
	Title   string            `yaml:"title"`
	ItemRef string            `yaml:"item_ref"`
	Folder  string            `yaml:"folder"`
	Fields  map[string]string `yaml:"fields"`
	Labels  []string          `yaml:"labels"`
	UpLinks string            `yaml:"up_links"`
}

// ParseYAMLDefinitions parses a .yaml file containing item definitions.
func ParseYAMLDefinitions(filePath string) ([]ItemDef, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return ParseYAMLDefinitionsFromString(string(data))
}

// ParseYAMLDefinitionsFromString parses YAML item definitions from a string.
func ParseYAMLDefinitionsFromString(content string) ([]ItemDef, error) {
	var file yamlDefinitionFile
	if err := yaml.Unmarshal([]byte(content), &file); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	var items []ItemDef
	for _, def := range file.Items {
		fields := def.Fields
		if fields == nil {
			fields = make(map[string]string)
		}

		var upLinks []string
		if def.UpLinks != "" {
			for _, ref := range strings.Split(def.UpLinks, ",") {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					upLinks = append(upLinks, ref)
				}
			}
		}

		items = append(items, ItemDef{
			Title:   def.Title,
			ItemRef: def.ItemRef,
			Folder:  def.Folder,
			Fields:  fields,
			Labels:  def.Labels,
			UpLinks: upLinks,
		})
	}

	return items, nil
}

// --- Common helpers ---

// ParseFile dispatches parsing based on file extension.
// Returns a list of SyncEntry items ready for sync.
func ParseFile(filePath string) ([]SyncEntry, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".yaml", ".yml":
		return parseYAMLFile(filePath)
	case ".py", ".go", ".ts":
		return parseSourceFile(filePath, ext)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func parseYAMLFile(filePath string) ([]SyncEntry, error) {
	items, err := ParseYAMLDefinitions(filePath)
	if err != nil {
		return nil, err
	}

	var entries []SyncEntry
	for _, item := range items {
		action := ActionCreate
		if item.ItemRef != "" {
			action = ActionUpdate
		}
		entries = append(entries, SyncEntry{
			Action: action,
			Item:   item,
		})
	}
	return entries, nil
}

func parseSourceFile(filePath string, ext string) ([]SyncEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	content := string(data)

	var entries []SyncEntry

	// Find NEW_ functions (create)
	newKeyword := KeywordForFileType(ext, "new")
	if newKeyword != "" {
		newFuncs := FindTestFunctions(content, newKeyword)
		for _, funcLine := range newFuncs {
			// For TypeScript update keyword "test('TC-" also matches "test('TC-NEW-"
			// so we only look for NEW_ functions here
			loc := ParseDocstring(content, funcLine)
			if loc != nil {
				entries = append(entries, SyncEntry{
					Action:   ActionCreate,
					Item:     loc.Item,
					Location: loc,
				})
			}
		}
	}

	// Find TC_### functions (update)
	updateKeyword := KeywordForFileType(ext, "update")
	if updateKeyword != "" {
		updateFuncs := FindTestFunctions(content, updateKeyword)
		for _, funcLine := range updateFuncs {
			// For TypeScript, skip TC-NEW- matches (already handled above)
			if ext == ".ts" && strings.Contains(funcLine, "TC-NEW-") {
				continue
			}
			loc := ParseDocstring(content, funcLine)
			if loc != nil {
				// Ensure item_ref is set from function name
				if loc.Item.ItemRef == "" {
					loc.Item.ItemRef = ExtractTCNumber(funcLine)
				}
				entries = append(entries, SyncEntry{
					Action:   ActionUpdate,
					Item:     loc.Item,
					Location: loc,
				})
			}
		}
	}

	return entries, nil
}

// CategoryFromFolderRef extracts the category short label from a folder reference.
// e.g., "F-TC-186" → "TC", "F-REQ-1" → "REQ"
func CategoryFromFolderRef(folderRef string) string {
	if !strings.HasPrefix(folderRef, "F-") {
		return ""
	}
	rest := folderRef[2:]
	idx := strings.LastIndex(rest, "-")
	if idx < 0 {
		return rest
	}
	return rest[:idx]
}

// CategoryFromItemRef extracts the category short label from an item reference.
// e.g., "TC-1377" → "TC", "REQ-42" → "REQ"
func CategoryFromItemRef(itemRef string) string {
	idx := strings.Index(itemRef, "-")
	if idx < 0 {
		return ""
	}
	return itemRef[:idx]
}
