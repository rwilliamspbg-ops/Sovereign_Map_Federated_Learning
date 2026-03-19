#!/usr/bin/env python3
"""Validate local markdown links in active documentation scope.

Scope:
- Documentation/**
- documentation/**
- .github/**/*.md

Excludes archive and node_modules paths.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from urllib.parse import unquote

REPO_ROOT = Path(__file__).resolve().parents[1]
TARGET_ROOTS = [
    REPO_ROOT / "Documentation",
    REPO_ROOT / "documentation",
    REPO_ROOT / ".github",
]
SKIP_DIRS = {"archive", "node_modules", ".git"}
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)")


def iter_markdown_files() -> list[Path]:
    files: list[Path] = []
    for root in TARGET_ROOTS:
        if not root.exists():
            continue
        for path in root.rglob("*.md"):
            if any(part in SKIP_DIRS for part in path.parts):
                continue
            files.append(path)
    return sorted(files)


def is_external(target: str) -> bool:
    lower = target.lower()
    return (
        "://" in target
        or lower.startswith("mailto:")
        or lower.startswith("tel:")
        or lower.startswith("javascript:")
        or lower.startswith("data:")
    )


def main() -> int:
    files = iter_markdown_files()
    broken: list[tuple[str, int, str, str]] = []

    for file_path in files:
        rel_file = file_path.relative_to(REPO_ROOT).as_posix()
        lines = file_path.read_text(encoding="utf-8", errors="ignore").splitlines()
        for line_no, line in enumerate(lines, start=1):
            for match in LINK_RE.finditer(line):
                raw_target = match.group(1).strip()
                if (
                    not raw_target
                    or raw_target.startswith("#")
                    or is_external(raw_target)
                ):
                    continue

                clean_target = raw_target.split("#", 1)[0].split("?", 1)[0].strip()
                clean_target = unquote(clean_target)
                if not clean_target:
                    continue

                if clean_target.startswith("/"):
                    candidate = (REPO_ROOT / clean_target.lstrip("/")).resolve()
                else:
                    candidate = (file_path.parent / clean_target).resolve()

                if not str(candidate).startswith(str(REPO_ROOT)):
                    continue

                if not candidate.exists():
                    # Historical docs links commonly point to /archive/** while
                    # archived content now lives under docs/archive/**.
                    try:
                        candidate_rel = candidate.relative_to(REPO_ROOT).as_posix()
                    except ValueError:
                        candidate_rel = ""

                    if candidate_rel == "archive" or candidate_rel.startswith(
                        "archive/"
                    ):
                        archive_fallback = (
                            REPO_ROOT / "docs" / candidate_rel
                        ).resolve()
                        if archive_fallback.exists():
                            candidate = archive_fallback

                if not candidate.exists():
                    broken.append(
                        (
                            rel_file,
                            line_no,
                            raw_target,
                            candidate.relative_to(REPO_ROOT).as_posix(),
                        )
                    )

    if broken:
        print(f"Broken local links: {len(broken)}")
        for rel_file, line_no, target, resolved in broken:
            print(f"- {rel_file}:{line_no} :: {target} -> {resolved}")
        return 1

    print(f"OK: checked {len(files)} markdown files with no broken local links.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
