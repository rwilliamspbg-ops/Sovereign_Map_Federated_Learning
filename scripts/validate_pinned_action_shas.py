#!/usr/bin/env python3
"""Validate that SHA-pinned GitHub Actions references resolve to real commits."""

from __future__ import annotations

import os
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path

USES_RE = re.compile(r"^\s*uses:\s*([^\s#]+)")
SHA_RE = re.compile(r"^[0-9a-fA-F]{40}$")
WORKFLOWS_DIR = Path(".github/workflows")


def collect_pinned_refs() -> set[tuple[str, str, str, str]]:
    refs: set[tuple[str, str, str, str]] = set()

    for workflow in sorted(WORKFLOWS_DIR.glob("*.yml")):
        lines = workflow.read_text(encoding="utf-8", errors="ignore").splitlines()
        for line in lines:
            match = USES_RE.match(line)
            if not match:
                continue

            ref = match.group(1)
            if ref.startswith("./") or ref.startswith("docker://") or "@" not in ref:
                continue

            action, rev = ref.rsplit("@", 1)
            if not SHA_RE.fullmatch(rev):
                continue

            parts = action.split("/")
            if len(parts) < 2:
                continue

            repo = "/".join(parts[:2])
            refs.add((workflow.as_posix(), action, repo, rev))

    return refs


def resolve_pins(
    refs: set[tuple[str, str, str, str]],
) -> list[tuple[str, str, str, str]]:
    token = os.environ.get("GITHUB_TOKEN", "")
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "workflow-action-pin-check",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"

    failures: list[tuple[str, str, str, str]] = []

    for workflow, action, repo, rev in sorted(refs):
        url = f"https://api.github.com/repos/{repo}/commits/{rev}"
        req = urllib.request.Request(url, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=20) as resp:
                if resp.status != 200:
                    failures.append((workflow, action, rev, f"HTTP {resp.status}"))
        except urllib.error.HTTPError as exc:
            failures.append((workflow, action, rev, f"HTTP {exc.code}"))
        except Exception as exc:  # pragma: no cover
            failures.append((workflow, action, rev, str(exc)))

    return failures


def main() -> int:
    refs = collect_pinned_refs()
    if not refs:
        print("No pinned action SHAs found to resolve.")
        return 0

    failures = resolve_pins(refs)
    if failures:
        print("Found unresolved pinned action SHAs:")
        for workflow, action, rev, reason in failures:
            print(f"- {workflow}: {action}@{rev} -> {reason}")
        return 1

    print(f"All pinned action SHAs resolved successfully ({len(refs)} refs).")
    return 0


if __name__ == "__main__":
    sys.exit(main())
