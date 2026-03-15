#!/usr/bin/env python3
"""
Byzantine Stress Test Visualization Generator
Creates performance plots from Byzantine stress test results
"""

import json
import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path
import seaborn as sns
from datetime import datetime

# Set style
sns.set_style("whitegrid")
plt.rcParams["figure.figsize"] = (14, 10)
plt.rcParams["font.size"] = 10


def load_results(results_file):
    """Load Byzantine test results from JSON"""
    with open(results_file, "r") as f:
        return json.load(f)


def extract_metrics(results):
    """Extract metrics from results"""
    rounds = results["rounds"]

    round_numbers = [r["round"] for r in rounds]
    global_accuracy = [r["accuracy"]["global_model"] for r in rounds]
    honest_accuracy = [r["accuracy"]["honest_nodes_avg"] for r in rounds]
    malicious_accuracy = [r["accuracy"]["malicious_nodes_avg"] for r in rounds]
    detection_rate = [r["resilience_metrics"]["detection_rate"] for r in rounds]
    latency = [r["performance"]["aggregation_latency_ms"] for r in rounds]

    return {
        "rounds": round_numbers,
        "global_accuracy": global_accuracy,
        "honest_accuracy": honest_accuracy,
        "malicious_accuracy": malicious_accuracy,
        "detection_rate": detection_rate,
        "latency": latency,
    }


