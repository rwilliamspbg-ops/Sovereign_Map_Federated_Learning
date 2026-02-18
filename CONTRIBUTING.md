# Contributing to Sovereign Map

First off, thank you for considering contributing to the **Sovereign Map DePIN Protocol**. It’s people like you that make the decentralized spatial web a reality.

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md) and ensure that all contributions respect the **SGP-001** privacy standards.

---

## 💖 Support the Evolution

Developing a coordinator-less, hardware-enforced mapping protocol requires significant R&D and infrastructure for large-scale simulations. If you aren't a developer but want to support the project, consider a donation to our community treasury.

[![GitHub Sponsor](https://img.shields.io/badge/Sponsor-GitHub-EA4AAA?style=flat&logo=github)](https://github.com/sponsors/rwilliamspbg-ops)

*All donations go directly toward maintaining the Genesis Node testnets and SGP-001 third-party audits.*

---

## 🛠️ Getting Started

### 1. Environment Setup
We recommend using the provided [Dev Container](.devcontainer/) or a local Linux environment.
* **Go**: 1.21+ (Protocol Core)
* **Python**: 3.10+ (Research & Analysis)
* **Node.js**: 18+ (CLI & Dashboard)

### 2. Running Simulations
Before submitting a PR, ensure your changes do not break the consensus engine by running the 200-node simulation:
```bash
bash scripts /run-test.sh --nodes 200
