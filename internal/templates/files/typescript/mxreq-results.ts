/**
 * Standalone mxreq execution result recorder for TypeScript/Playwright.
 * Copy this file into your test repository.
 *
 * Usage:
 *   import { test } from '@playwright/test';
 *   import { startTest, endTest, verifyEqual } from './mxreq-results';
 *
 *   test('example', async ({ page }) => {
 *       startTest('test_example');  // Must provide test name explicitly
 *       // Note: Auto-detection not supported in TypeScript/JavaScript due to:
 *       // - Minification removes function names
 *       // - Anonymous functions have no names
 *       // - Stack trace parsing is non-standard and brittle
 *
 *       // Staging step - not recorded (no requirement)
 *       await page.goto('http://localhost:3000');
 *
 *       // Verification step - recorded
 *       const result = await page.textContent('#result');
 *       verifyEqual('SOFT-123', result, '10');
 *
 *       endTest();  // Automatically writes results to file
 *   });
 *
 * Results are written incrementally - each endTest() writes a valid YAML file.
 */

import * as fs from "fs";

/**
 * Default output filename with date (generated once at module load)
 * Uses date only (YYYYMMDD) to ensure all tests in a run write to the same file
 */
const DEFAULT_OUTPUT_FILE = (() => {
  const now = new Date();
  const date = now.toISOString().split("T")[0].replace(/-/g, "");
  return `results_${date}.yaml`;
})();

interface StepResult {
  actual: string;
  status: "PASS" | "FAIL";
  requirement?: string;
  line?: number;
}

interface TestResult {
  test_name: string;
  result: "PASS" | "FAIL";
  steps: StepResult[];
}

class ResultRecorder {
  private currentTest: TestResult | null = null;
  private tester: string = "automation";
  private version: string | null = null;
  private outputFile: string | null = null;

  configure(
    tester: string = "automation",
    version: string | null = null,
    outputFile: string | null = null
  ): void {
    this.tester = tester;
    this.version = version;
    if (outputFile) {
      this.outputFile = outputFile;
    }
  }

  startTest(testName: string): void {
    // Extract only the TC-{number} portion from the test name
    const tcMatch = testName.match(/TC[-_]\d+/i);
    // Normalize to TC-{number} format (uppercase, dash separator)
    const extractedName = tcMatch
      ? tcMatch[0].replace('_', '-').toUpperCase()
      : testName;
    this.currentTest = { test_name: extractedName, result: "PASS", steps: [] };
  }

  recordStep(
    requirement: string | null,
    actual: string,
    status: "PASS" | "FAIL"
  ): void {
    if (!requirement) return; // Skip staging steps

    if (!this.currentTest) {
      throw new Error("No active test - call startTest() first");
    }

    // Get caller's line number from stack trace
    const stack = new Error().stack || "";
    const lineMatch = stack.split("\n")[3]?.match(/:(\d+):\d+\)?$/);
    const lineNumber = lineMatch ? parseInt(lineMatch[1], 10) : 0;

    const step: StepResult = { actual, status, line: lineNumber };
    if (requirement) {
      step.requirement = requirement;
    }
    this.currentTest.steps.push(step);

