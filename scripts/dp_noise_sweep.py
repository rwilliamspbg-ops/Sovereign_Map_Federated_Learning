#!/usr/bin/env python3
"""Sweep DP noise multipliers and estimate epsilon for FL training.

If Opacus is available, this uses RDPAccountant for tighter estimates.
Otherwise, it falls back to a conservative approximation useful for
ranking candidate noise multipliers.
"""

from __future__ import annotations

import argparse
import math
from dataclasses import dataclass


@dataclass
class SweepRow:
    noise_multiplier: float
    epsilon: float
    meets_target: bool


def estimate_epsilon_fallback(
    noise_multiplier: float,
    sample_rate: float,
    steps: int,
    delta: float,
) -> float:
    """Conservative closed-form approximation.

    For small q and moderate sigma, a commonly used approximation is:
        epsilon ~= q * sqrt(2 * T * log(1/delta)) / sigma
    where q is sample_rate, T is steps, sigma is noise multiplier.
    """
    if noise_multiplier <= 0:
        return float("inf")
    return (
        sample_rate * math.sqrt(2.0 * steps * math.log(1.0 / delta)) / noise_multiplier
    )


def estimate_epsilon_opacus(
    noise_multiplier: float,
    sample_rate: float,
    steps: int,
    delta: float,
) -> float | None:
    try:
        from opacus.accountants import RDPAccountant  # type: ignore[import-not-found]
    except Exception:
        return None

    accountant = RDPAccountant()
    for _ in range(steps):
        accountant.step(noise_multiplier=noise_multiplier, sample_rate=sample_rate)
    return float(accountant.get_epsilon(delta=delta))


def sweep(
    noise_values: list[float],
    sample_rate: float,
    steps: int,
    delta: float,
    epsilon_target: float,
) -> tuple[list[SweepRow], str]:
    rows: list[SweepRow] = []
    mode = "opacus"

    for sigma in noise_values:
        eps = estimate_epsilon_opacus(sigma, sample_rate, steps, delta)
        if eps is None:
            mode = "fallback"
            eps = estimate_epsilon_fallback(sigma, sample_rate, steps, delta)
        rows.append(
            SweepRow(
                noise_multiplier=sigma,
                epsilon=eps,
                meets_target=eps <= epsilon_target,
            )
        )
    return rows, mode


def parse_noise_values(raw: str) -> list[float]:
    values = []
    for part in raw.split(","):
        part = part.strip()
        if not part:
            continue
        values.append(float(part))
    if not values:
        raise ValueError("at least one noise multiplier is required")
    return values


def print_results(
    rows: list[SweepRow],
    mode: str,
    sample_rate: float,
    steps: int,
    delta: float,
    epsilon_target: float,
) -> None:
    print("\nDP Noise Multiplier Sweep")
    print(
        "Assumptions: "
        f"sample_rate={sample_rate:.6f}, steps={steps}, delta={delta}, epsilon_target={epsilon_target}"
    )
    if mode == "fallback":
        print("Estimator: conservative closed-form fallback (Opacus unavailable)")
    else:
        print("Estimator: Opacus RDPAccountant")

    print("| Noise Multiplier | Estimated Epsilon | Meets Target |")
    print("|---:|---:|:---:|")
    for row in rows:
        marker = "yes" if row.meets_target else "no"
        print(f"| {row.noise_multiplier:.4f} | {row.epsilon:.4f} | {marker} |")

    best = min(rows, key=lambda r: r.epsilon)
    closest = min(rows, key=lambda r: abs(r.epsilon - epsilon_target))
    print("\nSummary:")
    print(
        f"- Lowest epsilon in sweep: sigma={best.noise_multiplier:.4f}, epsilon={best.epsilon:.4f}"
    )
    print(
        f"- Closest to target: sigma={closest.noise_multiplier:.4f}, epsilon={closest.epsilon:.4f}"
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sweep DP noise multipliers")
    parser.add_argument(
        "--noise-values",
        type=str,
        default="0.5,0.8,1.1,1.5,2.0,3.0",
        help="Comma-separated candidate noise multipliers",
    )
    parser.add_argument(
        "--samples-per-node",
        type=int,
        default=120,
        help="Training samples per client/node",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=16,
        help="Local training batch size",
    )
    parser.add_argument(
        "--local-epochs",
        type=int,
        default=1,
        help="Local epochs per federated round",
    )
    parser.add_argument(
        "--rounds",
        type=int,
        default=30,
        help="Federated rounds composed for cumulative epsilon",
    )
    parser.add_argument("--delta", type=float, default=1e-5, help="DP delta")
    parser.add_argument(
        "--epsilon-target",
        type=float,
        default=10.0,
        help="Target cumulative epsilon",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    noise_values = parse_noise_values(args.noise_values)
    steps_per_round = args.local_epochs * math.ceil(
        args.samples_per_node / args.batch_size
    )
    total_steps = args.rounds * steps_per_round
    sample_rate = args.batch_size / args.samples_per_node

    rows, mode = sweep(
        noise_values=noise_values,
        sample_rate=sample_rate,
        steps=total_steps,
        delta=args.delta,
        epsilon_target=args.epsilon_target,
    )

    print_results(
        rows=rows,
        mode=mode,
        sample_rate=sample_rate,
        steps=total_steps,
        delta=args.delta,
        epsilon_target=args.epsilon_target,
    )


if __name__ == "__main__":
    main()
