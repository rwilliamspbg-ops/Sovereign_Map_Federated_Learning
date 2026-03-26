#!/usr/bin/env python3
"""Detect stale docker compose service references in shell scripts."""

from __future__ import annotations

import re
import shlex
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
COMPOSE_FILE = ROOT / "docker-compose.full.yml"
SCRIPTS = [
    ROOT / "scripts" / "demo-10min-auditable.sh",
    ROOT / "scripts" / "burst-scale-with-certs.sh",
]

UP_CMD_RE = re.compile(r"docker\s+compose[^\n]*\s+up\s+-d\s+(.+)")
SERVICE_KEY_RE = re.compile(r"^\s{2}([a-zA-Z0-9_-]+):\s*$")


def parse_services() -> set[str]:
    lines = COMPOSE_FILE.read_text(encoding="utf-8").splitlines()
    services: set[str] = set()
    in_services = False

    for line in lines:
        if line.startswith("services:"):
            in_services = True
            continue
        if in_services and line and not line.startswith(" "):
            break
        if in_services:
            match = SERVICE_KEY_RE.match(line)
            if match:
                services.add(match.group(1))

    return services


def parse_service_tokens(arg_text: str) -> list[str]:
    tokens = shlex.split(arg_text)
    services: list[str] = []
    for token in tokens:
        if token.startswith("-"):
            continue
        if "=" in token:
            continue
        if "$" in token:
            continue
        if re.fullmatch(r"[a-zA-Z0-9_-]+", token):
            services.append(token)
    return services


def main() -> int:
    known = parse_services()
    unknown: list[str] = []

    for script in SCRIPTS:
        lines = script.read_text(encoding="utf-8").splitlines()
        for idx, line in enumerate(lines, start=1):
            match = UP_CMD_RE.search(line)
            if not match:
                continue
            for service in parse_service_tokens(match.group(1)):
                if service not in known:
                    unknown.append(f"{script.relative_to(ROOT)}:{idx}: {service}")

    if unknown:
        print("Compose service reference check failed:")
        for item in unknown:
            print(f"- {item}")
        return 1

    print("Compose service reference check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
