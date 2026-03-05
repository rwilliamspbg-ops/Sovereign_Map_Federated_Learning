#!/usr/bin/env python3
"""
1000-Node NPU Performance Analysis
Generates plots and metrics from test results
"""

import json
import sys
import os
from pathlib import Path
from typing import Dict, List, Any


def load_json(filepath: str) -> Dict[str, Any]:
    """Load JSON file safely"""
    try:
        with open(filepath) as f:
            return json.load(f)
    except Exception as e:
        print(f"⚠️  Failed to load {filepath}: {e}")
        return {}


def generate_npu_comparison_plots(artifacts_dir: str, plots_dir: str):
    """Generate NPU CPU vs Accelerated comparison plots"""
    try:
        import matplotlib.pyplot as plt
        import numpy as np
    except ImportError:
        print("⚠️  matplotlib not available, skipping plots")
        return

    cpu_data = load_json(f"{artifacts_dir}/npu_baseline_cpu.json")
    npu_data = load_json(f"{artifacts_dir}/npu_accelerated.json")

    if not (cpu_data and npu_data):
        print("⚠️  NPU test data not available")
        return

    fig, axes = plt.subplots(2, 2, figsize=(15, 10))
    fig.suptitle("1000-Node NPU Performance Analysis", fontsize=16, fontweight="bold")

    # Plot 1: Throughput Comparison
    scenarios = ["CPU Baseline", "NPU Accelerated"]
    throughput = [cpu_data.get("throughput_rps", 0), npu_data.get("throughput_rps", 0)]
    colors = ["#FF6B6B", "#4ECDC4"]
    ax = axes[0, 0]
    bars = ax.bar(
        scenarios, throughput, color=colors, alpha=0.7, edgecolor="black", linewidth=2
    )
    ax.set_ylabel("Requests Per Second", fontweight="bold", fontsize=11)
    ax.set_title("Throughput: CPU vs NPU", fontweight="bold", fontsize=12)
    ax.grid(axis="y", alpha=0.3)

    if throughput[1] > 0 and throughput[0] > 0:
        speedup = throughput[1] / throughput[0]
        ax.text(
            0.5,
            max(throughput) * 0.95,
            f"{speedup:.2f}x speedup",
            ha="center",
            fontsize=11,
            fontweight="bold",
            bbox=dict(boxstyle="round", facecolor="yellow", alpha=0.7),
        )

    for i, v in enumerate(throughput):
        ax.text(
            i, v + max(throughput) * 0.02, f"{v:.0f}", ha="center", fontweight="bold"
        )

    # Plot 2: Latency Comparison
    latency = [cpu_data.get("avg_latency_ms", 0), npu_data.get("avg_latency_ms", 0)]
    ax = axes[0, 1]
    bars = ax.bar(
        scenarios, latency, color=colors, alpha=0.7, edgecolor="black", linewidth=2
    )
    ax.set_ylabel("Average Latency (ms)", fontweight="bold", fontsize=11)
    ax.set_title("Latency: CPU vs NPU", fontweight="bold", fontsize=12)
    ax.grid(axis="y", alpha=0.3)

    if latency[0] > 0 and latency[1] > 0:
        improvement = (latency[0] - latency[1]) / latency[0] * 100
        ax.text(
            0.5,
            max(latency) * 0.95,
            f"{improvement:.1f}% faster",
            ha="center",
            fontsize=11,
            fontweight="bold",
            bbox=dict(boxstyle="round", facecolor="lightgreen", alpha=0.7),
        )

    for i, v in enumerate(latency):
        ax.text(
            i, v + max(latency) * 0.02, f"{v:.2f}ms", ha="center", fontweight="bold"
        )

    # Plot 3: Resource Utilization
    ax = axes[1, 0]
    metrics = ["CPU\nUsage", "Memory\nUsage", "GPU\nMemory"]
    cpu_vals = [
        cpu_data.get("cpu_percent", 0),
        cpu_data.get("memory_percent", 0),
        cpu_data.get("gpu_memory_percent", 0),
    ]
    npu_vals = [
        npu_data.get("cpu_percent", 0),
        npu_data.get("memory_percent", 0),
        npu_data.get("gpu_memory_percent", 0),
    ]

    x = np.arange(len(metrics))
    width = 0.35
    ax.bar(
        x - width / 2,
        cpu_vals,
        width,
        label="CPU",
        alpha=0.7,
        color="#FF6B6B",
        edgecolor="black",
    )
    ax.bar(
        x + width / 2,
        npu_vals,
        width,
        label="NPU",
        alpha=0.7,
        color="#4ECDC4",
        edgecolor="black",
    )
    ax.set_ylabel("Usage (%)", fontweight="bold", fontsize=11)
    ax.set_title("Resource Utilization", fontweight="bold", fontsize=12)
    ax.set_xticks(x)
    ax.set_xticklabels(metrics)
    ax.legend()
    ax.grid(axis="y", alpha=0.3)

    # Plot 4: Metrics Summary Table
    ax = axes[1, 1]
    ax.axis("off")

    summary_data = [
        ["Metric", "CPU Baseline", "NPU Accelerated", "Improvement"],
        [
            "Throughput (RPS)",
            f"{cpu_data.get('throughput_rps', 0):.0f}",
            f"{npu_data.get('throughput_rps', 0):.0f}",
            f"{((npu_data.get('throughput_rps', 0) / max(cpu_data.get('throughput_rps', 1), 0.01)) - 1) * 100:.1f}%",
        ],
        [
            "Latency (ms)",
            f"{cpu_data.get('avg_latency_ms', 0):.2f}",
            f"{npu_data.get('avg_latency_ms', 0):.2f}",
            f"{((cpu_data.get('avg_latency_ms', 1) - npu_data.get('avg_latency_ms', 0)) / max(cpu_data.get('avg_latency_ms', 1), 0.01)) * 100:.1f}%",
        ],
        [
            "CPU (%)",
            f"{cpu_data.get('cpu_percent', 0):.1f}",
            f"{npu_data.get('cpu_percent', 0):.1f}",
            f"{((cpu_data.get('cpu_percent', 1) - npu_data.get('cpu_percent', 0)) / max(cpu_data.get('cpu_percent', 1), 0.01)) * 100:.1f}%",
        ],
    ]

    table = ax.table(
        cellText=summary_data,
        cellLoc="center",
        loc="center",
        colWidths=[0.25, 0.25, 0.25, 0.25],
    )
    table.auto_set_font_size(False)
    table.set_fontsize(10)
    table.scale(1, 2)

    # Style header row
    for i in range(4):
        table[(0, i)].set_facecolor("#4ECDC4")
        table[(0, i)].set_text_props(weight="bold", color="white")

    # Style data rows
    for i in range(1, len(summary_data)):
        for j in range(4):
            if i % 2 == 0:
                table[(i, j)].set_facecolor("#f0f0f0")

    plt.tight_layout()
    plt.savefig(
        f"{plots_dir}/01-npu-performance-analysis.png", dpi=300, bbox_inches="tight"
    )
    print(f"✓ Generated: 01-npu-performance-analysis.png")
    plt.close()


