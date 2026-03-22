# AGENTS.md — MatrixReq CLI (`mxreq`)

> Reference for AI agents working on this codebase. Covers architecture, conventions, patterns, and common tasks.

## Project Overview

**Binary:** `mxreq`
**Module:** `github.com/VladGavrila/matrixreq-cli`
**Go version:** 1.23+
**Direct dependencies:** cobra, viper, lipgloss (charmbracelet)
**Purpose:** CLI for the MatrixALM/MatrixQMS REST API v2.5

## Directory Structure

```
├── main.go                    # Entry point → cli.Execute()
├── Makefile                   # Build targets (macOS ARM64, Linux AMD64)
├── cli/                       # Cobra commands (presentation layer)
│   ├── root.go                # Root command, persistent flags, newService(), Execute()
│   ├── <resource>.go          # One file per resource domain (project, item, user, etc.)
│   └── init_templates.go      # Template scaffolding command
├── internal/
│   ├── api/                   # Hand-written request/response types (OpenAPI-derived)
│   │   ├── types.go           # Core types (ProjectType, CategoryType, FieldType, etc.)
│   │   ├── items.go           # TrimItem, TrimFolder, FieldValType, TrimLink
│   │   ├── requests.go        # CreateItemRequest, UpdateItemRequest, etc.
│   │   ├── responses.go       # AddItemAck, CopyItemAck, ListProjectAndSettings, etc.
│   │   └── <domain>.go        # users, groups, todos, search, jobs, files, audit, merge
│   ├── client/                # HTTP client with token auth
│   │   ├── client.go          # Get/Post/Put/Delete/PostForm/GetRaw methods
│   │   └── errors.go          # APIError type, IsNotFound/IsUnauthorized/IsForbidden
│   ├── config/                # Viper-based config (XDG: ~/.config/mxreq/config.yaml)
│   │   └── config.go          # Config struct, Load(), Save(), Validate(), ConfigPath()
│   ├── service/               # Business logic — interface-based domain services
│   │   ├── service.go         # MatrixService aggregate (holds all services + Client)
│   │   └── <domain>.go        # projects, items, categories, fields, users, groups, etc.
│   ├── output/                # Pluggable formatters
│   │   ├── output.go          # Formatter interface, Print(), PrintItem()
│   │   ├── json.go            # JSON output (MarshalIndent)
│   │   ├── table.go           # Lipgloss table output
│   │   └── text.go            # Plain text output
│   ├── fieldmap/              # Field ID resolution with disk cache
│   ├── itemsync/              # Item parsing (YAML/Python/Go/TypeScript) + sync/diff
│   ├── execution/             # Test execution result upload (TC→XTC mapping)
│   └── templates/             # Embedded code templates (go/python/typescript)
├── tests/                     # Shell-based integration tests
│   ├── helpers.sh             # Test helpers: assert, assert_fail, assert_output_contains
│   ├── run-all.sh             # Test runner
│   └── test-<resource>.sh     # One test script per command group
└── dist/                      # Build output (gitignored)
```

## Architecture & Data Flow

```
CLI command (cli/*.go)
  → newService() → config.Load() + client.New() + service.New()
  → service.<Domain>.<Method>(args)
  → client.Get/Post/Put/Delete(path)
  → HTTP request to MatrixALM API (base URL + /1 suffix)
  → Parse response → output.Print() or output.PrintItem()
```

### Layer Responsibilities

| Layer | Package | Role |
|-------|---------|------|
| Presentation | `cli/` | Cobra commands, flag parsing, output dispatch |
| Business Logic | `internal/service/` | Domain operations, URL construction, response mapping |
| Transport | `internal/client/` | HTTP requests, auth headers, error wrapping |
| Types | `internal/api/` | Request/response structs matching the API schema |
| Config | `internal/config/` | XDG config loading (flags > env > file) |
| Output | `internal/output/` | JSON, table (lipgloss), or text formatting |

## Key Conventions

### Configuration Precedence
Flags > Environment vars > Config file (`~/.config/mxreq/config.yaml`)

| Flag | Env Var | Config Key |
|------|---------|------------|
| `--url` | `MATRIX_URL` | `url` |
| `--token` | `MATRIX_TOKEN` | `token` |
| `--project` / `-p` | `MATRIX_DEFAULT_PROJECT` | `default_project` |
| `--output` / `-o` | — | — |
| `--debug` | — | — |

### Debugging
`--debug` prints every HTTP request and response as pretty-printed JSON to stderr. Output includes method, full URL, request body, response status, and response body. Normal command output goes to stdout unaffected. Implemented in `internal/client/client.go` via `debugRequest()`/`debugResponse()` helpers called from every HTTP method. The flag is passed from `cli/root.go` → `client.New(url, token, debug)`.