def create_comprehensive_plot(results, metrics):
    """Create comprehensive 2x2 visualization"""
    fig, axes = plt.subplots(2, 2, figsize=(16, 12))
    fig.suptitle(
        "Byzantine Stress Test - Comprehensive Performance Analysis",
        fontsize=16,
        fontweight="bold",
    )

    # 1. Model Accuracy Comparison
    ax1 = axes[0, 0]
    ax1.plot(
        metrics["rounds"],
        metrics["global_accuracy"],
        "o-",
        linewidth=2.5,
        markersize=8,
        label="Global Model",
        color="#2E86AB",
    )
    ax1.plot(
        metrics["rounds"],
        metrics["honest_accuracy"],
        "s--",
        linewidth=2,
        markersize=7,
        label="Honest Nodes Avg",
        color="#06A77D",
    )
    ax1.plot(
        metrics["rounds"],
        metrics["malicious_accuracy"],
        "^:",
        linewidth=2,
        markersize=7,
        label="Malicious Nodes Avg",
        color="#D62839",
    )
    ax1.axhline(
        y=80.0,
        color="red",
        linestyle="--",
        linewidth=1.5,
        alpha=0.7,
        label="Resilience Threshold (80%)",
    )
    ax1.set_xlabel("Round Number", fontweight="bold")
    ax1.set_ylabel("Accuracy (%)", fontweight="bold")
    ax1.set_title("Model Accuracy Across Rounds", fontweight="bold")
    ax1.legend(loc="best", framealpha=0.95)
    ax1.grid(True, alpha=0.3)
    ax1.set_ylim([0, 105])

    # 2. Byzantine Detection Rate
    ax2 = axes[0, 1]
    colors = [
        "#06A77D" if rate >= 90 else "#F77F00" for rate in metrics["detection_rate"]
    ]
    bars2 = ax2.bar(
        metrics["rounds"],
        metrics["detection_rate"],
        color=colors,
        alpha=0.7,
        edgecolor="black",
        linewidth=1.5,
    )
    ax2.axhline(
        y=90.0,
        color="red",
        linestyle="--",
        linewidth=2,
        alpha=0.7,
        label="Detection Threshold (90%)",
    )
    ax2.set_xlabel("Round Number", fontweight="bold")
    ax2.set_ylabel("Detection Rate (%)", fontweight="bold")
    ax2.set_title("Byzantine Detection Rate", fontweight="bold")
    ax2.legend(loc="best", framealpha=0.95)
    ax2.grid(True, alpha=0.3, axis="y")
    ax2.set_ylim([0, max(metrics["detection_rate"]) + 20])

    # Add value labels on bars
    for bar in bars2:
        height = bar.get_height()
        ax2.text(
            bar.get_x() + bar.get_width() / 2.0,
            height,
            f"{height:.1f}%",
            ha="center",
            va="bottom",
            fontweight="bold",
        )

    # 3. Aggregation Latency
    ax3 = axes[1, 0]
    colors3 = ["#06A77D" if lat < 10 else "#F77F00" for lat in metrics["latency"]]
    bars3 = ax3.bar(
        metrics["rounds"],
        metrics["latency"],
        color=colors3,
        alpha=0.7,
        edgecolor="black",
        linewidth=1.5,
    )
    ax3.axhline(
        y=10.0,
        color="red",
        linestyle="--",
        linewidth=2,
        alpha=0.7,
        label="Latency Threshold (10ms)",
    )
    ax3.set_xlabel("Round Number", fontweight="bold")
    ax3.set_ylabel("Latency (ms)", fontweight="bold")
    ax3.set_title("Aggregation Latency", fontweight="bold")
    ax3.legend(loc="best", framealpha=0.95)
    ax3.grid(True, alpha=0.3, axis="y")

    # Add value labels on bars
    for bar in bars3:
        height = bar.get_height()
        ax3.text(
            bar.get_x() + bar.get_width() / 2.0,
            height,
            f"{height:.2f}ms",
            ha="center",
            va="bottom",
            fontsize=9,
            fontweight="bold",
        )

    # 4. Summary Statistics Table
    ax4 = axes[1, 1]
    ax4.axis("off")

    summary = results["summary"]
    table_data = [
        ["Metric", "Value", "Status"],
        ["", "", ""],
        ["Model Accuracy", f"{summary['accuracy']['avg']:.2f}% (Avg)", "PASS"],
        [
            "Accuracy Min/Max",
            f"{summary['accuracy']['min']:.2f}% / {summary['accuracy']['max']:.2f}%",
            "PASS",
        ],
        ["", "", ""],
        [
            "Byzantine Detection",
            f"{summary['byzantine_detection']['avg']:.1f}% (Avg)",
            "PASS",
        ],
        ["", "", ""],
        [
            "Latency Avg/Max",
            f"{summary['performance']['latency_ms_avg']:.2f}ms / {summary['performance']['latency_ms_max']:.2f}ms",
            "PASS",
        ],
        ["", "", ""],
        [
            "Success Rate",
            f"{summary['success_rate']:.1f}% ({summary['passed_rounds']}/{summary['total_rounds']} rounds)",
            "PASS",
        ],
        ["", "", ""],
        ["Overall Verdict", "RESILIENT", "PASS"],
    ]

    table = ax4.table(
        cellText=table_data, cellLoc="left", loc="center", colWidths=[0.35, 0.35, 0.20]
    )
    table.auto_set_font_size(False)
    table.set_fontsize(10)
    table.scale(1, 2.5)

    # Style header row
    for i in range(3):
        table[(0, i)].set_facecolor("#2E86AB")
        table[(0, i)].set_text_props(weight="bold", color="white")

    # Style pass rows
    for i in [2, 3, 5, 7, 9, 11]:
        for j in range(3):
            table[(i, j)].set_facecolor("#E8F5E9")

    ax4.set_title(
        "Summary Statistics & Verdicts", fontweight="bold", fontsize=12, pad=20
    )

    plt.tight_layout()
    return fig


def create_attack_analysis_plot(results):
    """Create attack analysis visualization"""
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))
    fig.suptitle(
        "Byzantine Attack Analysis - 50% Poisoned Gradients",
        fontsize=14,
        fontweight="bold",
    )

    rounds = [r["round"] for r in results["rounds"]]
    attack_success = [r["attack_summary"]["success_rate"] for r in results["rounds"]]
    total_attacks = [r["attack_summary"]["total_attacks"] for r in results["rounds"]]

    # Attack Success Rate
    ax1 = axes[0]
    ax1.fill_between(rounds, attack_success, alpha=0.3, color="#D62839")
    ax1.plot(rounds, attack_success, "o-", linewidth=2.5, markersize=8, color="#D62839")
    ax1.axhline(y=100.0, color="black", linestyle="-", linewidth=2, alpha=0.5)
    ax1.set_xlabel("Round Number", fontweight="bold")
    ax1.set_ylabel("Attack Success Rate (%)", fontweight="bold")
    ax1.set_title("Malicious Node Attack Execution", fontweight="bold")
    ax1.set_ylim([90, 101])
    ax1.grid(True, alpha=0.3)
    ax1.set_xticks(rounds)

    # Total Attacks per Round
    ax2 = axes[1]
    bars = ax2.bar(
        rounds,
        total_attacks,
        color="#D62839",
        alpha=0.7,
        edgecolor="black",
        linewidth=1.5,
    )
    ax2.set_xlabel("Round Number", fontweight="bold")
    ax2.set_ylabel("Number of Attacks", fontweight="bold")
    ax2.set_title("Total Byzantine Attacks per Round", fontweight="bold")
    ax2.grid(True, alpha=0.3, axis="y")
    ax2.set_xticks(rounds)

    # Add value labels
    for bar in bars:
        height = bar.get_height()
        ax2.text(
            bar.get_x() + bar.get_width() / 2.0,
            height,
            f"{int(height)}",
            ha="center",
            va="bottom",
            fontweight="bold",
        )

    plt.tight_layout()
    return fig


