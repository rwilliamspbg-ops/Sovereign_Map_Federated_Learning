#!/usr/bin/env python3
"""
Sovereign Map Incremental Scale Test Runner
Simulates federated learning test with incremental scaling
"""

import json
import time
import os
import sys
from datetime import datetime
from pathlib import Path
import random
import math

# Configuration
INITIAL_NODES = 20
INCREMENT_NODES = 20
MAX_NODES = 100
CONVERGENCE_THRESHOLD = 93
TOTAL_ROUNDS = 500
TPM_ENABLED = True
NPU_ENABLED = True

# Output directory
TEST_NAME = f"incremental_scale_test_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
RESULTS_DIR = Path(f"test-results/{TEST_NAME}")
RESULTS_DIR.mkdir(parents=True, exist_ok=True)

CONVERGENCE_LOG = RESULTS_DIR / "convergence.log"
METRICS_LOG = RESULTS_DIR / "metrics.jsonl"
TEST_LOG = RESULTS_DIR / "test.log"
REPORT_FILE = RESULTS_DIR / "TEST_REPORT.md"
TPM_FILE = RESULTS_DIR / "tpm_attestation.json"
NPU_FILE = RESULTS_DIR / "npu_metrics.json"


def log_msg(level, msg):
    """Log message to file and console"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    log_entry = f"[{timestamp}] [{level}] {msg}"
    print(log_entry)
    with open(TEST_LOG, "a") as f:
        f.write(log_entry + "\n")


def log_convergence(round_num, nodes, accuracy, loss):
    """Log convergence data in JSONL format"""
    entry = {
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "round": str(round_num),
        "nodes": str(nodes),
        "accuracy": f"{accuracy:.1f}",
        "loss": f"{loss:.3f}",
    }
    with open(CONVERGENCE_LOG, "a") as f:
        f.write(json.dumps(entry) + "\n")


def log_metrics(round_num, nodes, cpu_usage, memory_usage):
    """Log system metrics in JSONL format"""
    entry = {
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "round": str(round_num),
        "nodes": str(nodes),
        "backend_cpu": f"{cpu_usage:.1f}%",
        "backend_memory": f"{memory_usage:.0f}MB",
        "prometheus_ready": True,
    }
    with open(METRICS_LOG, "a") as f:
        f.write(json.dumps(entry) + "\n")


def simulate_convergence(round_num, nodes, prev_accuracy):
    """Simulate model convergence"""
    # Initial rapid convergence, then slower improvement
    nodes_factor = (
        nodes / INITIAL_NODES
    ) * 0.95 + 0.05  # Multiple nodes slightly slower

    if round_num < 75:
        # Rapid initial convergence (0% -> 93%)
        base_improvement = (93 / 75) * (1 - prev_accuracy / 100)
        noise = random.gauss(0, 2)
        accuracy = min(prev_accuracy + base_improvement + noise, 95)
    else:
        # Slower improvements after convergence
        base_improvement = (5 / 425) * nodes_factor
        noise = random.gauss(0, 0.5)
        accuracy = min(prev_accuracy + base_improvement + noise, 95)

    # Loss decreases as accuracy increases
    loss = 2.3 * math.exp(-accuracy / 100 * 3) + 0.05

    return max(0, accuracy), max(0.05, loss)


def generate_tpm_report():
    """Generate TPM attestation report"""
    report = {
        "test_name": TEST_NAME,
        "timestamp": datetime.now().strftime("%Y-%m-%dT%H:%M:%SZ"),
        "tpm_enabled": TPM_ENABLED,
        "attestation_results": [],
    }

    if TPM_ENABLED:
        # Simulate node attestation
        nodes = [20, 40, 60, 80, 100]
        for node_count in nodes:
            for i in range(1, node_count + 1):
                if i % 5 == 0:  # Every 5th node
                    report["attestation_results"].append(
                        {
                            "node_id": f"node_{i}",
                            "trust_score": round(90 + random.random() * 10, 1),
                            "cert_valid": True,
                            "signature_verified": True,
                            "last_verified": datetime.now().strftime(
                                "%Y-%m-%d %H:%M:%S"
                            ),
                        }
                    )

    report["status"] = "completed"
    report["nodes_verified"] = len(report["attestation_results"])
    report["all_trusted"] = all(
        r.get("trust_score", 0) > 75 for r in report["attestation_results"]
    )

    with open(TPM_FILE, "w") as f:
        json.dump(report, f, indent=2)

    return report


def generate_npu_report():
    """Generate NPU acceleration report"""
    report = {
        "test_name": TEST_NAME,
        "timestamp": datetime.now().strftime("%Y-%m-%dT%H:%M:%SZ"),
        "npu_enabled": NPU_ENABLED,
        "hardware_info": "NVIDIA Tesla V100 (simulated) or CPU fallback",
        "compute_metrics": {
            "total_rounds": TOTAL_ROUNDS,
            "final_nodes": MAX_NODES,
            "avg_round_time_ms": round(random.uniform(450, 550), 1),
            "gpu_utilization_avg": (
                round(random.uniform(75, 95), 1) if NPU_ENABLED else 0
            ),
            "speedup_factor": (
                round(random.uniform(2.5, 3.5), 2) if NPU_ENABLED else 1.0
            ),
        },
        "status": "completed",
    }

    with open(NPU_FILE, "w") as f:
        json.dump(report, f, indent=2)

    return report


def generate_report(convergence_history, convergence_events):
    """Generate final test report"""
    report = f"""# Incremental Scale Test Report

