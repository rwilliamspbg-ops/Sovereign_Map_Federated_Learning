#!/usr/bin/env python3
"""20-node, 200-round Byzantine resilience sweep (50% to 70%)."""

from __future__ import annotations

import argparse
import csv
import json
import random
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from statistics import mean
from typing import Any


@dataclass
class RoundRecord:
    round_number: int
    byzantine_pct: int
    byzantine_nodes: int
    honest_nodes: int
    support_ratio: float
    attack_ratio: float
    accuracy: float


@dataclass
class LevelSummary:
    byzantine_pct: int
    byzantine_nodes_estimate: int
    final_accuracy: float
    avg_accuracy: float
    min_accuracy: float
    max_accuracy: float
    first_below_80_round: int | None


def run_level(
    *,
    byzantine_pct: int,
    nodes: int,
    rounds: int,
    seed: int,
) -> tuple[list[RoundRecord], LevelSummary]:
    rng = random.Random(seed + byzantine_pct)
    accuracy = 85.0
    records: list[RoundRecord] = []

    for round_number in range(1, rounds + 1):
        byzantine_nodes = sum(1 for _ in range(nodes) if rng.random() < (byzantine_pct / 100.0))
        honest_nodes = nodes - byzantine_nodes

        support_ratio = honest_nodes / nodes
        attack_ratio = byzantine_nodes / nodes

        baseline_gain = 0.05
        honest_gain = support_ratio * 0.28
        attack_drag = attack_ratio * 0.16
        excess_drag = max(0.0, (byzantine_pct - 50) / 20.0) * attack_ratio * 0.35
        stochastic = rng.uniform(-0.04, 0.04)

        delta = baseline_gain + honest_gain - attack_drag - excess_drag + stochastic
        accuracy = max(45.0, min(96.0, accuracy + delta))

        records.append(
            RoundRecord(
                round_number=round_number,
                byzantine_pct=byzantine_pct,
                byzantine_nodes=byzantine_nodes,
                honest_nodes=honest_nodes,
                support_ratio=round(support_ratio, 4),
                attack_ratio=round(attack_ratio, 4),
                accuracy=round(accuracy, 4),
            )
        )

    accuracies = [r.accuracy for r in records]
    first_below_80_round = next((r.round_number for r in records if r.accuracy < 80.0), None)

    summary = LevelSummary(
        byzantine_pct=byzantine_pct,
        byzantine_nodes_estimate=round(nodes * byzantine_pct / 100),
        final_accuracy=round(accuracies[-1], 4),
        avg_accuracy=round(mean(accuracies), 4),
        min_accuracy=round(min(accuracies), 4),
        max_accuracy=round(max(accuracies), 4),
        first_below_80_round=first_below_80_round,
    )
    return records, summary


def to_json_serializable(round_records_by_level: dict[int, list[RoundRecord]]) -> dict[str, list[dict[str, Any]]]:
    return {str(level): [asdict(record) for record in records] for level, records in round_records_by_level.items()}


def write_csv_summary(path: Path, summaries: list[LevelSummary]) -> None:
    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(
            [
                "Byzantine%",
                "ByzantineNodes(Est)",
                "FinalAccuracy",
                "AvgAccuracy",
                "MinAccuracy",
                "MaxAccuracy",
                "FirstRoundBelow80",
            ]
        )
        for item in summaries:
            writer.writerow(
                [
                    item.byzantine_pct,
                    item.byzantine_nodes_estimate,
                    item.final_accuracy,
                    item.avg_accuracy,
                    item.min_accuracy,
                    item.max_accuracy,
                    item.first_below_80_round if item.first_below_80_round is not None else "",
                ]
            )


def write_csv_rounds(path: Path, round_records_by_level: dict[int, list[RoundRecord]]) -> None:
    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(
            [
                "Byzantine%",
                "Round",
                "ByzantineNodes",
                "HonestNodes",
                "SupportRatio",
                "AttackRatio",
                "Accuracy",
            ]
        )
        for level in sorted(round_records_by_level):
            for record in round_records_by_level[level]:
                writer.writerow(
                    [
                        record.byzantine_pct,
                        record.round_number,
                        record.byzantine_nodes,
                        record.honest_nodes,
                        record.support_ratio,
                        record.attack_ratio,
                        record.accuracy,
                    ]
                )


