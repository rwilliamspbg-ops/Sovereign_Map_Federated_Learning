#!/usr/bin/env python3
"""Validate Grafana dashboard queries reference known metrics.

This catches broken dashboards early in CI by ensuring queried metrics are either:
- explicitly exported runtime metrics (allowlist), or
- recording rule outputs defined in dashboard compatibility rules.
"""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DASH_DIR = ROOT / "grafana" / "provisioning" / "dashboards"
RULES_FILE = ROOT / "dashboard_compat_rules.yml"
ALLOWLIST_FILE = ROOT / "scripts" / "known_metrics_allowlist.txt"

METRIC_RE = re.compile(r"\b([a-zA-Z_][a-zA-Z0-9_:]*)\b")
LABEL_KEY_RE = re.compile(r"\b([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:=~|!~|=|!=)")
GROUP_LABELS_RE = re.compile(r"\b(?:by|without|on|ignoring|group_left|group_right)\s*\(([^)]*)\)")
PROMQL_KEYWORDS = {
    "sum",
    "avg",
    "min",
    "max",
    "rate",
    "irate",
    "increase",
    "delta",
    "deriv",
    "avg_over_time",
    "predict_linear",
    "histogram_quantile",
    "clamp_min",
    "clamp_max",
    "vector",
    "by",
    "without",
    "on",
    "group_left",
    "group_right",
    "and",
    "or",
    "unless",
    "bool",
}


def load_allowlist() -> set[str]:
    lines = ALLOWLIST_FILE.read_text(encoding="utf-8").splitlines()
    return {line.strip() for line in lines if line.strip() and not line.startswith("#")}


def load_recording_rules() -> set[str]:
    records: set[str] = set()
    for line in RULES_FILE.read_text(encoding="utf-8").splitlines():
        stripped = line.strip()
        if stripped.startswith("- record:"):
            records.add(stripped.split(":", 1)[1].strip())
    return records


def extract_metrics(expr: str) -> set[str]:
    label_keys = {m.group(1) for m in LABEL_KEY_RE.finditer(expr)}
    for match in GROUP_LABELS_RE.finditer(expr):
        for raw_label in match.group(1).split(","):
            label = raw_label.strip()
            if label:
                label_keys.add(label)
    candidates = set()
    for token in METRIC_RE.findall(expr):
        if token in PROMQL_KEYWORDS:
            continue
        if token[0].isdigit():
            continue
        if token in {"true", "false", "inf", "nan"}:
            continue
        if token.startswith("__"):
            continue
        # Label keys inside selectors (for example, metric{event_type="x"})
        # are not metric names and should not be validated as metrics.
        if token in label_keys:
            continue
        # Keep only metric-like identifiers.
        if "_" not in token and ":" not in token:
            continue
        candidates.add(token)
    return candidates


def main() -> int:
    allowlist = load_allowlist()
    rules = load_recording_rules()
    known = allowlist | rules

    unknown: dict[str, set[str]] = {}

    for dash_file in sorted(DASH_DIR.glob("*.json")):
        payload = json.loads(dash_file.read_text(encoding="utf-8"))
        missing: set[str] = set()
        for panel in payload.get("panels", []):
            for target in panel.get("targets", []):
                expr = target.get("expr", "")
                for metric in extract_metrics(expr):
                    if metric not in known:
                        missing.add(metric)
        if missing:
            unknown[dash_file.name] = missing

    if unknown:
        print("Dashboard query validation failed:")
        for dash_name, metrics in unknown.items():
            print(f"- {dash_name}: {', '.join(sorted(metrics))}")
        return 1

    print("Dashboard query validation passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
