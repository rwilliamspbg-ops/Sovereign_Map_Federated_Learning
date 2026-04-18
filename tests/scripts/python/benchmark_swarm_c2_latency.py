#!/usr/bin/env python3
"""Concurrent latency benchmark for C2 swarm endpoints.

Usage:
  python3 tests/scripts/python/benchmark_swarm_c2_latency.py \
    --base-url http://localhost:8000 \
    --token local-dev-admin-token \
    --iterations 200 --concurrency 16
"""

from __future__ import annotations

import argparse
import collections
import json
import math
import statistics
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from typing import Dict, List


def _request_json(
    url: str,
    method: str = "GET",
    headers: Dict[str, str] | None = None,
    body: Dict | None = None,
) -> Dict:
    payload = None
    request_headers = headers.copy() if headers else {}
    if body is not None:
        payload = json.dumps(body).encode("utf-8")
        request_headers["Content-Type"] = "application/json"

    req = urllib.request.Request(
        url=url, method=method, headers=request_headers, data=payload
    )
    with urllib.request.urlopen(req, timeout=8) as resp:
        response_raw = resp.read().decode("utf-8")
        if not response_raw:
            return {}
        return json.loads(response_raw)


def _percentile(values: List[float], pct: float) -> float:
    if not values:
        return 0.0
    idx = max(0, min(len(values) - 1, int(math.ceil((pct / 100.0) * len(values))) - 1))
    ordered = sorted(values)
    return ordered[idx]


def run(args: argparse.Namespace) -> int:
    token = args.token.strip()
    headers = {"X-API-Role": args.role}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    map_latencies_ms: List[float] = []
    cmd_latencies_ms: List[float] = []
    map_errors = 0
    cmd_errors = 0
    map_http_errors = collections.Counter()
    cmd_http_errors = collections.Counter()
    command_iterations = (
        args.command_iterations if args.command_iterations > 0 else args.iterations
    )

    def benchmark_map(index: int) -> None:
        nonlocal map_errors
        start = time.perf_counter()
        try:
            _request_json(
                f"{args.base_url.rstrip('/')}/swarm/map?limit=300&layers=paths,coverage",
                headers=headers,
            )
            latency_ms = (time.perf_counter() - start) * 1000.0
            map_latencies_ms.append(latency_ms)
        except urllib.error.HTTPError as exc:
            map_errors += 1
            map_http_errors[str(exc.code)] += 1
        except (urllib.error.URLError, TimeoutError, ValueError):
            map_errors += 1

    def benchmark_command(index: int) -> None:
        nonlocal cmd_errors
        start = time.perf_counter()
        try:
            _request_json(
                f"{args.base_url.rstrip('/')}/swarm/command",
                method="POST",
                headers=headers,
                body={
                    "command": "hold_position",
                    "target_scope": "group",
                    "target_ids": ["bench-squad"],
                    "client_nonce": f"bench-{int(time.time())}-{index}",
                    "parameters": {},
                },
            )
            latency_ms = (time.perf_counter() - start) * 1000.0
            cmd_latencies_ms.append(latency_ms)
        except urllib.error.HTTPError as exc:
            cmd_errors += 1
            cmd_http_errors[str(exc.code)] += 1
        except (urllib.error.URLError, TimeoutError, ValueError):
            cmd_errors += 1

    with ThreadPoolExecutor(max_workers=max(1, args.concurrency)) as pool:
        list(pool.map(benchmark_map, range(args.iterations)))
        list(pool.map(benchmark_command, range(command_iterations)))

    result = {
        "status": "ok",
        "base_url": args.base_url,
        "role": args.role,
        "iterations": args.iterations,
        "command_iterations": command_iterations,
        "concurrency": args.concurrency,
        "map": {
            "count": len(map_latencies_ms),
            "errors": map_errors,
            "http_error_codes": dict(map_http_errors),
            "mean_ms": (
                round(statistics.fmean(map_latencies_ms), 3)
                if map_latencies_ms
                else 0.0
            ),
            "p95_ms": round(_percentile(map_latencies_ms, 95.0), 3),
            "p99_ms": round(_percentile(map_latencies_ms, 99.0), 3),
        },
        "command": {
            "count": len(cmd_latencies_ms),
            "errors": cmd_errors,
            "http_error_codes": dict(cmd_http_errors),
            "mean_ms": (
                round(statistics.fmean(cmd_latencies_ms), 3)
                if cmd_latencies_ms
                else 0.0
            ),
            "p95_ms": round(_percentile(cmd_latencies_ms, 95.0), 3),
            "p99_ms": round(_percentile(cmd_latencies_ms, 99.0), 3),
        },
    }

    print(json.dumps(result, sort_keys=True))
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Benchmark swarm C2 map/command endpoint latency."
    )
    parser.add_argument(
        "--base-url", default="http://localhost:8000", help="Backend base URL"
    )
    parser.add_argument("--token", default="", help="Admin token for /swarm/command")
    parser.add_argument(
        "--role", default="admin", help="Role header sent as X-API-Role"
    )
    parser.add_argument(
        "--iterations", type=int, default=120, help="Requests per endpoint"
    )
    parser.add_argument(
        "--command-iterations",
        type=int,
        default=0,
        help="Requests for /swarm/command (0 uses --iterations)",
    )
    parser.add_argument(
        "--concurrency", type=int, default=12, help="Concurrent workers"
    )
    return parser.parse_args()


if __name__ == "__main__":
    raise SystemExit(run(parse_args()))
