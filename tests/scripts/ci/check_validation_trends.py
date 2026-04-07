#!/usr/bin/env python3
"""Enforce validation trend SLOs from latest full-validation artifacts."""

from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[3]
OUT_DIR = REPO_ROOT / "test-results" / "full-validation"


def load_latest_json() -> dict[str, Any]:
    candidates = sorted(
        OUT_DIR.glob("full_validation_*.json"), key=lambda p: p.stat().st_mtime
    )
    if not candidates:
        raise FileNotFoundError("No full_validation_*.json artifact found")
    return json.loads(candidates[-1].read_text(encoding="utf-8"))


def load_history() -> list[dict[str, Any]]:
    history_file = OUT_DIR / "history.jsonl"
    if not history_file.exists():
        return []
    records: list[dict[str, Any]] = []
    for line in history_file.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            records.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return records


def main() -> int:
    latest = load_latest_json()
    history = load_history()
    failures: list[str] = []

    summary = latest.get("summary", {})
    categories = latest.get("categories", {})
    total_failed = int(summary.get("failed", 0))

    if total_failed > 0:
        failures.append(f"suite has failing checks: {total_failed}")

    enforce_zero_fail = os.getenv("VALIDATION_ENFORCE_CATEGORY_ZERO_FAIL", "1") == "1"
    if enforce_zero_fail:
        for category, stats in categories.items():
            failed = int(stats.get("failed", 0))
            if failed > 0:
                failures.append(f"category '{category}' has failures: {failed}")

    if len(history) >= 2:
        current = history[-1]
        previous = history[-2]
        current_duration = float(current.get("duration_seconds", 0.0))
        previous_duration = float(previous.get("duration_seconds", 0.0))
        if previous_duration > 0:
            regression_pct = (
                (current_duration - previous_duration) / previous_duration
            ) * 100.0
            max_regression_pct = float(
                os.getenv("VALIDATION_MAX_DURATION_REGRESSION_PCT", "35")
            )
            print(
                f"duration regression: {regression_pct:.2f}% (threshold {max_regression_pct:.2f}%)"
            )
            if regression_pct > max_regression_pct:
                failures.append(
                    f"duration regression {regression_pct:.2f}% exceeds threshold {max_regression_pct:.2f}%"
                )
    else:
        print("duration trend check skipped: insufficient history")

    if failures:
        print("trend SLO check failed:")
        for item in failures:
            print(f"- {item}")
        return 1

    print("trend SLO check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