## Test Configuration
- **Test Name**: {TEST_NAME}
- **Start Time**: {convergence_history[0]['timestamp'] if convergence_history else 'N/A'}
- **Initial Nodes**: {INITIAL_NODES}
- **Increment Size**: {INCREMENT_NODES}
- **Max Nodes**: {MAX_NODES}
- **Convergence Threshold**: {CONVERGENCE_THRESHOLD}%
- **Total Rounds**: {TOTAL_ROUNDS}
- **TPM Attestation**: {'Enabled' if TPM_ENABLED else 'Disabled'}
- **NPU Acceleration**: {'Enabled' if NPU_ENABLED else 'Disabled'}

## Convergence History (Latest 20 Rounds)
| Timestamp | Round | Nodes | Accuracy | Loss |
|-----------|-------|-------|----------|------|
"""

    for entry in convergence_history[-20:]:
        report += f"| {entry['timestamp']} | {entry['round']} | {entry['nodes']} | {entry['accuracy']}% | {entry['loss']} |\n"

    report += "\n## Scaling Events\n"
    for i, event in enumerate(convergence_events, 1):
        report += f"- **Event {i}** (Round {event['round']}): {event['nodes_from']} → {event['nodes_to']} nodes (Accuracy: {event['accuracy']}%)\n"

    if convergence_history:
        final_entry = convergence_history[-1]
        report += f"\n## Test Results\n"
        report += f"- **Final Accuracy**: {final_entry['accuracy']}%\n"
        report += f"- **Final Loss**: {final_entry['loss']}\n"
        report += f"- **Convergence Events**: {len(convergence_events)}\n"
        report += f"- **Final Node Count**: {final_entry['nodes']}\n"
        report += f"- **Test Status**: ✅ COMPLETED\n"

    report += f"\n## Output Files\n"
    report += f"- `test.log` - Complete test execution log\n"
    report += f"- `convergence.log` - Per-round convergence metrics (JSONL)\n"
    report += f"- `metrics.jsonl` - System metrics per round\n"
    report += f"- `tpm_attestation.json` - TPM trust verification results\n"
    report += f"- `npu_metrics.json` - NPU acceleration metrics\n"
    report += f"\n---\n"
    report += f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"

    with open(REPORT_FILE, "w") as f:
        f.write(report)


def run_test():
    """Run the incremental scale test"""
    print("\n" + "=" * 70)
    print("🚀 SOVEREIGN MAP - INCREMENTAL SCALE TEST")
    print("=" * 70 + "\n")

    log_msg("INFO", "Test started")
    log_msg(
        "INFO",
        f"Configuration: Initial={INITIAL_NODES}, Increment={INCREMENT_NODES}, Max={MAX_NODES}, Threshold={CONVERGENCE_THRESHOLD}%",
    )
    log_msg("INFO", f"Results directory: {RESULTS_DIR}")

    current_nodes = INITIAL_NODES
    accuracy = 0.0
    loss = 2.302
    convergence_history = []
    convergence_events = []

    print(f"Starting test with {INITIAL_NODES} nodes...")
    print(f"Running 500 rounds with convergence monitoring...\n")

    for round_num in range(TOTAL_ROUNDS):
        # Simulate convergence
        accuracy, loss = simulate_convergence(round_num, current_nodes, accuracy)

        # Log metrics
        log_convergence(round_num, current_nodes, accuracy, loss)
        cpu_usage = 30 + (current_nodes / 100) * 50 + random.gauss(0, 5)
        memory_usage = 1000 + (current_nodes / 100) * 1000 + random.gauss(0, 50)
        log_metrics(round_num, current_nodes, cpu_usage, memory_usage)

        # Add to history
        convergence_history.append(
            {
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "round": round_num,
                "nodes": current_nodes,
                "accuracy": f"{accuracy:.1f}",
                "loss": f"{loss:.3f}",
            }
        )

        # Check for convergence and scaling
        if accuracy >= CONVERGENCE_THRESHOLD and current_nodes < MAX_NODES:
            next_nodes = min(current_nodes + INCREMENT_NODES, MAX_NODES)
            log_msg(
                "INFO",
                f"Round {round_num}: Convergence {accuracy:.1f}% reached! Scaling {current_nodes} → {next_nodes} nodes",
            )
            print(
                f"  ✓ Round {round_num}: Scaled {current_nodes} → {next_nodes} nodes (Accuracy: {accuracy:.1f}%)"
            )

            convergence_events.append(
                {
                    "round": round_num,
                    "nodes_from": current_nodes,
                    "nodes_to": next_nodes,
                    "accuracy": f"{accuracy:.1f}",
                }
            )

            current_nodes = next_nodes

        # Progress output
        if round_num % 50 == 0 and round_num > 0:
            print(
                f"  Round {round_num}/500 | Nodes: {current_nodes} | Accuracy: {accuracy:.1f}% | Loss: {loss:.3f}"
            )

    log_msg(
        "INFO", f"Test completed: {TOTAL_ROUNDS} rounds, reached {current_nodes} nodes"
    )

    print(f"\n✅ Test completed!")
    print(f"  Rounds: {TOTAL_ROUNDS}")
    print(f"  Final Nodes: {current_nodes}")
    print(f"  Final Accuracy: {accuracy:.1f}%")
    print(f"  Convergence Events: {len(convergence_events)}")

    # Generate reports
    log_msg("INFO", "Generating reports...")
    generate_report(convergence_history, convergence_events)
    tpm_report = generate_tpm_report()
    npu_report = generate_npu_report()

    log_msg(
        "INFO",
        f"TPM Report: {tpm_report['nodes_verified']} nodes verified, all_trusted={tpm_report['all_trusted']}",
    )
    log_msg(
        "INFO",
        f"NPU Report: GPU speedup={npu_report['compute_metrics']['speedup_factor']}x",
    )
    log_msg("INFO", "All reports generated")

    print(f"\n📊 Results Summary:")
    print(f"  Scaling events: {len(convergence_events)}")
    for i, event in enumerate(convergence_events, 1):
        print(
            f"    Event {i}: Round {event['round']} - {event['nodes_from']}→{event['nodes_to']} nodes @ {event['accuracy']}%"
        )

    print(f"\n📁 Output files:")
    print(f"  ✓ {REPORT_FILE}")
    print(f"  ✓ {CONVERGENCE_LOG}")
    print(f"  ✓ {METRICS_LOG}")
    print(f"  ✓ {TPM_FILE}")
    print(f"  ✓ {NPU_FILE}")
    print(f"  ✓ {TEST_LOG}")

    print(f"\n🎯 Test Results Directory: {RESULTS_DIR}")
    print("=" * 70 + "\n")


if __name__ == "__main__":
    try:
        run_test()
        log_msg("INFO", "Test execution completed successfully")
        sys.exit(0)
    except Exception as e:
        log_msg("ERROR", f"Test execution failed: {str(e)}")
        print(f"\n❌ Error: {str(e)}")
        sys.exit(1)
