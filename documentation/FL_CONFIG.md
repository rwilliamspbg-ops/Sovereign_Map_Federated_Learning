# 🛡️ Federated Learning Privacy Configuration (SGP-001)

This document defines the mathematical constraints for Differential Privacy (DP) and Federated Learning rounds within the [Sovereign Map](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning) ecosystem.

## 🔒 Privacy Parameters (ε, δ)
To satisfy SGP-001 audit standards, the following thresholds are enforced at the node level:

| Parameter | Value | Description |
| :--- | :--- | :--- |
| **Epsilon (ε)** | `1.0` | Privacy budget per round; lower values provide stronger privacy. |
| **Delta (δ)** | `1e-5` | Probability of information leakage; must be smaller than the inverse of the dataset size. |
| **Clipping Norm** | `1.0` | Maximum $L2$ norm for local gradient updates to prevent outlier leakage. |
| **Noise Multiplier** | `0.1` | Standard deviation of Gaussian noise relative to the clipping norm. |

## ⚙️ Orchestration Settings
As defined in the MOHAWK Framework, these settings optimize for node heterogeneity:

* **Min. Participants**: `10 Nodes` — Minimum required nodes to begin a secure aggregation round.
* **Target Throughput**: `85 TOPS` — Minimum NPU performance required for Genesis Node participation.
* **Aggregation Method**: `FedAvg` (with Secure Multi-Party Computation).
* **Update Frequency**: `Synchronous` — Rounds occur after all selected nodes report local gradients.

---

## 🛰️ Verification (dAuth)
Every configuration update must be signed by the **Orchestrator** and verified by the node via the dAuth coordinator-less handshake.

> **Audit Note**: Any deviation from the $ε = 1.0$ threshold will trigger an automatic SIA Audit alert and temporarily move the node into **Independent Island** mode.
