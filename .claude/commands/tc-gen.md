---
description: Generate MatrixALM test cases in sync-ready YAML from requirements
argument-hint: [requirements-yaml-file or requirement refs]
allowed-tools: Read, Write
model: claude-sonnet-4-6
---

# Generate MatrixALM Test Cases YAML

Generate test cases from the provided requirements and output a YAML file compatible with `mxreq sync`.

## Input

The user provides either:
- A path to a requirements YAML file (e.g. `RaC/requirements.yaml`): $ARGUMENTS
- A list of requirement refs (e.g. `SOFT-3454, SOFT-3455, SOFT-3456`)
- A plain-text description of the behaviour to test

## Step 1: Read Requirements

If given a YAML file path, read it and extract the relevant items (title, item_ref, Description field).
If given refs or plain text, use them directly as the source material.

## Step 2: Determine Target Folder Ref

Ask the user which `folder` ref to use for all generated test cases:

- **Question:** "Which F-TC folder ref should these test cases be stored in? (Press Enter to use `F-TC-UPDATE`)"
- If the user provides a ref (e.g. `F-TC-186`), use it for all items.
- If the user skips or does not provide one, use `F-TC-UPDATE` as the default.

Do **not** call any `mxreq` commands to look up or create folders.

## Step 3: Group Requirements into Test Cases

Analyse the requirements and group related ones into coherent test cases. Follow these rules:

- **Prefer CRUD groupings:** gather Create + Read + Update + Delete operations for the same resource into a single test case.
- **Order steps to follow the operation lifecycle:** Create → Read/List → Update → Delete.
- **Each test case must cover 3–6 requirements** (link every step to the requirement it exercises via `RequirementLink`).
- **Each test case must have at least 5 steps.**
- **Step actions must be concrete CLI commands** using the `mxreq` binary (e.g. `mxreq user create --login testuser ...`).
- If requirements do not form clean CRUD groups, group by feature area and order steps logically.

## Step 4: Write YAML File

Write the generated test cases to **`TaC/test-cases.yaml`** (create the `TaC/` directory if it doesn't exist). Use this exact structure:

```yaml
items:
  - title: "Short descriptive name for the test case"
    folder: F-TC-<ID>
    fields:
      Description: "<p>What this test case verifies, in one or two sentences.</p>"
    steps:
      - action: "Run: mxreq <command> <flags>"
        expected: "Command exits with code 0 and output confirms <observable result>."
        RequirementLink: "SOFT-XXXX"
      - action: "Run: mxreq <command> <flags>"
        expected: "<Observable result that can be checked by the tester.>"
        RequirementLink: "SOFT-XXXX"
    labels:
      - Draft
    up_links: "SOFT-XXXX, SOFT-YYYY, SOFT-ZZZZ"
```

## Field Reference

| Field | Required | Notes |
|-------|----------|-------|
| `title` | Yes | Short identifier for the test case (not the full description) |
| `item_ref` | No | Omit for new test cases; include only when updating existing TC items |
| `folder` | Yes | TC folder ref resolved in Step 2 |
| `fields.Description` | Yes | HTML string wrapped in `<p>` tags |
| `steps` | Yes | Ordered YAML list; each entry has `action`, `expected`, `RequirementLink` |
| `labels` | Yes | Default to `Draft` |
| `up_links` | Yes | Comma-separated requirement refs this test case covers |

## Rules

1. **No `item_ref`** unless updating an existing TC item.
2. **`up_links`** must list every requirement ref exercised by the test case.
3. **`RequirementLink`** in each step must be a single ref string (e.g. `"SOFT-3483"`).
4. **`expected`** must describe an observable outcome a tester can verify — not just "success".
5. **`action`** must be a concrete `mxreq` CLI invocation or a clearly described manual action.
6. **Step order must reflect the operation lifecycle** (Create first, then Read/List, then Update, then Delete).
7. **Write the YAML to `TaC/test-cases.yaml`** using the Write tool. Do not just print it — save it to the file.

## Example

For requirements covering user CRUD (SOFT-3481 List, SOFT-3482 Get, SOFT-3483 Create, SOFT-3484 Update, SOFT-3485 Delete):

```yaml
items:
  - title: "User Management CRUD"
    folder: F-TC-186
    fields:
      Description: "<p>Verify the full user lifecycle: create, list, get details, update attributes, and delete a user using the <code>mxreq user</code> command group.</p>"
    steps:
      - action: "Run: mxreq user create --login testuser --email test@example.com --password Secret123!"
        expected: "Command exits with code 0 and output confirms user 'testuser' was created."
        RequirementLink: "SOFT-3483"
      - action: "Run: mxreq user list"
        expected: "Output table contains a row with login 'testuser'."
        RequirementLink: "SOFT-3481"
      - action: "Run: mxreq user get testuser"
        expected: "Output displays ID, login, email, name, status, admin level, and token list for 'testuser'."
        RequirementLink: "SOFT-3482"
      - action: "Run: mxreq user update testuser --email updated@example.com"
        expected: "Command exits with code 0 and output confirms the email has been updated to 'updated@example.com'."
        RequirementLink: "SOFT-3484"
      - action: "Run: mxreq user delete testuser, then verify with mxreq user list"
        expected: "Delete command exits with code 0; subsequent list does not contain 'testuser'."
        RequirementLink: "SOFT-3485"
    labels:
      - Draft
    up_links: "SOFT-3481, SOFT-3482, SOFT-3483, SOFT-3484, SOFT-3485"
```
