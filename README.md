# mxreq

A command-line interface for the [MatrixREQ](https://docs.matrixreq.com/en/) REST API.

## Installation

Download a pre-built binary from the [Releases](https://github.com/VladGavrila/matrixreq-cli/releases) page

### Build from source

```bash
git clone https://github.com/VladGavrila/matrixreq-cli.git
cd matrixreq-cli
go build -o mxreq .
```

## Configuration

Run `mxreq config init` to create a config file at `~/.config/mxreq/config.yaml`:

```yaml
url: https://your-instance.matrixreq.com
token: your-api-token
default_project: PROJ
```

You can also use flags or environment variables:

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `MATRIX_URL` | Matrix instance URL |
| `--token` | `MATRIX_TOKEN` | API token |
| `--project` / `-p` | `MATRIX_DEFAULT_PROJECT` | Default project |
| `--output` / `-o` | — | Output format: `table`, `json`, `text` |
| `--debug` | — | Print HTTP details to stderr |

Precedence: flags > environment variables > config file.

## Usage

```bash
# Help
mxreq --help
mxreq <command> --help
mxreq <command> <sub-command> --help

# Projects
mxreq project list
mxreq project get PROJ
mxreq project tree PROJ

# Items
mxreq item get -p PROJ REQ-1
mxreq item create -p PROJ --category REQ --title "New requirement" -r "initial creation"
mxreq item update -p PROJ REQ-1 --title "Updated title" -r "fixed typo"

# Categories & Fields
mxreq category list -p PROJ
mxreq field get -p PROJ REQ 101

# Users & Groups
mxreq user list
mxreq group list

# Search, Files, Jobs
mxreq search -p PROJ --term "login"
mxreq file list -p PROJ REQ-1
mxreq job list

# Todos
mxreq todo list
mxreq todo create --text "Review spec" --project PROJ

# Branching & Merging
mxreq branch create -p PROJ --label BRANCH --reason "feature work"
mxreq branch merge -p PROJ --label BRANCH --reason "merge to main"

# Export & Import
mxreq export -p PROJ
mxreq import -p PROJ --file export.xml -r "restore"

# Admin
mxreq admin status
```

**Write operations require a `--reason` / `-r` flag.**

## Output Formats

```bash
mxreq project list                  # table (default)
mxreq project list -o json          # JSON
mxreq project list -o text          # plain text
```

## Testing

```bash
go test ./...                                  # unit tests
./tests/run-all.sh ./mxreq                     # integration tests (offline)
MXREQ_LIVE_TESTS=1 ./tests/run-all.sh ./mxreq # integration tests (live)
```

## License

See [LICENSE](LICENSE) for details.
