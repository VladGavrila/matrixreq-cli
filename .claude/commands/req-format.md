---
description: Convert requirements into MatrixALM YAML definition format
argument-hint: <requirements text or description>
---

# Generate MatrixALM Requirements YAML

Convert the provided requirements into a YAML definition file compatible with `mxreq sync`.

## Input

The user will provide one or more requirements as the argument: $ARGUMENTS

## Step 1: Determine Target Folder

Before generating the YAML, ask the user where to store the requirements. Present the choice using AskUserQuestion:

- **Question:** "Which folder should these requirements be stored in?"
- **Options:**
  1. `MXREQ` (default) — a dedicated folder for AI-generated requirements
  2. Let the user type a custom folder name

Once the user picks a folder name (e.g. "MXREQ" or a custom one), you need to check if a folder with that label already exists under the target category:

1. Run `mxreq project tree -p <project> --filter <CATEGORY> -o json` to get the project tree for the relevant category (e.g. REQ, SPEC).
2. Search the tree output for a folder whose title/label matches the chosen name.
3. **If found:** use its folder reference (e.g. `F-REQ-42`) for all items.
4. **If NOT found:** create it with `mxreq folder create -p <project> --parent F-<CATEGORY>-1 --label "<folder name>" -r "Auto-created folder for requirements"` and use the newly created folder reference.

Use the project from the `--project`/`-p` flag or `default_project` config. If no project is available, ask the user.

## Step 2: Generate YAML

## Output Format

Output a valid YAML file using this exact structure:

```yaml
items:
  - title: "Short component/topic identifier"
    folder: F-<CATEGORY>-<ID>
    fields:
      Description: "<p>The full imperative requirement statement (shall/must language)</p>"
    labels:
      - Draft
    up_links: "PARENT-REF-1, PARENT-REF-2"
```

## Field Reference

Each item under `items:` supports these fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | **Yes** | Short identifier for the component or topic (e.g. "SSO Authentication", "Login Audit Logging") |
| `item_ref` | string | No | Existing item ID (e.g. `REQ-42`) — include ONLY when updating an existing item. Omit entirely for new items. |
| `folder` | string | No | Target folder reference (e.g. `F-REQ-1`, `F-SPEC-10`). Format: `F-<CATEGORY>-<NUMBER>` |
| `fields` | map | No | Field label → value pairs matching the target category's fields |
| `labels` | list | No | String tags (e.g. `Draft`, `Security`, `High Priority`) |
| `up_links` | string | No | Comma-separated parent item references (e.g. `"SPEC-100, SPEC-101"`) |

## Rules

1. **title** is always required — use it as a short identifier for the component or topic (e.g. "SSO Authentication", "Input Validation", "Login Audit Logging"). NOT the full requirement statement.
2. **Description field** contains the full imperative requirement statement using shall/must language (e.g. "The system shall authenticate users via SSO"). This is where the actual requirement text goes.
3. **Do NOT include `item_ref`** unless the user explicitly says they are updating existing items and provides the item IDs
4. **fields** keys are arbitrary strings that must match the category's configured field labels in Matrix. Common ones:
   - `Description` — The full imperative requirement statement in HTML (wrap in `<p>` tags)
   - `Rationale` — Why this requirement exists
   - `Priority` — e.g. High, Medium, Low
   - `Status` — e.g. Draft, Approved, In Progress
5. **up_links** is a single comma-separated string, NOT a list
6. **labels** is a YAML list of strings
7. **folder** must use the folder reference resolved in Step 1 — the same value for all items in a batch
8. If the user specifies a category (REQ, SPEC, RISK, etc.), use it consistently; otherwise default to REQ
9. If the user doesn't specify labels, default to `Draft`
10. Field values support HTML — use `<p>`, `<ul>`, `<li>`, `<ol>` for rich formatting when the requirement has multiple clauses or details
11. Output ONLY the YAML content inside a yaml code block — no extra commentary before or after

## Example

For input: "The system must authenticate users via SSO and must log all failed login attempts"

1. Ask user for folder name → user picks "MXREQ" (default)
2. Run `mxreq project tree -p PROJ --filter REQ -o json` → no "MXREQ" folder found
3. Run `mxreq folder create -p PROJ --parent F-REQ-1 --label "MXREQ" -r "Auto-created folder for requirements"` → Created folder ID=42
4. Use `F-REQ-42` as folder for all items
5. Output:

```yaml
items:
  - title: "SSO Authentication"
    folder: F-REQ-42
    fields:
      Description: "<p>The system shall authenticate users via Single Sign-On (SSO) for all login flows.</p>"
    labels:
      - Draft
      - Security

  - title: "Failed Login Audit Logging"
    folder: F-REQ-42
    fields:
      Description: "<p>The system shall log all failed login attempts including timestamp, username, source IP, and failure reason for security audit purposes.</p>"
    labels:
      - Draft
      - Security
```
