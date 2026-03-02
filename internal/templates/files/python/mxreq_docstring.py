"""
mxreq YAML Docstring Generator.
Copy this file into your test repository and import it.

Generates YAML-formatted docstrings and inserts them directly into source files.
Use the DS=1 environment variable to enable docstring generation mode.

Usage:
    # Generate docstrings (test code is skipped):
    DS=1 python -m pytest tests/test_example.py -s

    # Normal test execution (ds_* calls are no-ops):
    python -m pytest tests/test_example.py

Example:
    from mxreq_docstring import ds, ds_header, ds_step, ds_manual, ds_user, ds_footer
    from mxreq_results import verify_equal, start_test, end_test

    def test_NEW_short_test_name():
        ds_header(
            title="Long Test Name",
            description="Test description",
            folder="F-TC-123",
            assumptions=["* Assumption 1", "* Assumption 2"]
        )

        start_test()

        # Step 1: Setup
        ds_step("Description of step")
        if not ds:
            run = setup("arg1", "arg2")
            execute(run)

        # Step 2: User intervention required
        ds_user("User will perform an action", expected="Action has been performed", req="SOFT-123")

        # Step 3: Verify (automated)
        ds_step("Verify action is as per TSPEC-82", expected="Action is as per TSPEC", req="SOFT-124")
        if not ds:
            verify_equal("SOFT-124", action(), "argToCompare")

        # Step 4: Manual verification
        ds_manual("Visually inspect the display for correct formatting", expected="Time is displayed in MM:SS format", req="SOFT-125")

        # Step 5: Cleanup
        ds_step("Discard process")
        if not ds:
            discard()

        ds_footer(labels=["Automated", "VM"])
        if not ds: end_test()
"""

import os
import re
import inspect
from typing import Optional
from threading import Lock


# Environment toggle - True when DS=1 is set
ds = os.getenv("DS") == "1"


class _DocstringBuilder:
    """Thread-safe singleton builder for YAML docstrings."""

    def __init__(self):
        self._lock = Lock()
        self._current: Optional[dict] = None

    def start(
        self,
        title: str,
        description: str = "",
        folder: str = "",
        assumptions: Optional[list[str]] = None,
    ):
        """Start building a new docstring."""
        with self._lock:
            self._current = {
                "title": title,
                "description": description,
                "folder": folder,
                "assumptions": assumptions or [],
                "steps": [],
                "labels": [],
                "up_links": set(),
                "source_file": None,
                "function_name": None,
            }
            # Auto-detect source file and function from caller
            # Go up 2 frames: start() -> ds_header() -> test function
            frame = inspect.currentframe()
            if frame and frame.f_back and frame.f_back.f_back:
                caller = frame.f_back.f_back
                self._current["source_file"] = caller.f_code.co_filename
                self._current["function_name"] = caller.f_code.co_name

    def add_step(self, action: str, expected: str = "N/A", req: str = ""):
        """Add a test step to the current docstring."""
        with self._lock:
            if not self._current:
                raise RuntimeError("Call ds_header() before ds_step()")
            self._current["steps"].append(
                {
                    "action": action,
                    "expected": expected,
                    "RequirementLink": req,
                }
            )
            # Collect requirements for up_links
            if req:
                for r in req.split(","):
                    self._current["up_links"].add(r.strip())

    def finish(self, labels: Optional[list[str]] = None):
        """Finish building and insert the docstring into the source file."""
        with self._lock:
            if not self._current:
                raise RuntimeError("Call ds_header() before ds_footer()")
            if labels:
                self._current["labels"] = labels

            yaml_str = self._generate_yaml()
            self._insert_docstring(yaml_str)
            self._current = None

    def _generate_yaml(self) -> str:
        """Generate YAML content for the docstring."""
        c = self._current
        lines = ['"""', "---"]

        # Title
        lines.append(f'title: "{c["title"]}"')

        # Folder
        if c["folder"]:
            lines.append(f"folder: {c['folder']}")

        # Description (literal block scalar)
        if c["description"]:
            lines.append("description: |")
            for line in c["description"].split("\n"):
                lines.append(f"  {line}")

        # Assumptions (formatted as HTML list)
        if c["assumptions"]:
            lines.append("assumptions: |")
            lines.append("  <ul>")
            for assumption in c["assumptions"]:
                # Strip leading "* " if present and wrap in <li> tags
                text = assumption[2:] if assumption.startswith("* ") else assumption
                lines.append(f"   <li>{text}</li>")
            lines.append("  </ul>")

        # Steps
        lines.append("steps:")
        for step in c["steps"]:
            lines.append(f'  - action: "{step["action"]}"')
            if step["expected"] != "N/A":
                lines.append(f'    expected: "{step["expected"]}"')
            if step["RequirementLink"]:
                lines.append(f'    RequirementLink: "{step["RequirementLink"]}"')

        # Labels
        if c["labels"]:
            lines.append("labels:")
            for label in c["labels"]:
                lines.append(f"  - {label}")

        # Up links (collected from steps)
        if c["up_links"]:
            links = ", ".join(sorted(c["up_links"]))
            lines.append(f'up_links: "{links}"')

        lines.append("---")
        lines.append('"""')

        return "\n".join(lines)

    def _insert_docstring(self, yaml_str: str):
        """Insert or replace the docstring in the source file."""
        source_file = self._current["source_file"]
        func_name = self._current["function_name"]

        with open(source_file, "r") as f:
            content = f.read()

        # Pattern to find function definition and optional existing docstring
        # Matches: def func_name(...): followed by optional docstring
        pattern = rf'(def {re.escape(func_name)}\([^)]*\):)\s*\n(\s*"""[\s\S]*?"""\s*\n)?'

        def replacer(match):
            func_def = match.group(1)
            # Indent the YAML docstring with 4 spaces
            indent = "    "
            indented_yaml = "\n".join(indent + line for line in yaml_str.split("\n"))
            return f"{func_def}\n{indented_yaml}\n"

        new_content, count = re.subn(pattern, replacer, content)

        if count == 0:
            print(f"Could not find function '{func_name}' in {source_file}")
            return

        with open(source_file, "w") as f:
            f.write(new_content)

        print(f"Generated docstring for {func_name} in {source_file}")


