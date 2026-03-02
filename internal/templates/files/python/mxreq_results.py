"""
Standalone mxreq execution result recorder.
Copy this file into your test repository and import it.

Usage:
    from mxreq_results import verify_equal, start_test, end_test

    def test_example():
        start_test()  # Auto-detects test name from function
        # Or explicitly: start_test("test_example")

        # Staging step - not recorded (no requirement)
        data = setup()

        # Verification step - recorded
        result = calculate(data)
        verify_equal("SOFT-123", result, 10)

        end_test()  # Automatically writes results to file

    # Results are written incrementally - each end_test() writes a valid YAML file.
"""

import json
import inspect
import os
from datetime import date, datetime
from typing import Optional, Any
from threading import Lock

# Default output filename with date (YYYYMMDD) to ensure all tests write to same file
DEFAULT_OUTPUT_FILE = f"results_{datetime.now().strftime('%Y%m%d')}.yaml"


class _ResultRecorder:
    """Thread-safe singleton recorder."""

    def __init__(self):
        self._lock = Lock()
        self._current_test = None
        self._tester = "automation"
        self._version = None
        self._output_file = None  # Configured output file

    def configure(
        self,
        tester: str = "automation",
        version: Optional[str] = None,
        output_file: Optional[str] = None,
    ):
        """Configure global metadata."""
        self._tester = tester
        self._version = version
        if output_file:
            self._output_file = output_file

    def start_test(self, test_name: Optional[str]):
        """Begin recording a test."""
        with self._lock:
            # Extract only the TC-{number} portion from the test name
            import re
            extracted_name = test_name
            if test_name:
                tc_match = re.search(r'TC[-_]\d+', test_name, re.IGNORECASE)
                if tc_match:
                    # Normalize to TC-{number} format
                    extracted_name = tc_match.group(0).replace('_', '-').upper()
            self._current_test = {"test_name": extracted_name, "result": "PASS", "steps": []}

    def record_step(self, requirement: Optional[str], actual: str, status: str):
        """Record a step result (only if requirement is provided)."""
        if not requirement:
            return  # Skip staging steps

        with self._lock:
            if not self._current_test:
                raise RuntimeError("No active test - call start_test() first")

            # Get caller's line number
            frame = inspect.currentframe()
            line_number = 0
            if frame and frame.f_back and frame.f_back.f_back:
                line_number = frame.f_back.f_back.f_lineno

            step = {"actual": actual, "status": status.upper(), "line": line_number}
            if requirement:
                step["requirement"] = requirement

            self._current_test["steps"].append(step)

            # Update overall test result if any step fails
            if status.upper() == "FAIL":
                self._current_test["result"] = "FAIL"

    def end_test(self, output_file: Optional[str] = None) -> None:
        """Finish recording current test and write to file.

        Test completes fully, results are written, then test fails if any requirements failed.
        """
        with self._lock:
            if not self._current_test or not self._current_test["steps"]:
                self._current_test = None
                return

            # Determine output file
            file_to_use = output_file or self._output_file or DEFAULT_OUTPUT_FILE

            # Step 1: Write results to file (always happens first)
            self._write_test_to_file(file_to_use, self._current_test)

            # Step 2: After writing results, fail the test if any requirements failed
            if self._current_test["result"] == "FAIL":
                failed_reqs = [f"{s['requirement']} (line {s.get('line', '?')})"
                              for s in self._current_test["steps"]
                              if s["status"] == "FAIL" and s.get("requirement")]
                self._current_test = None
                raise AssertionError(f"Failed requirements: {', '.join(failed_reqs)}")

            self._current_test = None

    def _write_test_to_file(self, output_file: str, test: dict):
        """Write a single test to file, appending if file exists."""
        if os.path.exists(output_file):
            # Read existing content, remove trailing ---
            with open(output_file, "r") as f:
                content = f.read()
            if content.endswith("---\n"):
                content = content[:-4]
            # Append new test and closing ---
            with open(output_file, "w") as f:
                f.write(content)
                self._write_test_entry(f, test)
                f.write("---\n")
        else:
            # Create new file with header
            with open(output_file, "w") as f:
                f.write("---\n")
                f.write(f"execution_date: '{date.today().isoformat()}'\n")
                f.write(f"tester: {self._tester}\n")
                if self._version:
                    f.write(f"sut_version: {self._version}\n")
                f.write("results:\n")
                self._write_test_entry(f, test)
                f.write("---\n")

    def _write_test_entry(self, f, test: dict):
        """Write a single test entry to file handle."""
        f.write(f"- test_name: {test['test_name']}\n")
        f.write(f"  result: {test['result']}\n")
        f.write("  steps:\n")
        for step in test["steps"]:
            f.write(f"  - actual: {json.dumps(step['actual'])}\n")
            f.write(f"    status: {step['status']}\n")
            if "requirement" in step:
                f.write(f"    requirement: {step['requirement']}\n")

    def clear(self):
        """Clear all recorded results."""
        with self._lock:
            self._current_test = None


