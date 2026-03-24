#!/usr/bin/env python3
"""Export a compact incident bundle from local observability endpoints.

This script captures raw endpoint payloads plus a small metadata summary to
speed first-response and postmortem evidence collection.
"""

from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path
import urllib.error
import urllib.request


DEFAULT_TARGETS = {
    "backend_metrics": "http://localhost:8000/metrics",
    "backend_metrics_summary": "http://localhost:8000/metrics_summary",
    "tpm_metrics": "http://localhost:9091/metrics",
    "prometheus_targets": "http://localhost:9090/api/v1/targets",
    "prometheus_alerts": "http://localhost:9090/api/v1/alerts",
}


def fetch_text(url: str, timeout: float) -> tuple[bool, str]:
    try:
        with urllib.request.urlopen(url, timeout=timeout) as response:
            body = response.read().decode("utf-8", errors="replace")
            return True, body
    except urllib.error.URLError as exc:
        return False, f"ERROR: {exc}"


def write_bundle(output_root: Path, timeout: float) -> Path:
    ts = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    bundle_dir = output_root / f"incident-bundle-{ts}"
    bundle_dir.mkdir(parents=True, exist_ok=False)

    summary: dict[str, object] = {
        "timestamp_utc": ts,
        "targets": {},
    }

    for key, url in DEFAULT_TARGETS.items():
        ok, body = fetch_text(url, timeout)
        ext = "json" if key.startswith("prometheus_") or key.endswith("summary") else "prom"
        out_file = bundle_dir / f"{key}.{ext}"
        out_file.write_text(body, encoding="utf-8")
        summary["targets"][key] = {
            "url": url,
            "ok": ok,
            "bytes": len(body.encode("utf-8", errors="replace")),
            "file": out_file.name,
        }

    (bundle_dir / "summary.json").write_text(
        json.dumps(summary, indent=2, sort_keys=True),
        encoding="utf-8",
    )
    return bundle_dir


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export incident evidence bundle")
    parser.add_argument(
        "--output-dir",
        default="artifacts/incidents",
        help="Directory where incident bundles are written",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=5.0,
        help="HTTP timeout in seconds per endpoint",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    output_root = Path(args.output_dir)
    bundle_dir = write_bundle(output_root, args.timeout)
    print(f"Incident bundle written: {bundle_dir}")
    print(f"Summary file: {bundle_dir / 'summary.json'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
