#!/usr/bin/env python3
"""Browser runtime E2E check for render-cadence stability."""

from __future__ import annotations

import os
import subprocess
import sys


def run_command(command: list[str], env: dict[str, str], timeout: int) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        command,
        text=True,
        capture_output=True,
        env=env,
        timeout=timeout,
        check=False,
    )


def main() -> int:
    if os.getenv("PLAYWRIGHT_E2E_ENABLED", "0") != "1":
        print('{"status":"skipped","reason":"PLAYWRIGHT_E2E_ENABLED!=1"}')
        return 0

    env = {
        **os.environ,
        "CI": os.getenv("CI", "1"),
        "FRONTEND_E2E_URL": os.getenv("FRONTEND_E2E_URL", "http://127.0.0.1:4173"),
        # Force a slower chart update cadence to make throttle validation deterministic.
        "VITE_CHART_THROTTLE_MS": os.getenv("VITE_CHART_THROTTLE_MS", "1000"),
    }

    install = run_command(
        ["npx", "--yes", "playwright@1.52.0", "install", "chromium"],
        env=env,
        timeout=900,
    )
    sys.stdout.write(install.stdout)
    sys.stderr.write(install.stderr)
    if install.returncode != 0:
        return install.returncode

    test = run_command(
        [
            "npx",
            "--yes",
            "playwright@1.52.0",
            "test",
            "tests/e2e/runtime-cadence.spec.js",
            "--config",
            "tests/e2e/playwright.config.js",
            "--reporter=line",
        ],
        env=env,
        timeout=1200,
    )
    sys.stdout.write(test.stdout)
    sys.stderr.write(test.stderr)
    return test.returncode


if __name__ == "__main__":
    raise SystemExit(main())