def generate_throughput_plots(artifacts_dir: str, plots_dir: str):
    """Generate detailed throughput analysis plots"""
    try:
        import matplotlib.pyplot as plt
        import numpy as np
    except ImportError:
        return

    throughput_data = load_json(f"{artifacts_dir}/throughput_test.json")
    if not throughput_data or "latency" not in throughput_data:
        return

    fig, axes = plt.subplots(2, 2, figsize=(15, 10))
    fig.suptitle(
        "1000-Node Throughput & Latency Analysis", fontsize=16, fontweight="bold"
    )

    latencies = [l * 1000 for l in throughput_data["latency"]]  # Convert to ms

    # Plot 1: Latency Histogram
    ax = axes[0, 0]
    ax.hist(
        latencies, bins=50, color="#3498db", alpha=0.7, edgecolor="black", linewidth=1.5
    )
    ax.set_xlabel("Latency (ms)", fontweight="bold", fontsize=11)
    ax.set_ylabel("Frequency", fontweight="bold", fontsize=11)
    ax.set_title("Latency Distribution (1000 Requests)", fontweight="bold", fontsize=12)
    ax.axvline(
        np.mean(latencies),
        color="red",
        linestyle="--",
        linewidth=2,
        label=f"Mean: {np.mean(latencies):.2f}ms",
    )
    ax.axvline(
        np.median(latencies),
        color="green",
        linestyle="--",
        linewidth=2,
        label=f"Median: {np.median(latencies):.2f}ms",
    )
    ax.legend()
    ax.grid(alpha=0.3)

    # Plot 2: CDF
    ax = axes[0, 1]
    sorted_latencies = sorted(latencies)
    cdf = np.arange(1, len(sorted_latencies) + 1) / len(sorted_latencies)
    ax.plot(sorted_latencies, cdf, linewidth=2.5, color="#e74c3c")
    ax.set_xlabel("Latency (ms)", fontweight="bold", fontsize=11)
    ax.set_ylabel("Cumulative Probability", fontweight="bold", fontsize=12)
    ax.set_title("Latency CDF", fontweight="bold", fontsize=12)
    ax.grid(alpha=0.3)

    # Plot 3: Percentiles
    ax = axes[1, 0]
    percentiles = [50, 75, 90, 95, 99]
    perc_values = [np.percentile(latencies, p) for p in percentiles]
    colors_perc = ["#2ecc71", "#f39c12", "#e74c3c", "#c0392b", "#8b0000"]
    bars = ax.bar(
        [f"p{p}" for p in percentiles],
        perc_values,
        color=colors_perc,
        alpha=0.7,
        edgecolor="black",
        linewidth=1.5,
    )
    ax.set_ylabel("Latency (ms)", fontweight="bold", fontsize=11)
    ax.set_title("Latency Percentiles", fontweight="bold", fontsize=12)
    ax.grid(axis="y", alpha=0.3)

    for bar, val in zip(bars, perc_values):
        height = bar.get_height()
        ax.text(
            bar.get_x() + bar.get_width() / 2.0,
            height,
            f"{val:.2f}ms",
            ha="center",
            va="bottom",
            fontweight="bold",
            fontsize=9,
        )

    # Plot 4: Summary Statistics
    ax = axes[1, 1]
    ax.axis("off")

    stats_text = f"""
THROUGHPUT & LATENCY STATISTICS

Total Requests:        {throughput_data.get('total_requests', 0)}
Throughput:            {throughput_data.get('throughput_rps', 0):.2f} RPS
Total Duration:        {throughput_data.get('total_time_seconds', 0):.2f} seconds

LATENCY METRICS
Minimum:               {np.min(latencies):.3f} ms
Maximum:               {np.max(latencies):.3f} ms
Average (Mean):        {np.mean(latencies):.3f} ms
Median (p50):          {np.percentile(latencies, 50):.3f} ms
p75:                   {np.percentile(latencies, 75):.3f} ms
p90:                   {np.percentile(latencies, 90):.3f} ms
p95:                   {np.percentile(latencies, 95):.3f} ms
p99:                   {np.percentile(latencies, 99):.3f} ms

VARIABILITY
Std Dev:               {np.std(latencies):.3f} ms
Variance:              {np.var(latencies):.3f} ms²
Coefficient of Var:    {(np.std(latencies) / np.mean(latencies)):.3f}
"""

    ax.text(
        0.05,
        0.95,
        stats_text,
        transform=ax.transAxes,
        fontsize=10,
        verticalalignment="top",
        family="monospace",
        bbox=dict(boxstyle="round", facecolor="wheat", alpha=0.8),
    )

    plt.tight_layout()
    plt.savefig(
        f"{plots_dir}/02-throughput-latency-analysis.png", dpi=300, bbox_inches="tight"
    )
    print(f"✓ Generated: 02-throughput-latency-analysis.png")
    plt.close()


