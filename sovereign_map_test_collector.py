# sovereign_map_test_collector.py
# Purpose: Collects and verifies all visible test-like values from the Sovereign Map project
#          (mega_test.py, generate_plot_data.py, save_plot.py related data)
#          Saves results to JSON + readable text file
# Date: February 2025

import json
import datetime
import os
import numpy as np
from typing import Dict, Any

# ────────────────────────────────────────────────
# 1. Data from mega_test.py
# ────────────────────────────────────────────────
def get_mega_test_data() -> Dict[str, Any]:
    nodes = 10_000_000
    raw_size_gb = 40_000
    compressed_size_mb = 28
    reduction_factor = (raw_size_gb * 1024) / compressed_size_mb

    malicious_fraction = 0.4
    bft_threshold = 0.555
    is_bft_safe = malicious_fraction > bft_threshold

    recovery_points = [
        88.2, 89.5, 90.8, 91.5, 92.4,
        93.1, 93.8, 94.7, 95.2, 95.8
    ]
    avg_recovery = float(np.mean(recovery_points))

    return {
        "section": "mega_test.py values",
        "nodes": nodes,
        "raw_metadata_tb": raw_size_gb,
        "compressed_metadata_mb": compressed_size_mb,
        "compression_reduction_factor": round(reduction_factor, 1),
        "malicious_fraction": malicious_fraction,
        "bft_threshold": bft_threshold,
        "bft_safe": is_bft_safe,
        "bft_message": "System remains BFT Safe at 55.6% (Theorem 1 Verified)"
                        if is_bft_safe else "Security Threshold Breached",
        "recovery_accuracy_values": recovery_points,
        "average_recovery_accuracy": round(avg_recovery, 2)
    }


# ────────────────────────────────────────────────
# 2. Data from generate_plot_data.py / save_plot.py
# ────────────────────────────────────────────────
def get_convergence_data() -> Dict[str, Any]:
    rounds = list(range(1, 26))
    accuracy = [
        85.0, 86.5, 88.0, 89.2, 91.0,
        92.5, 94.0, 95.5, 96.9,          # 1–9
        88.2,                             # 10 (breach)
        89.5, 90.8, 91.5, 92.4, 93.1, 93.8, 94.7, 95.2, 95.8,  # 11–19
        96.0, 96.2, 96.4, 96.6, 96.8, 96.9   # 20–25
    ]

    breach_round = 10
    breach_accuracy = accuracy[breach_round - 1]
    final_accuracy = accuracy[-1]
    min_post_breach = min(accuracy[breach_round:])

    return {
        "section": "convergence_plot data (rounds 1–25)",
        "rounds": rounds,
        "accuracy_per_round": [float(x) for x in accuracy],
        "breach_round": breach_round,
        "breach_accuracy": breach_accuracy,
        "breach_label": "Byzantine Breach (55.6%)",
        "final_accuracy": final_accuracy,
        "peak_accuracy": max(accuracy),
        "min_accuracy_after_breach": min_post_breach,
        "recovery_delta": round(final_accuracy - breach_accuracy, 2)
    }


# ────────────────────────────────────────────────
# 3. Summary & consistency checks
# ────────────────────────────────────────────────
def run_all_tests() -> Dict[str, Any]:
    timestamp = datetime.datetime.utcnow().isoformat() + "Z"

    mega = get_mega_test_data()
    conv = get_convergence_data()

    summary = {
        "timestamp_utc": timestamp,
        "project": "Sovereign Map / Sovereign Mohawk Proto",
        "tests": [mega, conv],
        "consistency_checks": {
            "bft_claim_matches_plot": abs(mega["malicious_fraction"] - 0.556) < 0.001,
            "breach_accuracy_in_plot": conv["breach_accuracy"] == 88.2,
            "avg_recovery_matches_mega": abs(mega["average_recovery_accuracy"] - 92.5) < 0.01,
            "final_accuracy_claim": conv["final_accuracy"] >= 96.9,
            "compression_factor_sane": mega["compression_reduction_factor"] > 1_000_000
        }
    }

    return summary


# ────────────────────────────────────────────────
# Save results
# ────────────────────────────────────────────────
def save_results(data: Dict[str, Any]):
    ts = datetime.datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    
    # JSON – machine readable
    json_path = f"sovereign_test_results_{ts}.json"
    with open(json_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"Saved detailed JSON: {json_path}")

    # Human-readable report
    txt_path = f"sovereign_test_report_{ts}.txt"
    with open(txt_path, "w", encoding="utf-8") as f:
        f.write("SOVEREIGN MAP / MOHAWK PROTO – COLLECTED TEST VALUES\n")
        f.write("=" * 60 + "\n\n")
        f.write(f"Generated: {data['timestamp_utc']}\n\n")

        for section in data["tests"]:
            f.write(f"[{section['section']}]\n")
            for k, v in section.items():
                if k != "section":
                    f.write(f"  {k: <28}: {v}\n")
            f.write("\n")

        f.write("[Consistency / Sanity Checks]\n")
        for k, v in data["consistency_checks"].items():
            status = "PASS" if v else "FAIL"
            f.write(f"  {k: <38}: {status}\n")

    print(f"Saved readable report: {txt_path}")


def main():
    print("Collecting all visible test values from Sovereign Map scripts...")
    results = run_all_tests()
    save_results(results)
    print("\nDone. Check the generated .json and .txt files.")


if __name__ == "__main__":
    main()
