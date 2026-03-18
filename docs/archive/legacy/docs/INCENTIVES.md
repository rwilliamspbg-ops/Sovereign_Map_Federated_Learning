# 💰 Node Operator Incentive Model (v1.0)

This document outlines the economic framework for the **Sovereign Map Federated Learning Network**. Our goal is to reward high-performance [Genesis Nodes](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/HARDWARE.md) that provide the computational backbone for privacy-preserving AI.

---

## 💎 Reward Tiers & Multipliers
Incentives are distributed based on a combination of **Compute Availability**, **Privacy Integrity**, and **Network Reliability**.

| Tier | Requirement | Reward Multiplier | Base Allocation |
| :--- | :--- | :--- | :--- |
| **Genesis Elite** | >85 TOPS + 99.9% Uptime | 1.5x | 5,000 $SOV / Month |
| **Standard Node** | 85 TOPS + 95% Uptime | 1.0x | 3,300 $SOV / Month |
| **Edge Contributor** | Variable TOPS (Community) | 0.5x | Dynamic |

---

## 📈 Revenue Streams for Operators

### 1. Training Participation (Proof-of-Training)
Operators earn rewards for every federated learning round successfully completed.
* **MOHAWK Bonus**: Nodes selected by the [MOHAWK Orchestrator](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) for specialized heterogeneous workloads receive a **20% premium** on base rewards.
* **Privacy Premium**: Successfully passing the [SGP-001 Audit](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) without divergence triggers a "Zero-Knowledge Bonus."

### 2. Infrastructure Availability (Staking & Uptime)
Nodes that maintain a consistent connection in [Independent Island Mode](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) during network partitions are rewarded for preserving local state integrity.

### 3. Data Sovereignty Fees
Enterprises utilizing the network for private model training pay a "Sovereignty Fee" in stablecoins, which is distributed directly to the active node cluster participating in that specific training set.

---

## 🛡️ Governance & Slashing
To protect the [dAuth Protocol](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) and network integrity:
* **Audit Failure**: Nodes that fail a [SGP-001 Privacy Audit](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) (e.g., attempting to egress raw data) will be immediately slashed and jailed from the Genesis pool.
* **Latency Penalties**: Nodes consistently falling below the required memory bandwidth baseline ([228 GB/s](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/HARDWARE.md)) will have their multipliers adjusted downward.

---

## 🚀 How to Participate
1. **Verify Hardware**: Ensure your system meets the [85 TOPS NPU requirement](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/HARDWARE.md).
2. **Onboard**: Complete the [dAuth Handshake](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md) via our CLI.
3. **Earn**: Begin contributing to the [v0.2.0-alpha Stress Test](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/documentation/TECH_SPEC_V0.2.0.md).

---
**SEE ALSO**: [Genesis Hardware Specs](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/HARDWARE.md) | [Technical Roadmap](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/README.md)