    // Update overall test result if any step fails
    if (status === "FAIL") {
      this.currentTest.result = "FAIL";
    }
  }

  endTest(outputFile?: string): void {
    if (!this.currentTest || this.currentTest.steps.length === 0) {
      this.currentTest = null;
      return;
    }

    // Determine output file
    const fileToUse = outputFile || this.outputFile || DEFAULT_OUTPUT_FILE;

    // Step 1: Write results to file (always happens first)
    this.writeTestToFile(fileToUse, this.currentTest);

    // Step 2: After writing results, fail the test if any requirements failed
    if (this.currentTest.result === "FAIL") {
      const failedReqs = this.currentTest.steps
        .filter(s => s.status === "FAIL" && s.requirement)
        .map(s => `${s.requirement} (line ${s.line || "?"})`);
      this.currentTest = null;
      throw new Error(`Failed requirements: ${failedReqs.join(", ")}`);
    }

    this.currentTest = null;
  }

  private writeTestToFile(outputFile: string, test: TestResult): void {
    if (fs.existsSync(outputFile)) {
      // File exists - read, remove trailing ---, append test, write back
      let content = fs.readFileSync(outputFile, "utf-8");
      if (content.endsWith("---\n")) {
        content = content.slice(0, -4);
      }

      const testEntry = this.formatTestEntry(test);
      fs.writeFileSync(outputFile, content + testEntry + "---\n");
    } else {
      // File doesn't exist - create with header
      const today = new Date().toISOString().split("T")[0];
      let yaml = "---\n";
      yaml += `execution_date: '${today}'\n`;
      yaml += `tester: ${this.tester}\n`;
      if (this.version) {
        yaml += `sut_version: ${this.version}\n`;
      }
      yaml += "results:\n";
      yaml += this.formatTestEntry(test);
      yaml += "---\n";

      fs.writeFileSync(outputFile, yaml);
    }
  }

  private formatTestEntry(test: TestResult): string {
    let entry = `- test_name: ${test.test_name}\n`;
    entry += `  result: ${test.result}\n`;
    entry += "  steps:\n";
    for (const step of test.steps) {
      entry += `  - actual: ${JSON.stringify(step.actual)}\n`;
      entry += `    status: ${step.status}\n`;
      if (step.requirement) {
        entry += `    requirement: ${step.requirement}\n`;
      }
    }
    return entry;
  }

  clear(): void {
    this.currentTest = null;
  }
}

const globalRecorder = new ResultRecorder();

// Public API

/**
 * Configure global metadata for all tests.
 */
export function configure(
  tester: string = "automation",
  version: string | null = null,
  outputFile: string | null = null
): void {
  globalRecorder.configure(tester, version, outputFile);
}

/**
 * Begin recording a test.
 */
export function startTest(testName: string): void {
  globalRecorder.startTest(testName);
}

/**
 * Finish recording current test and write results to file.
 * Results are written incrementally - each call writes a valid YAML file.
 * Failed requirements are logged but don't cause test failure.
 */
export function endTest(outputFile?: string): void {
  globalRecorder.endTest(outputFile);
}

/**
 * Clear all recorded results.
 */
export function clearResults(): void {
  globalRecorder.clear();
}

// Verification helpers

/**
 * Internal helper to verify a test step and record the result.
 * Only steps with a non-null requirement are recorded.
 *
 * @param requirement - Requirement link (e.g., "SOFT-123"). Null for staging steps.
 * @param condition - Boolean assertion result.
 * @param actual - Description of actual result.
 * @param fatal - If true, throw Error on failure.
 * @returns The condition value (true/false).
 */
function verifyStep(
  requirement: string | null,
  condition: boolean,
  actual: string,
  fatal: boolean = false
): boolean {
  const status = condition ? "PASS" : "FAIL";
  globalRecorder.recordStep(requirement, actual, status);

  if (!condition && fatal) {
    throw new Error(`Fatal verification failed: ${actual}`);
  }

  return condition;
}

/**
 * Verify two values are equal.
 */
export function verifyEqual<T>(
  requirement: string | null,
  actualValue: T,
  expectedValue: T,
  fatal: boolean = false
): boolean {
  const condition = actualValue === expectedValue;
  const actual = `Expected ${expectedValue}, got ${actualValue}`;
  return verifyStep(requirement, condition, actual, fatal);
}

/**
 * Verify value is true.
 */
export function verifyTrue(
  requirement: string | null,
  value: boolean,
  message: string = "",
  fatal: boolean = false
): boolean {
  const actual = message || `Value is ${value}`;
  return verifyStep(requirement, value === true, actual, fatal);
}

/**
 * Verify value is false.
 */
export function verifyFalse(
  requirement: string | null,
  value: boolean,
  message: string = "",
  fatal: boolean = false
): boolean {
  const actual = message || `Value is ${value}`;
  return verifyStep(requirement, value === false, actual, fatal);
}

/**
 * Verify two values are not equal.
 */
export function verifyNotEqual<T>(
  requirement: string | null,
  actualValue: T,
  expectedValue: T,
  fatal: boolean = false
): boolean {
  const condition = actualValue !== expectedValue;
  const actual = `Expected not ${expectedValue}, got ${actualValue}`;
  return verifyStep(requirement, condition, actual, fatal);
}
