#!/usr/bin/env python3
"""Runtime status helper for Python liboqs hybrid KEM support."""

from __future__ import annotations

import argparse
import json


def status() -> dict[str, object]:
    result: dict[str, object] = {
        "adapter": "python-liboqs-hybrid-kem",
        "available": False,
        "kem_algorithm": "ML-KEM-768",
        "hybrid_kem_enabled": False,
        "reason": "python oqs module not available",
    }

    try:
        import oqs  # type: ignore

        _ = oqs.get_enabled_kem_mechanisms()
    except Exception as exc:  # pragma: no cover - environment dependent
        result["reason"] = f"python oqs import failed: {exc}"
        return result

    result["available"] = True
    result["hybrid_kem_enabled"] = True
    result.pop("reason", None)
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--status", action="store_true", help="print adapter status as JSON"
    )
    args = parser.parse_args()

    if args.status:
        print(json.dumps(status(), indent=2))
        return 0

    parser.print_help()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