# Global singleton
_builder = _DocstringBuilder()


# Public API
def ds_header(
    title: str,
    description: str = "",
    folder: str = "",
    assumptions: Optional[list[str]] = None,
):
    """Start building a test case docstring.

    Args:
        title: Test case title as it will appear in Matrix.
        description: HTML-formatted description of the test.
        folder: Parent folder reference (e.g., "F-TC-521").
        assumptions: List of assumption/precondition strings.
    """
    if ds:
        _builder.start(title, description, folder, assumptions)


def ds_step(action: str, expected: str = "N/A", req: str = ""):
    """Add a test step to the docstring.

    Args:
        action: Description of the test action.
        expected: Expected result. Defaults to "N/A" for setup/cleanup steps.
        req: Requirement link(s), comma-separated (e.g., "SOFT-387" or "SOFT-387,SOFT-388").
    """
    if ds:
        _builder.add_step(action, expected, req)


def ds_manual(action: str, expected: str = "N/A", req: str = ""):
    """Convenience wrapper around ds_step that prepends '<b>Manual:</b> ' to the action.

    Equivalent to: ds_step(f"<b>Manual:</b> {action}", expected, req)

    Args:
        action: Description of the test action.
        expected: Expected result. Defaults to "N/A" for setup/cleanup steps.
        req: Requirement link(s), comma-separated (e.g., "SOFT-387" or "SOFT-387,SOFT-388").
    """
    ds_step(f"<b>Manual:</b> {action}", expected, req)


def ds_user(action: str, expected: str = "N/A", req: str = ""):
    """Convenience wrapper around ds_step that prepends '<b><i>USER INTERVENTION</b></i> ' to the action.

    Equivalent to: ds_step(f"<b><i>USER INTERVENTION</b></i> {action}", expected, req)

    Args:
        action: Description of the test action.
        expected: Expected result. Defaults to "N/A" for setup/cleanup steps.
        req: Requirement link(s), comma-separated (e.g., "SOFT-387" or "SOFT-387,SOFT-388").
    """
    ds_step(f"<b><i>USER INTERVENTION</b></i> {action}", expected, req)


def ds_footer(labels: Optional[list[str]] = None):
    """Finish building and insert the docstring into the source file.

    Args:
        labels: List of label strings (e.g., ["Automated", "VM"]).
    """
    if ds:
        _builder.finish(labels)
