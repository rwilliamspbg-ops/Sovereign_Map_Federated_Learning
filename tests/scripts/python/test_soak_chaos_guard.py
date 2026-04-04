#!/usr/bin/env python3
"""Run chaos soak checks when explicitly enabled in CI."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys


def main() -> int:
    if os.getenv("SOAK_CHAOS_ENABLED", "0") != "1":
        print('{"status":"skipped","reason":"SOAK_CHAOS_ENABLED!=1"}')
        return 0

    if shutil.which("docker") is None:
        print('{"status":"skipped","reason":"docker not available"}')
        return 0 if os.getenv("SOAK_CHAOS_STRICT", "0") != "1" else 1

    run = subprocess.run(
        ["python", "tests/scripts/python/testnet-chaos-suite.py"],
        text=True,
        capture_output=True,
        timeout=1200,
        check=False,
    )
    sys.stdout.write(run.stdout)
    sys.stderr.write(run.stderr)
    return run.returncode


if __name__ == "__main__":
    raise SystemExit(main())
