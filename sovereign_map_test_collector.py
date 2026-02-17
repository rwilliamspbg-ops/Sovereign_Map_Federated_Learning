# sovereign_map_test_collector.py
# Now reads from mega_test output instead of hard-coding everything

import json
import datetime
import os
import argparse
import subprocess
import numpy as np
import matplotlib.pyplot as plt

def load_mega_results(path: str) -> dict:
    if not os.path.exists(path):
        raise FileNotFoundError(f"Mega test output not found: {path}")
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def generate_report(mega_data: dict):
    timestamp = datetime.datetime.utcnow().isoformat() + "Z"

    conv = {
        "section": "convergence_plot data (rounds 1–25)",
        "rounds": list(range(1, len(mega_data["accuracy_per_round"]) + 1)),
        "accuracy_per_round": mega_data["accuracy_per_round"],
        "breach_round": mega_data["breach_round"],
        "breach_accuracy": mega_data["breach_accuracy"],
        "breach_label": "Byzantine Breach (55.6%)",
        "final_accuracy": mega_data["final_accuracy"],
        "peak_accuracy": mega_data["peak_accuracy"],
        "min_accuracy_after_breach": mega_data["min_accuracy_after_breach"],
        "recovery_delta": round(mega_data["final_accuracy"] - mega_data["breach_accuracy"], 2)
    }

    summary = {
        "timestamp_utc": timestamp,
        "project": "Sovereign Map / Sovereign Mohawk Proto",
        "tests": [
            {
                "section": "mega_test.py values",
                **{k: v for k, v in mega_data.items() if k not in [
                    "accuracy_per_round", "timestamp", "breach_round",
                    "breach_accuracy", "final_accuracy", "peak_accuracy",
                    "min_accuracy_after_breach", "recovery_delta"
                ]}
            },
            conv
        ],
        "consistency_checks": {
            "bft_claim_matches_plot": mega_data["malicious_fraction"] >= 0.555,
            "breach_accuracy_in_plot": abs(mega_data["breach_accuracy"] - 88.2) < 1.0,
            "avg_recovery_matches_mega": abs(mega_data["average_recovery_accuracy"] - 92.5) < 1.0,
            "final_accuracy_claim": mega_data["final_accuracy"] >= 96.0,
            "compression_factor_sane": mega_data["compression_reduction_factor"] > 1_000_000
        }
    }

    return summary, conv


def save_report(summary: dict, malicious: float):
    ts = datetime.datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    os.makedirs("audit_results", exist_ok=True)

    base = f"sovereign_test_{ts}_mal{int(malicious*100)}"

    json_path = f"audit_results/{base}.json"
    with open(json_path, "w", encoding="utf-8") as f:
        json.dump(summary, f, indent=2)

    txt_path = f"audit_results/{base}.txt"
    with open(txt_path, "w", encoding="utf-8") as f:
        f.write("SOVEREIGN MAP / MOHAWK PROTO – COLLECTED TEST VALUES\n")
        f.write("="*60 + "\n\n")
        f.write(f"Generated: {summary['timestamp_utc']}\n\n")
        for sec in summary["tests"]:
            f.write(f"[{sec['section']}]\n")
            for k, v in sec.items():
                if k != "section":
                    f.write(f"  {k:<28}: {v}\n")
            f.write("\n")
        f.write("[Consistency / Sanity Checks]\n")
        for k, v in summary["consistency_checks"].items():
            f.write(f"  {k:<38}: {'PASS' if v else 'FAIL'}\n")

    return json_path, txt_path


def plot_convergence(conv: dict, malicious: float):
    ts = datetime.datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    path = f"audit_results/convergence_{ts}_mal{int(malicious*100)}.png"

    plt.figure(figsize=(10,6))
    plt.plot(conv["rounds"], conv["accuracy_per_round"], "b-o", label="Accuracy")
    plt.axvline(conv["breach_round"], color="r", ls="--", label=conv["breach_label"])
    plt.title(f"Convergence – malicious = {malicious:.2%}")
    plt.xlabel("Round")
    plt.ylabel("Accuracy (%)")
    plt.grid(True)
    plt.legend()
    plt.savefig(path, dpi=150, bbox_inches="tight")
    plt.close()
    return path


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--malicious", type=float, default=0.55)
    parser.add_argument("--mega-output", default="mega_test_output.json")
    parser.add_argument("--run-mega", action="store_true", help="Run mega_test.py first")
    args = parser.parse_args()

    if args.run_mega:
        print("Running mega_test simulation...")
        subprocess.run([
            "python", "mega_test.py",
            "--malicious", str(args.malicious),
            "--output", args.mega_output
        ], check=True)

    print("Loading results...")
    mega = load_mega_results(args.mega_output)
    report, conv = generate_report(mega)
    json_p, txt_p = save_report(report, args.malicious)
    plot_p = plot_convergence(conv, args.malicious)

    print("\nGenerated:")
    print(f"  Report JSON: {json_p}")
    print(f"  Report TXT : {txt_p}")
    print(f"  Plot       : {plot_p}")


if __name__ == "__main__":
    main()