### API Base Path
The client automatically appends `/rest/1` to the base URL (e.g., `http://host:8080` → `http://host:8080/rest/1`).

### Authentication
`Authorization: Token <value>` header on every request.

### Write Operations
Commands that modify data require `--reason` / `-r` flag, marked as required via `MarkFlagRequired("reason")`.

## Adding a New Command

### 1. Create or extend a CLI file

Each resource gets one file: `cli/<resource>.go`. Follow this exact pattern:

```go
package cli

import (
    "fmt"
    "github.com/VladGavrila/matrixreq-cli/internal/api"
    "github.com/VladGavrila/matrixreq-cli/internal/output"
    "github.com/spf13/cobra"
)

// Register commands in init()
func init() {
    rootCmd.AddCommand(thingCmd)
    thingCmd.AddCommand(thingListCmd)
    thingCmd.AddCommand(thingGetCmd)
}

// Parent command (no RunE — just a grouping node)
var thingCmd = &cobra.Command{
    Use:     "thing",
    Aliases: []string{"th"},
    Short:   "Manage things",
}

// Subcommand — list
var thingListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all things",
    RunE: func(cmd *cobra.Command, args []string) error {
        svc, err := newService()
        if err != nil {
            return err
        }
        project, err := requireProject()  // if project-scoped
        if err != nil {
            return err
        }

        things, err := svc.Things.List(project)
        if err != nil {
            return err
        }

        // JSON: pass the raw struct
        if getOutputFormat() == "json" {
            return output.PrintItem(getOutputFormat(), things)
        }

        // Table/text: build headers + rows
        headers := []string{"ID", "Name", "Status"}
        var rows [][]string
        for _, t := range things {
            rows = append(rows, []string{
                fmt.Sprint(t.ID), t.Name, t.Status,
            })
        }
        return output.Print(getOutputFormat(), headers, rows)
    },
}
```

**Naming conventions:**
- Parent: `<resource>Cmd` (e.g., `thingCmd`)
- Children: `<resource><Action>Cmd` (e.g., `thingListCmd`, `thingCreateCmd`)
- All registered in `init()` functions

**Flags for write commands:**
```go
func init() {
    thingCreateCmd.Flags().StringP("reason", "r", "", "Reason for creation")
    _ = thingCreateCmd.MarkFlagRequired("reason")
    thingCreateCmd.Flags().String("name", "", "Thing name")
    _ = thingCreateCmd.MarkFlagRequired("name")
}
```

### 2. Add API types

In `internal/api/`, add request/response structs. Use these naming patterns:
- Request structs: `Create<Thing>Request`, `Update<Thing>Request`
- Response structs: `<Thing>Type`, `Add<Thing>Ack`
- List wrappers: named structs used in service layer for JSON unmarshaling

### 3. Add a service

In `internal/service/`:

```go
// Interface
type ThingService interface {
    List(project string) ([]api.ThingType, error)
    Get(project string, id int) (*api.ThingType, error)
    Create(project string, req *api.CreateThingRequest) (*api.AddThingAck, error)
}

// Private implementation
type thingService struct {
    client *client.Client
}

// Constructor
func newThingService(c *client.Client) ThingService {
    return &thingService{client: c}
}
```

Then add the field to `MatrixService` in `service.go` and initialize it in `New()`.

### 4. Add tests

- **Go unit tests:** Table-driven, in `<package>_test.go`, same package
- **Shell integration test:** `tests/test-<resource>.sh` following the existing pattern

## Service Layer Patterns

### MatrixService Aggregate
```go
type MatrixService struct {
    Client     *client.Client
    Projects   ProjectService
    Items      ItemService
    Categories CategoryService
    // ... all domain services
}
```

Commands access services via `svc.<Domain>.<Method>()`. For edge cases not covered by a service, use `svc.Client` directly.

### URL Construction
Services build paths using `fmt.Sprintf` with `url.PathEscape()`:
```go
path := fmt.Sprintf("/%s/cat", url.PathEscape(project))
```

### Response Unmarshaling
Services define private wrapper types for list endpoints:
```go
type getThingListAck struct {
    ThingList []api.ThingType `json:"thingList"`
}
```

## Error Handling

- All errors propagate up via early `return err`
- Wrap with context: `fmt.Errorf("creating thing: %w", err)`
- The root command catches errors and prints to stderr with `os.Exit(1)`
- `SilenceUsage: true` and `SilenceErrors: true` on root command
- Client provides `IsNotFound(err)`, `IsUnauthorized(err)`, `IsForbidden(err)` helpers

## Output Formatting

Three modes controlled by `--output` / `-o` flag (default: `table`):