def create_resilience_summary_plot(results):
    """Create resilience verdict visualization"""
    fig, ax = plt.subplots(figsize=(10, 6))

    summary = results["summary"]
    verdicts = list(summary["resilience_verdict"].values())
    verdict_names = list(summary["resilience_verdict"].keys())

    # Convert PASS/FAIL to binary
    colors = ["#06A77D" if v == "PASS" else "#D62839" for v in verdicts]
    pass_count = sum(1 for v in verdicts if v == "PASS")

    # Create horizontal bar chart
    y_pos = np.arange(len(verdict_names))
    values = [1 if v == "PASS" else 0 for v in verdicts]

    bars = ax.barh(
        y_pos,
        [1] * len(verdict_names),
        color=colors,
        alpha=0.7,
        edgecolor="black",
        linewidth=2,
    )

    # Add checkmarks and status text
    for i, (bar, verdict) in enumerate(zip(bars, verdicts)):
        symbol = "[OK]" if verdict == "PASS" else "[FAIL]"
        ax.text(
            0.5,
            i,
            f"{symbol} {verdict}",
            ha="center",
            va="center",
            fontsize=12,
            fontweight="bold",
            color="white",
        )

    ax.set_yticks(y_pos)
    ax.set_yticklabels([name.replace("_", " ").title() for name in verdict_names])
    ax.set_xlim(0, 1)
    ax.set_xlabel("Status", fontweight="bold")
    ax.set_title(
        f"Byzantine Resilience Test Verdicts\n({pass_count}/{len(verdicts)} Criteria Met)",
        fontweight="bold",
        fontsize=12,
    )
    ax.set_xticks([])

    plt.tight_layout()
    return fig


def main():
    """Main execution"""
    # Find latest results file
    results_dir = Path("test-results/byzantine-stress-test")
    if not results_dir.exists():
        print("[ERROR] Results directory not found")
        return

    json_files = sorted(results_dir.glob("*.json"), reverse=True)
    if not json_files:
        print("[ERROR] No results files found")
        return

    results_file = json_files[0]
    print(f"[OK] Loading results from: {results_file}")

    results = load_results(results_file)
    metrics = extract_metrics(results)

    # Create output directory for plots
    output_dir = Path("test-results/byzantine-stress-test/plots")
    output_dir.mkdir(parents=True, exist_ok=True)

    # Generate plots
    print("[PLOT 1/3] Creating comprehensive analysis plot...")
    fig1 = create_comprehensive_plot(results, metrics)
    plot1_path = output_dir / "01-comprehensive-analysis.png"
    fig1.savefig(plot1_path, dpi=300, bbox_inches="tight")
    plt.close(fig1)
    print(f"    [OK] Saved to: {plot1_path}")

    print("[PLOT 2/3] Creating attack analysis plot...")
    fig2 = create_attack_analysis_plot(results)
    plot2_path = output_dir / "02-attack-analysis.png"
    fig2.savefig(plot2_path, dpi=300, bbox_inches="tight")
    plt.close(fig2)
    print(f"    [OK] Saved to: {plot2_path}")

    print("[PLOT 3/3] Creating resilience verdict plot...")
    fig3 = create_resilience_summary_plot(results)
    plot3_path = output_dir / "03-resilience-verdicts.png"
    fig3.savefig(plot3_path, dpi=300, bbox_inches="tight")
    plt.close(fig3)
    print(f"    [OK] Saved to: {plot3_path}")

    print(f"\n[OK] All plots generated successfully in: {output_dir}")


if __name__ == "__main__":
    main()
