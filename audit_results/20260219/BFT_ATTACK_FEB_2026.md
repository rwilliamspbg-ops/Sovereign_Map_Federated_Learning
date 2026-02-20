# 🛡️ BFT Attack Simulation Report (Feb 2026)

## Executive Summary
* **Test Date:** February 18, 2026
* **Protocol Version:** v0.2.1-alpha
* **Attack Profile:** 30% Byzantine Nodes (Adversarial Gradients)
* **Pass/Fail:** ✅ PASS
* **Final Accuracy:** 85.42%

## 🧪 Simulation Parameters
This test utilized the `phase-4-execute-test.sh` script to stress-test the coordinator-less mesh under adversarial conditions.

| Parameter | Value |
| :--- | :--- |
| **Total Nodes** | 10 |
| **Malicious Nodes** | 3 (30%) |
| **Privacy Standard** | SGP-001 (ε=0.98) |
| **Consensus Mechanism** | dAuth Weighted BFT |

## 📈 Key Findings
1. **Convergence Resilience:** Despite 30% of the network reporting "noise" or inverted gradients, the global model converged within 12% of the "Clean" baseline.
2. **Privacy Integrity:** The SGP-001 privacy layer successfully throttled nodes attempting to leak membership data during the attack.
3. **Latency:** Consensus overhead remained under 2s per round, proving the scalability of the NPU-accelerated attestation.

## 📂 Source Data
Raw logs and JSON metrics are archived in the `audit_results/` directory under the following timestamps:
- `results-20260218-080750`
- `results-20260218-081515`