| Format | Function | Behavior |
|--------|----------|----------|
| `json` | `output.PrintItem(format, struct)` | `json.MarshalIndent` the raw API struct |
| `table` | `output.Print(format, headers, rows)` | Lipgloss table with styled headers |
| `text` | `output.Print(format, headers, rows)` | Plain `header: value` lines |

**Pattern in commands:**
```go
if getOutputFormat() == "json" {
    return output.PrintItem(getOutputFormat(), data)
}
// Build headers + rows for table/text
return output.Print(getOutputFormat(), headers, rows)
```

**Lipgloss table notes:**
- `Row(...string)` method takes variadic strings
- `Headers(...string)` sets column headers
- `StyleFunc` for conditional row/cell styling

## Building & Testing

### Build
```bash
make build          # Both platforms
make macos-arm      # dist/mxreq-macos-arm64
make linux-amd64    # dist/mxreq-linux-amd64
go build -o mxreq . # Local dev build
```

Build flags: `CGO_ENABLED=0`, `-ldflags "-s -w"` (stripped binaries).

### Go Unit Tests
```bash
go test ./...                        # All tests
go test ./internal/fieldmap/...      # Single package
go test -run TestParsePython ./internal/itemsync/...  # Single test
```

Tests use table-driven patterns, `t.TempDir()` for isolation, `t.Setenv()` for env vars. No mock frameworks — tests construct real `api.*` types.

### Shell Integration Tests
```bash
./tests/run-all.sh [path-to-binary]  # Run all
./tests/test-item.sh [path-to-binary] # Run one

# Environment controls:
MXREQ_LIVE_TESTS=1    # Enable server integration tests (default: skip)
MXREQ_ADMIN_TESTS=1   # Enable admin/destructive tests
MXREQ_TEST_PROJECT=XX  # Project for live tests (default: SW_Sandbox)
MXREQ_TEST_FOLDER=XX   # Folder for item/folder tests
```

Each test script has three phases:
1. **Offline help tests** — always run, verify `--help` output
2. **Validation tests** — always run, verify required flag/arg rejection
3. **Live tests** — gated by `MXREQ_LIVE_TESTS=1`, test actual API calls

Shell test helpers: `assert`, `assert_fail`, `assert_output_contains`, `assert_output_not_contains`, `print_report`.

## Specialized Subsystems

### Field Map (`internal/fieldmap/`)
Resolves field labels to numeric IDs. Caches in `~/.config/mxreq/fieldcache.json`. Keys: `"Category.FieldLabel"` → field ID. Use `LoadOrFetch(svc, project)` to get or populate.

### Item Sync (`internal/itemsync/`)
Parses test files (Python/Go/TypeScript) and YAML definition files to sync items with the server. Supports:
- YAML docstrings embedded in test functions (`---` delimited)
- Standalone YAML definition files (`items: [...]`)
- `NEW_` prefix convention for items to be created
- Automatic function renaming after creation (e.g., `test_NEW_foo` → `test_TC_42_foo`)
- Diff comparison (fields, labels, links) with whitespace/JSON normalization

### Execution Upload (`internal/execution/`)
Uploads test execution results. Maps test cases (TC) to execution cases (XTC) via title parsing (expects `"Title (TC-1377)"` format). Tracks worst-case requirement coverage across multiple test cases.

### Templates (`internal/templates/`)
Embedded templates for Go/Python/TypeScript scaffolding. Used by `cli/init_templates.go`.

## Common Pitfalls

1. **Base URL `/rest/1` suffix** — The client appends `/rest/1` automatically. Don't include it in service paths.
2. **URL encoding** — Always use `url.PathEscape()` for path segments and `url.QueryEscape()` for query params.
3. **Lipgloss table API** — `Row()` takes variadic `...string`, not `[]string`. Use `Row(slice...)` to unpack.
4. **Config validation** — `config.Validate()` requires both `URL` and `Token`. Commands that don't need the API (like `config init`, `version`) skip `newService()`.
5. **JSON output** — Always check `getOutputFormat() == "json"` before building table rows, to avoid unnecessary work.
6. **Required flags** — Use `MarkFlagRequired()` in `init()` — cobra handles validation before `RunE` runs.
7. **Project resolution** — `requireProject()` checks flag first, then config `default_project`. Fail early if neither set.

## Release

When the user requests release notes after completing an implementation:

1. **Ask the user what version number to release** — do not assume or auto-increment.
2. Once the user provides the version, update `var version` in `cli/root.go`:
   ```go
   var version = "<new-version>"
   ```
3. Confirm the build passes with `make build` and that `mxreq --version` prints the new version.

4. Write the release notes in a RELEASE_<new-version>.md file