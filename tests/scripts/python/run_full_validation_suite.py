#!/usr/bin/env python3
"""Run a consolidated capability + security + performance validation suite."""

from __future__ import annotations

import json
import os
import argparse
import subprocess
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional

REPO_ROOT = Path(__file__).resolve().parents[3]
OUT_DIR = REPO_ROOT / "test-results" / "full-validation"
OUT_DIR.mkdir(parents=True, exist_ok=True)
STAMP = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
JSON_OUT = OUT_DIR / f"full_validation_{STAMP}.json"
MD_OUT = OUT_DIR / f"full_validation_{STAMP}.md"
HISTORY_FILE = OUT_DIR / "history.jsonl"


@dataclass
class CheckResult:
    name: str
    category: str
    command: str
    passed: bool
    return_code: int
    duration_seconds: float
    output_file: str


BASE_CHECKS = [
    {
        "name": "API coverage validator",
        "category": "capabilities",
        "command": "node scripts/validate-api-coverage.js",
        "timeout": 180,
    },
    {
        "name": "Python communication contracts",
        "category": "capabilities",
        "command": "python tests/scripts/python/test_communication_contracts.py",
        "timeout": 180,
    },
    {
        "name": "Live backend HTTP E2E",
        "category": "capabilities",
        "command": "python tests/scripts/python/test_backend_live_e2e.py",
        "timeout": 420,
    },
    {
        "name": "Mobile gradient signature contract",
        "category": "capabilities",
        "command": "python tests/scripts/python/test_mobile_verify_gradient_contract.py",
        "timeout": 240,
    },
    {
        "name": "Marketplace local contracts",
        "category": "functions",
        "command": "python tests/scripts/python/test_marketplace_local_contracts.py",
        "timeout": 300,
    },
    {
        "name": "Marketplace negative-path contracts",
        "category": "functions",
        "command": "python tests/scripts/python/test_marketplace_negative_paths.py",
        "timeout": 300,
    },
    {
        "name": "Backend security controls",
        "category": "security",
        "command": "python tests/scripts/python/test_security_controls.py",
        "timeout": 240,
    },
    {
        "name": "Security fuzz controls",
        "category": "security",
        "command": "python tests/scripts/python/test_security_fuzz_controls.py",
        "timeout": 300,
    },
    {
        "name": "Performance regression thresholds",
        "category": "performance",
        "command": "python tests/scripts/python/test_performance_regression_thresholds.py",
        "timeout": 300,
    },
    {
        "name": "Dependency security audit",
        "category": "security",
        "command": "npm run security:check",
        "timeout": 420,
    },
    {
        "name": "Frontend unit tests",
        "category": "functions",
        "command": "npm --prefix frontend run test -- --run",
        "timeout": 420,
    },
    {
        "name": "SDK package unit tests",
        "category": "capabilities",
        "command": "npm run test:ci",
        "timeout": 900,
    },
]

DEEP_ONLY_CHECKS = [
    {
        "name": "Browser runtime render cadence E2E",
        "category": "performance",
        "command": "python tests/scripts/python/test_browser_runtime_e2e.py",
        "timeout": 1200,
    },
    {
        "name": "Scheduled chaos soak guard",
        "category": "performance",
        "command": "python tests/scripts/python/test_soak_chaos_guard.py",
        "timeout": 1200,
    },
]

CHECKS_BY_PROFILE = {
    "fast": BASE_CHECKS,
    "deep": BASE_CHECKS + DEEP_ONLY_CHECKS,
}


def _safe_write_text(path: Path, content: str) -> None:
    """Write text and recover if a concurrent cleanup removed the parent dir."""
    try:
        path.write_text(content, encoding="utf-8")
    except FileNotFoundError:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")


def run_check(name: str, category: str, command: str, timeout: int) -> CheckResult:
    started = datetime.now(timezone.utc)
    safe_name = (
        name.lower()
        .replace(" ", "_")
        .replace("/", "_")
        .replace(":", "")
        .replace("-", "_")
    )
    output_file = OUT_DIR / f"{STAMP}_{safe_name}.log"
    output_file.parent.mkdir(parents=True, exist_ok=True)

    try:
        proc = subprocess.run(
            command,
            cwd=REPO_ROOT,
            shell=True,
            text=True,
            capture_output=True,
            timeout=timeout,
            env={**os.environ, "PYTHONUNBUFFERED": "1"},
        )
        _safe_write_text(output_file, (proc.stdout or "") + "\n" + (proc.stderr or ""))
        passed = proc.returncode == 0
        return_code = proc.returncode
    except subprocess.TimeoutExpired as exc:
        _safe_write_text(output_file, (exc.stdout or "") + "\n" + (exc.stderr or "") + "\n[TIMEOUT]")
        passed = False
        return_code = 124

    duration = (datetime.now(timezone.utc) - started).total_seconds()
    return CheckResult(
        name=name,
        category=category,
        command=command,
        passed=passed,
        return_code=return_code,
        duration_seconds=round(duration, 2),
        output_file=str(output_file.relative_to(REPO_ROOT)),
    )