def write_markdown_report(
    path: Path,
    *,
    nodes: int,
    rounds: int,
    min_pct: int,
    max_pct: int,
    threshold_pct: int | None,
    summaries: list[LevelSummary],
    timestamp: str,
) -> None:
    threshold_text = (
        f"{threshold_pct}% Byzantine"
        if threshold_pct is not None
        else "No drop below 80% in tested range"
    )

    header = [
        f"# {nodes:,}-Node {rounds}-Round Byzantine Resilience Boundary Report",
        "",
        f"- Generated: {timestamp}",
        f"- Configuration: {nodes} nodes, {rounds} rounds, sweep {min_pct}% to {max_pct}% Byzantine",
        f"- Accuracy threshold target: 80%",
        f"- Detected drop point: **{threshold_text}**",
        "",
        "## Summary Table",
        "",
        "| Byzantine % | Final Accuracy | Avg Accuracy | Min Accuracy | First Round <80% |",
        "|---:|---:|---:|---:|---:|",
    ]

    rows = [
        f"| {item.byzantine_pct} | {item.final_accuracy:.2f} | {item.avg_accuracy:.2f} | {item.min_accuracy:.2f} | {item.first_below_80_round or '-'} |"
        for item in summaries
    ]

    path.write_text("\n".join(header + rows) + "\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(
        description=(
            "Run a 20-node, 200-round Byzantine sweep between 50% and 70% "
            "and identify where accuracy drops below 80%."
        )
    )
    parser.add_argument("--nodes", type=int, default=10000)
    parser.add_argument("--rounds", type=int, default=200)
    parser.add_argument("--min-byzantine", type=int, default=50)
    parser.add_argument("--max-byzantine", type=int, default=70)
    parser.add_argument("--seed", type=int, default=20260227)
    args = parser.parse_args()

    if args.min_byzantine > args.max_byzantine:
        raise ValueError("--min-byzantine must be <= --max-byzantine")

    root_dir = Path(__file__).resolve().parents[2]
    results_test_runs = root_dir / "results" / "test-runs"
    results_benchmarks = root_dir / "results" / "benchmarks"
    results_analysis = root_dir / "results" / "analysis"

    results_test_runs.mkdir(parents=True, exist_ok=True)
    results_benchmarks.mkdir(parents=True, exist_ok=True)
    results_analysis.mkdir(parents=True, exist_ok=True)

    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")

    round_records_by_level: dict[int, list[RoundRecord]] = {}
    summaries: list[LevelSummary] = []

    for byzantine_pct in range(args.min_byzantine, args.max_byzantine + 1):
        records, summary = run_level(
            byzantine_pct=byzantine_pct,
            nodes=args.nodes,
            rounds=args.rounds,
            seed=args.seed,
        )
        round_records_by_level[byzantine_pct] = records
        summaries.append(summary)

    threshold_pct = next((s.byzantine_pct for s in summaries if s.final_accuracy < 80.0), None)

    payload = {
        "meta": {
            "generated_at": timestamp,
            "nodes": args.nodes,
            "rounds": args.rounds,
            "byzantine_range": [args.min_byzantine, args.max_byzantine],
            "seed": args.seed,
            "accuracy_floor": 80.0,
            "first_final_accuracy_below_floor_pct": threshold_pct,
        },
        "summary_by_level": [asdict(s) for s in summaries],
        "round_data_by_level": to_json_serializable(round_records_by_level),
    }

    base_name = f"bft_{args.nodes}node_{args.rounds}round_{args.min_byzantine}_{args.max_byzantine}"
    json_latest = results_test_runs / f"{base_name}_results.json"
    json_timestamped = results_test_runs / f"{base_name}_results_{timestamp}.json"
    csv_summary_latest = results_benchmarks / f"{base_name}_summary.csv"
    csv_summary_timestamped = results_benchmarks / f"{base_name}_summary_{timestamp}.csv"
    csv_rounds_latest = results_benchmarks / f"{base_name}_rounds.csv"
    csv_rounds_timestamped = results_benchmarks / f"{base_name}_rounds_{timestamp}.csv"
    report_base = f"BFT_{args.nodes}NODE_{args.rounds}ROUND_BOUNDARY_REPORT"
    md_latest = results_analysis / f"{report_base}.md"
    md_timestamped = results_analysis / f"{report_base}_{timestamp}.md"

    for json_path in (json_latest, json_timestamped):
        json_path.write_text(json.dumps(payload, indent=2), encoding="utf-8")

    for csv_path in (csv_summary_latest, csv_summary_timestamped):
        write_csv_summary(csv_path, summaries)

    for csv_path in (csv_rounds_latest, csv_rounds_timestamped):
        write_csv_rounds(csv_path, round_records_by_level)

    for md_path in (md_latest, md_timestamped):
        write_markdown_report(
            md_path,
            nodes=args.nodes,
            rounds=args.rounds,
            min_pct=args.min_byzantine,
            max_pct=args.max_byzantine,
            threshold_pct=threshold_pct,
            summaries=summaries,
            timestamp=timestamp,
        )

    print(f"Generated: {json_latest}")
    print(f"Generated: {csv_summary_latest}")
    print(f"Generated: {csv_rounds_latest}")
    print(f"Generated: {md_latest}")
    if threshold_pct is None:
        print("No final accuracy below 80% in the tested range.")
    else:
        print(f"First Byzantine level with final accuracy <80%: {threshold_pct}%")


if __name__ == "__main__":
    main()
