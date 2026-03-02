/**
 * mxreq YAML Docstring Generator for TypeScript/Playwright.
 * Copy this file into your test repository and import it.
 *
 * Generates YAML-formatted docstrings and inserts them directly into source files.
 * Use the DS=1 environment variable to enable docstring generation mode.
 */

import * as fs from "fs";

/** Environment toggle - true when DS=1 is set */
export const ds = process.env.DS === "1";

interface TestStep {
  action: string;
  expected: string;
  RequirementLink: string;
}

interface DocstringState {
  title: string;
  description: string;
  folder: string;
  assumptions: string[];
  steps: TestStep[];
  labels: string[];
  upLinks: Set<string>;
  sourceFile: string;
  testName: string;
}

let currentBuilder: DocstringState | null = null;

interface DsHeaderOptions {
  /** Test case title as it will appear in Matrix */
  title: string;
  /** Description of the test */
  description?: string;
  /** Parent folder reference (e.g., "F-TC-521") */
  folder?: string;
  /** List of assumption/precondition strings */
  assumptions?: string[];
}

/**
 * Start building a test case docstring.
 *
 * @param opts - Header options containing title, description, folder, and assumptions
 */
export function dsHeader(opts: DsHeaderOptions): void {
  if (!ds) return;

  // Get caller info from Error stack
  const stack = new Error().stack || "";
  const stackLines = stack.split("\n");

  // Find the source file and line number (skip first 2 lines: Error and dsHeader itself)
  let sourceFile = "";
  let lineNumber = 0;

  // Look through stack lines to find the test file (not this mxreq-docstring file)
  // Node.js/Playwright stack formats: "at filepath:line:col" or "at func (filepath:line:col)"
  for (let i = 2; i < stackLines.length; i++) {
    const line = stackLines[i];
    // Try format with parentheses first: at func (filepath:line:col)
    let match = line.match(/\((.+\.(?:ts|js)):(\d+):\d+\)/);
    // If no match, try format without parentheses: at filepath:line:col
    if (!match) {
      match = line.match(/at\s+(.+\.(?:ts|js)):(\d+):\d+/);
    }
    if (match && !match[1].includes("mxreq-docstring")) {
      sourceFile = match[1];
      lineNumber = parseInt(match[2], 10);
      break;
    }
  }

  // Extract test name from source file
  let testName = "";
  if (sourceFile && lineNumber > 0) {
    try {
      const content = fs.readFileSync(sourceFile, "utf-8");
      const lines = content.split("\n");

      // Search backwards from the call line to find the test() declaration
      // Search up to 50 lines back to account for large docstrings
      for (let i = lineNumber - 1; i >= Math.max(0, lineNumber - 50); i--) {
        const testMatch = lines[i].match(/(?:test|it)\s*\(\s*['"`]([^'"`]+)['"`]/);
        if (testMatch) {
          testName = testMatch[1];
          break;
        }
      }
    } catch {
      // If we can't read the file, testName remains empty
    }
  }

  currentBuilder = {
    title: opts.title,
    description: opts.description || "",
    folder: opts.folder || "",
    assumptions: opts.assumptions || [],
    steps: [],
    labels: [],
    upLinks: new Set(),
    sourceFile,
    testName,
  };
}

interface DsStepOptions {
  /** Expected result. Defaults to "N/A" for setup/cleanup steps */
  expected?: string;
  /** Requirement link(s), comma-separated (e.g., "SOFT-387") */
  req?: string;
}

/**
 * Add a test step to the docstring.
 *
 * @param action - Description of the test action
 * @param opts - Optional step options with expected result and requirement link
 */
export function dsStep(action: string, opts?: DsStepOptions): void {
  if (!ds || !currentBuilder) return;

  const expected = opts?.expected || "N/A";
  const req = opts?.req || "";

  currentBuilder.steps.push({
    action,
    expected,
    RequirementLink: req,
  });

  // Collect requirements for up_links
  if (req) {
    req.split(",").forEach((r) => currentBuilder!.upLinks.add(r.trim()));
  }
}

/**
 * Convenience wrapper around dsStep that prepends '<b>Manual:</b> ' to the action.
 *
 * Equivalent to: dsStep(`<b>Manual:</b> ${action}`, opts)
 *
 * @param action - Description of the test action
 * @param opts - Optional step options with expected result and requirement link
 */
export function dsManual(action: string, opts?: DsStepOptions): void {
  dsStep(`<b>Manual:</b> ${action}`, opts);
}

/**
 * Convenience wrapper around dsStep that prepends '<b><i>USER INTERVENTION</b></i> ' to the action.
 *
 * Equivalent to: dsStep(`<b><i>USER INTERVENTION</b></i> ${action}`, opts)
 *
 * @param action - Description of the test action
 * @param opts - Optional step options with expected result and requirement link
 */
export function dsUser(action: string, opts?: DsStepOptions): void {
  dsStep(`<b><i>USER INTERVENTION</b></i> ${action}`, opts);
}

interface DsFooterOptions {
  /** List of label strings (e.g., ["Automated", "VM"]) */
  labels?: string[];
}

/**
 * Finish building and insert the docstring into the source file.
 *
 * @param opts - Optional footer options with labels
 */
export function dsFooter(opts?: DsFooterOptions): void {
  if (!ds || !currentBuilder) return;

  if (opts?.labels) {
    currentBuilder.labels = opts.labels;
  }

  const yaml = generateYAML();
  insertDocstring(yaml);
  currentBuilder = null;
}

function generateYAML(): string {
  const b = currentBuilder!;
  const lines: string[] = [];

  lines.push("/*");
  lines.push("---");

  // Title
  lines.push(`title: "${b.title}"`);

  // Folder
  if (b.folder) {
    lines.push(`folder: ${b.folder}`);
  }

  // Description (literal block scalar)
  if (b.description) {
    lines.push("description: |");
    for (const line of b.description.split("\n")) {
      lines.push(`  ${line}`);
    }
  }

  // Assumptions (formatted as HTML list)
  if (b.assumptions.length > 0) {
    lines.push("assumptions: |");
    lines.push("  <ul>");
    for (const a of b.assumptions) {
      // Strip leading "* " if present and wrap in <li> tags
      const text = a.startsWith("* ") ? a.substring(2) : a;
      lines.push(`   <li>${text}</li>`);
    }
    lines.push("  </ul>");
  }

  // Steps
  lines.push("steps:");
  for (const step of b.steps) {
    lines.push(`  - action: "${step.action}"`);
    if (step.expected !== "N/A") {
      lines.push(`    expected: "${step.expected}"`);
    }
    if (step.RequirementLink) {
      lines.push(`    RequirementLink: "${step.RequirementLink}"`);
    }
  }

  // Labels
  if (b.labels.length > 0) {
    lines.push("labels:");
    for (const l of b.labels) {
      lines.push(`  - ${l}`);
    }
  }

  // Up links (collected from steps)
  if (b.upLinks.size > 0) {
    const links = Array.from(b.upLinks).sort().join(", ");
    lines.push(`up_links: "${links}"`);
  }

  lines.push("---");
  lines.push("*/");

  return lines.join("\n");
}

function escapeRegex(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function insertDocstring(yaml: string): void {
  const b = currentBuilder!;

  let content: string;
  try {
    content = fs.readFileSync(b.sourceFile, "utf-8");
  } catch (err) {
    console.log(`Error reading file ${b.sourceFile}: ${err}`);
    return;
  }

  // Pattern to find the test definition and optional existing docstring
  // Handles: test('name', ...) or it('name', ...) with either /* or /** style comments
  const testPattern = new RegExp(
    `((?:test|it)\\s*\\(\\s*['"\`]${escapeRegex(b.testName)}['"\`]\\s*,\\s*(?:async\\s*)?\\([^)]*\\)\\s*=>\\s*\\{)\\s*\\n(\\s*/\\*\\*?[\\s\\S]*?\\*/\\s*\\n)?`,
    "g"
  );

  // Indent the YAML with 2 spaces (TypeScript convention)
  const indentedYAML = yaml
    .split("\n")
    .map((line) => "  " + line)
    .join("\n");
  const replacement = `$1\n${indentedYAML}\n`;

  const newContent = content.replace(testPattern, replacement);

  if (newContent === content) {
    // Check if the test exists but the docstring is already up-to-date
    if (testPattern.test(content)) {
      console.log(`Docstring for '${b.testName}' is already up-to-date in ${b.sourceFile}`);
      return;
    }
    console.log(
      `Could not find test '${b.testName}' in ${b.sourceFile}`
    );
    return;
  }

  try {
    fs.writeFileSync(b.sourceFile, newContent);
  } catch (err) {
    console.log(`Error writing file ${b.sourceFile}: ${err}`);
    return;
  }

  console.log(`Generated docstring for '${b.testName}' in ${b.sourceFile}`);
}