def aggregate_by_category(results: List[CheckResult]) -> Dict[str, Dict[str, int]]:
    grouped: Dict[str, Dict[str, int]] = {}
    for row in results:
        grouped.setdefault(row.category, {"passed": 0, "failed": 0})
        grouped[row.category]["passed" if row.passed else "failed"] += 1
    return grouped



def load_last_history_record() -> Optional[Dict[str, object]]:
    if not HISTORY_FILE.exists():
        return None

    lines = HISTORY_FILE.read_text(encoding="utf-8").splitlines()
    for line in reversed(lines):
        line = line.strip()
        if not line:
            continue
        try:
            return json.loads(line)
        except json.JSONDecodeError:
            continue
    return None



def append_history_record(payload: Dict[str, object]) -> None:
    HISTORY_FILE.parent.mkdir(parents=True, exist_ok=True)
    summary = payload.get("summary", {})
    duration = sum(float(item.get("duration_seconds", 0.0)) for item in payload.get("results", []))
    record = {
        "timestamp_utc": payload.get("timestamp_utc"),
        "total": int(summary.get("total", 0)),
        "passed": int(summary.get("passed", 0)),
        "failed": int(summary.get("failed", 0)),
        "duration_seconds": round(duration, 2),
        "categories": payload.get("categories", {}),
    }
    with HISTORY_FILE.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(record) + "\n")



def build_markdown(results: List[CheckResult], previous: Optional[Dict[str, object]]) -> str:
    total = len(results)
    passed = sum(1 for r in results if r.passed)
    failed = total - passed
    by_category = aggregate_by_category(results)
    duration = round(sum(r.duration_seconds for r in results), 2)

    lines = [
        "# Full Validation Suite Summary",
        "",
        f"- Timestamp (UTC): {datetime.now(timezone.utc).isoformat()}",
        f"- Total checks: {total}",
        f"- Passed: {passed}",
        f"- Failed: {failed}",
        f"- Total suite duration (s): {duration}",
        "",
        "## Trend vs Previous Run",
        "",
    ]

    if previous is None:
        lines.append("- No previous history record available.")
    else:
        prev_passed = int(previous.get("passed", 0))
        prev_failed = int(previous.get("failed", 0))
        prev_duration = float(previous.get("duration_seconds", 0.0))
        lines.append(f"- Passed delta: {passed - prev_passed:+d}")
        lines.append(f"- Failed delta: {failed - prev_failed:+d}")
        lines.append(f"- Duration delta (s): {round(duration - prev_duration, 2):+}")

    lines.extend([
        "",
        "## By Category",
        "",
    ])

    for category in sorted(by_category.keys()):
        stats = by_category[category]
        lines.append(f"- {category}: {stats['passed']} passed / {stats['failed']} failed")

    lines.extend([
        "",
        "## Detailed Results",
        "",
        "| Check | Category | Status | Exit | Duration (s) | Log |",
        "|------|----------|--------|------|--------------|-----|",
    ])

    for row in results:
        status = "PASS" if row.passed else "FAIL"
        lines.append(
            f"| {row.name} | {row.category} | {status} | {row.return_code} | {row.duration_seconds} | {row.output_file} |"
        )

    lines.extend([
        "",
        "## Immediate Gaps To Close",
        "",
        "1. Fix any failed checks shown above and re-run this suite.",
        "2. Keep live HTTP E2E, fuzz, and performance budgets required in CI.",
        "3. Track history.jsonl trends and investigate pass-rate or duration regressions.",
        "4. Add chaos/soak runs on schedule for long-horizon reliability.",
    ])

    return "\n".join(lines) + "\n"



def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run consolidated validation checks")
    parser.add_argument(
        "--profile",
        choices=sorted(CHECKS_BY_PROFILE.keys()),
        default=os.getenv("VALIDATION_PROFILE", "fast"),
        help="Suite profile to run (default: fast)",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    checks = CHECKS_BY_PROFILE[args.profile]
    results: List[CheckResult] = []
    for check in checks:
        results.append(
            run_check(
                name=check["name"],
                category=check["category"],
                command=check["command"],
                timeout=check["timeout"],
            )
        )

    categories = aggregate_by_category(results)
    payload = {
        "timestamp_utc": datetime.now(timezone.utc).isoformat(),
        "profile": args.profile,
        "results": [asdict(r) for r in results],
        "categories": categories,
        "summary": {
            "total": len(results),
            "passed": sum(1 for r in results if r.passed),
            "failed": sum(1 for r in results if not r.passed),
        },
    }

    previous = load_last_history_record()
    JSON_OUT.parent.mkdir(parents=True, exist_ok=True)
    JSON_OUT.write_text(json.dumps(payload, indent=2), encoding="utf-8")
    MD_OUT.write_text(build_markdown(results, previous), encoding="utf-8")
    append_history_record(payload)

    print(f"JSON summary: {JSON_OUT.relative_to(REPO_ROOT)}")
    print(f"Markdown summary: {MD_OUT.relative_to(REPO_ROOT)}")
    print(f"History file: {HISTORY_FILE.relative_to(REPO_ROOT)}")

    return 0 if payload["summary"]["failed"] == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
