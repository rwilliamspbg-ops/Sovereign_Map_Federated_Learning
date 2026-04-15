#!/usr/bin/env python3
"""Validate local image assets referenced by markdown and inline HTML image tags."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path
from urllib.parse import unquote

REPO_ROOT = Path(__file__).resolve().parents[1]
MD_IMAGE_RE = re.compile(r"!\[[^\]]*\]\(([^)]+)\)")
HTML_IMAGE_RE = re.compile(r"<img\b[^>]*\bsrc=[\"']([^\"']+)[\"']", re.IGNORECASE)


def is_external(target: str) -> bool:
    lower = target.lower()
    return (
        "://" in target
        or lower.startswith("mailto:")
        or lower.startswith("tel:")
        or lower.startswith("javascript:")
        or lower.startswith("data:")
    )


def resolve_file(path_str: str) -> Path:
    candidate = (REPO_ROOT / path_str).resolve()
    if not str(candidate).startswith(str(REPO_ROOT)):
        raise ValueError(f"path escapes repository root: {path_str}")
    if not candidate.exists():
        raise FileNotFoundError(f"missing file: {path_str}")
    return candidate


def resolve_asset(base_file: Path, raw_target: str) -> Path | None:
    clean_target = raw_target.split("#", 1)[0].split("?", 1)[0].strip()
    clean_target = unquote(clean_target)
    if not clean_target or is_external(clean_target):
        return None

    if clean_target.startswith("/"):
        candidate = (REPO_ROOT / clean_target.lstrip("/")).resolve()
    else:
        candidate = (base_file.parent / clean_target).resolve()

    if not str(candidate).startswith(str(REPO_ROOT)):
        return None

    return candidate


def validate_file(file_path: Path) -> list[tuple[str, int, str, str]]:
    broken: list[tuple[str, int, str, str]] = []
    rel_file = file_path.relative_to(REPO_ROOT).as_posix()
    lines = file_path.read_text(encoding="utf-8", errors="ignore").splitlines()

    for line_no, line in enumerate(lines, start=1):
        for pattern in (MD_IMAGE_RE, HTML_IMAGE_RE):
            for match in pattern.finditer(line):
                raw_target = match.group(1).strip()
                candidate = resolve_asset(file_path, raw_target)
                if candidate is None:
                    continue
                if not candidate.exists():
                    broken.append(
                        (
                            rel_file,
                            line_no,
                            raw_target,
                            candidate.relative_to(REPO_ROOT).as_posix(),
                        )
                    )

    return broken


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Validate image asset references for selected markdown files"
    )
    parser.add_argument("files", nargs="+", help="Repository-relative markdown files")
    args = parser.parse_args()

    broken: list[tuple[str, int, str, str]] = []
    checked = 0

    for file_arg in args.files:
        file_path = resolve_file(file_arg)
        checked += 1
        broken.extend(validate_file(file_path))

    if broken:
        print(f"Broken local image references: {len(broken)}")
        for rel_file, line_no, target, resolved in broken:
            print(f"- {rel_file}:{line_no} :: {target} -> {resolved}")
        return 1

    print(
        f"OK: checked {checked} markdown files with no broken local image references."
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
