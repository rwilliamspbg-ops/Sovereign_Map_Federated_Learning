#!/usr/bin/env python3
"""Write latest-vs-previous validation diff to GitHub step summary."""

from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[3]
OUT_DIR = REPO_ROOT / "test-results" / "full-validation"


def load_json_artifacts() -> list[tuple[Path, dict[str, Any]]]:
    artifacts = sorted(
        OUT_DIR.glob("full_validation_*.json"), key=lambda p: p.stat().st_mtime
    )
    result: list[tuple[Path, dict[str, Any]]] = []
    for path in artifacts:
        try:
            payload = json.loads(path.read_text(encoding="utf-8"))
            result.append((path, payload))
        except json.JSONDecodeError:
            continue
    return result


def render_diff(previous: dict[str, Any], current: dict[str, Any]) -> list[str]:
    prev_summary = previous.get("summary", {})
    curr_summary = current.get("summary", {})
    prev_total = int(prev_summary.get("total", 0))
    prev_passed = int(prev_summary.get("passed", 0))
    prev_failed = int(prev_summary.get("failed", 0))
    curr_total = int(curr_summary.get("total", 0))
    curr_passed = int(curr_summary.get("passed", 0))
    curr_failed = int(curr_summary.get("failed", 0))

    lines = [
        "## Full Validation Diff Summary",
        "",
        "| Metric | Previous | Current | Delta |",
        "|---|---:|---:|---:|",
        f"| Total checks | {prev_total} | {curr_total} | {curr_total - prev_total:+d} |",
        f"| Passed checks | {prev_passed} | {curr_passed} | {curr_passed - prev_passed:+d} |",
        f"| Failed checks | {prev_failed} | {curr_failed} | {curr_failed - prev_failed:+d} |",
    ]

    prev_categories = previous.get("categories", {})
    curr_categories = current.get("categories", {})
    all_categories = sorted(set(prev_categories.keys()) | set(curr_categories.keys()))
    if all_categories:
        lines.extend(["", "### Category Deltas", ""])
        lines.extend(
            [
                "| Category | Prev Failed | Curr Failed | Delta |",
                "|---|---:|---:|---:|",
            ]
        )
        for category in all_categories:
            prev_failed_cat = int(prev_categories.get(category, {}).get("failed", 0))
            curr_failed_cat = int(curr_categories.get(category, {}).get("failed", 0))
            lines.append(
                f"| {category} | {prev_failed_cat} | {curr_failed_cat} | {curr_failed_cat - prev_failed_cat:+d} |"
            )

    return lines


def main() -> int:
    artifacts = load_json_artifacts()
    if not artifacts:
        print("No validation artifacts available for diff summary")
        return 0

    summary_path = os.getenv("GITHUB_STEP_SUMMARY")
    output_lines: list[str]

    if len(artifacts) == 1:
        current_path, current_payload = artifacts[-1]
        curr_summary = current_payload.get("summary", {})
        output_lines = [
            "## Full Validation Diff Summary",
            "",
            f"Only one artifact found: `{current_path.name}`",
            f"- Total checks: {int(curr_summary.get('total', 0))}",
            f"- Passed: {int(curr_summary.get('passed', 0))}",
            f"- Failed: {int(curr_summary.get('failed', 0))}",
        ]
    else:
        _, previous = artifacts[-2]
        _, current = artifacts[-1]
        output_lines = render_diff(previous, current)

    body = "\n".join(output_lines) + "\n"
    if summary_path:
        with open(summary_path, "a", encoding="utf-8") as handle:
            handle.write(body)
    else:
        print(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