# Global singleton
_recorder = _ResultRecorder()


# Public API
def configure(
    tester: str = "automation",
    version: Optional[str] = None,
    output_file: Optional[str] = None,
):
    """Configure global metadata for all tests.

    Args:
        tester: Name of the tester (default: "automation")
        version: Version of the software under test
        output_file: Output filename for results (default: results_TIMESTAMP.yaml)
    """
    _recorder.configure(tester, version, output_file)


def start_test(test_name: Optional[str] = None):
    """Begin recording a test.

    Args:
        test_name: Test name to record. If None, auto-detects from calling function.
    """
    if test_name is None:
        # Auto-detect from caller's frame
        frame = inspect.currentframe()
        if frame and frame.f_back:
            test_name = frame.f_back.f_code.co_name
        else:
            raise RuntimeError("Could not auto-detect test name")
    _recorder.start_test(test_name)


def end_test(output_file: Optional[str] = None) -> None:
    """Finish recording current test and write results to file.

    Results are written incrementally - each call writes a valid YAML file.
    If the file exists, the closing '---' is removed, the new test is appended,
    and the closing '---' is added back.

    Args:
        output_file: Optional output filename. If None, uses configured or default file.

    Note:
        Failed requirements are logged but don't cause test failure.
    """
    _recorder.end_test(output_file)


def clear_results():
    """Clear all recorded results."""
    _recorder.clear()


# Verification helpers
def _verify_step(
    requirement: Optional[str], condition: bool, actual: str, fatal: bool = False
) -> bool:
    """Internal helper to verify a test step and record the result.

    Args:
        requirement: Requirement link (e.g., "SOFT-123"). None for staging steps.
        condition: Boolean assertion result.
        actual: Description of actual result.
        fatal: If True, raise AssertionError on failure.

    Returns:
        The condition value (True/False).
    """
    status = "PASS" if condition else "FAIL"
    _recorder.record_step(requirement, actual, status)

    if not condition and fatal:
        raise AssertionError(f"Fatal verification failed: {actual}")

    return condition


def verify_equal(
    requirement: Optional[str],
    actual_value: Any,
    expected_value: Any,
    fatal: bool = False,
) -> bool:
    """Verify two values are equal."""
    condition = actual_value == expected_value
    actual_msg = f"Expected {expected_value}, got {actual_value}"
    return _verify_step(requirement, condition, actual_msg, fatal)


def verify_true(
    requirement: Optional[str], value: bool, message: str = "", fatal: bool = False
) -> bool:
    """Verify value is True."""
    actual_msg = message or f"Value is {value}"
    return _verify_step(requirement, value is True, actual_msg, fatal)


def verify_false(
    requirement: Optional[str], value: bool, message: str = "", fatal: bool = False
) -> bool:
    """Verify value is False."""
    actual_msg = message or f"Value is {value}"
    return _verify_step(requirement, value is False, actual_msg, fatal)


def verify_not_equal(
    requirement: Optional[str],
    actual_value: Any,
    expected_value: Any,
    fatal: bool = False,
) -> bool:
    """Verify two values are not equal."""
    condition = actual_value != expected_value
    actual_msg = f"Expected not {expected_value}, got {actual_value}"
    return _verify_step(requirement, condition, actual_msg, fatal)


# Pytest integration (optional - only if pytest is available and used as a plugin)
try:
    import pytest

    def pytest_addoption(parser):
        """Add mxreq command-line options."""
        group = parser.getgroup("mxreq", "mxreq execution result generation")
        group.addoption("--mxreq-output", action="store", default=None)
        group.addoption("--mxreq-tester", action="store", default="automation")
        group.addoption("--mxreq-version", action="store", default=None)

    def pytest_configure(config):
        """Enable mxreq if --mxreq-output is provided."""
        output = config.getoption("--mxreq-output", None)
        if output:
            config._mxreq_enabled = True
            config._mxreq_output = output

            tester = config.getoption("--mxreq-tester", "automation")
            version = config.getoption("--mxreq-version", None)
            # Configure with output file so end_test() knows where to write
            configure(tester, version, output)

    def pytest_runtest_setup(item):
        """Auto-start test recording."""
        if hasattr(item.config, "_mxreq_enabled"):
            start_test(item.name)

    def pytest_runtest_teardown(item):
        """Auto-end test recording (writes results incrementally)."""
        if hasattr(item.config, "_mxreq_enabled"):
            end_test()  # Writes to file immediately with closing ---

    def pytest_sessionfinish(session):
        """Session finish hook - results already written by end_test()."""
        pass  # No action needed, each end_test() writes valid YAML

except ImportError:
    # Pytest not available - manual mode only
    pass