def generate_consensus_plots(artifacts_dir: str, plots_dir: str):
    """Generate consensus efficiency plots"""
    try:
        import matplotlib.pyplot as plt
        import numpy as np
    except ImportError:
        return

    consensus_data = load_json(f"{artifacts_dir}/consensus_efficiency.json")
    if not consensus_data:
        return

    fig, axes = plt.subplots(2, 2, figsize=(15, 10))
    fig.suptitle(
        "1000-Node Consensus Efficiency Analysis", fontsize=16, fontweight="bold"
    )

    # Plot 1: Messages per Round
    ax = axes[0, 0]
    rounds = consensus_data.get("rounds", list(range(1, 11)))
    msg_counts = consensus_data.get("message_counts", [0] * len(rounds))
    ax.plot(
        rounds, msg_counts, marker="o", linewidth=2.5, markersize=8, color="#3498db"
    )
    ax.set_xlabel("Consensus Round", fontweight="bold", fontsize=11)
    ax.set_ylabel("Messages Sent", fontweight="bold", fontsize=11)
    ax.set_title("Messages Per Consensus Round", fontweight="bold", fontsize=12)
    ax.grid(alpha=0.3)

    # Plot 2: Success Rate
    ax = axes[0, 1]
    success_rate = consensus_data.get("success_rate", 0) * 100
    failure_rate = 100 - success_rate
    colors = ["#2ecc71", "#e74c3c"]
    wedges, texts, autotexts = ax.pie(
        [success_rate, failure_rate],
        labels=["Success", "Failure"],
        autopct="%1.1f%%",
        colors=colors,
        startangle=90,
        textprops={"fontweight": "bold", "fontsize": 11},
    )
    ax.set_title("Consensus Success Rate", fontweight="bold", fontsize=12)

    # Plot 3: Round Duration
    ax = axes[1, 0]
    round_durations = consensus_data.get("round_durations", [0] * len(rounds))
    ax.bar(
        rounds,
        round_durations,
        color="#9b59b6",
        alpha=0.7,
        edgecolor="black",
        linewidth=1.5,
    )
    ax.set_xlabel("Consensus Round", fontweight="bold", fontsize=11)
    ax.set_ylabel("Duration (seconds)", fontweight="bold", fontsize=11)
    ax.set_title("Consensus Round Duration", fontweight="bold", fontsize=12)
    ax.grid(axis="y", alpha=0.3)

    # Plot 4: Summary
    ax = axes[1, 1]
    ax.axis("off")

    summary = f"""
CONSENSUS EFFICIENCY METRICS

Total Rounds:          {consensus_data.get('total_rounds', 0)}
Successful Rounds:     {consensus_data.get('successful_rounds', 0)}
Failed Rounds:         {consensus_data.get('failed_rounds', 0)}
Success Rate:          {success_rate:.1f}%

MESSAGE EFFICIENCY
Avg Messages/Round:    {np.mean(msg_counts) if msg_counts else 0:.0f}
Min Messages:          {min(msg_counts) if msg_counts else 0}
Max Messages:          {max(msg_counts) if msg_counts else 0}
Total Messages:        {sum(msg_counts) if msg_counts else 0}

TIMING
Avg Round Duration:    {np.mean(round_durations) if round_durations else 0:.2f}s
Min Round Duration:    {min(round_durations) if round_durations else 0:.2f}s
Max Round Duration:    {max(round_durations) if round_durations else 0:.2f}s
Total Time:            {consensus_data.get('total_duration', 0):.2f}s

BYZANTINE RESILIENCE
Byzantine Nodes:       {consensus_data.get('byzantine_nodes', 0)}
Byzantine Detected:    {consensus_data.get('byzantine_detected', 0)}
Detection Rate:        {(consensus_data.get('byzantine_detected', 0) / max(consensus_data.get('byzantine_nodes', 1), 1)) * 100:.1f}%
"""

    ax.text(
        0.05,
        0.95,
        summary,
        transform=ax.transAxes,
        fontsize=10,
        verticalalignment="top",
        family="monospace",
        bbox=dict(boxstyle="round", facecolor="lightblue", alpha=0.8),
    )

    plt.tight_layout()
    plt.savefig(
        f"{plots_dir}/03-consensus-efficiency-analysis.png",
        dpi=300,
        bbox_inches="tight",
    )
    print(f"✓ Generated: 03-consensus-efficiency-analysis.png")
    plt.close()


def main():
    if len(sys.argv) < 3:
        print("Usage: python3 generate-plots.py <artifacts_dir> <plots_dir>")
        sys.exit(1)

    artifacts_dir = sys.argv[1]
    plots_dir = sys.argv[2]

    Path(plots_dir).mkdir(parents=True, exist_ok=True)

    print("📊 Generating visualization plots...")
    print("")

    generate_npu_comparison_plots(artifacts_dir, plots_dir)
    generate_throughput_plots(artifacts_dir, plots_dir)
    generate_consensus_plots(artifacts_dir, plots_dir)

    print("")
    print("✅ Plot generation complete")


if __name__ == "__main__":
    main()
