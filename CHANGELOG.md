# Changelog

All notable changes to the Sovereign Map project will be documented in this file. This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-beta] - 2026-02-11

### Added
- **SGP-001 Audit Framework**: Integrated automated GitHub Action to verify privacy parameters ($ε=1.0$, $δ=1e-5$) on every pull request.
- **MOHAWK Orchestrator**: Initial logic for dynamic community selection based on hardware telemetry.
- **dAuth Protocol**: Coordinator-less node verification for secure Genesis Node onboarding.
- **Independent Island Mode**: Operational safety feature for offline edge compute.

### Changed
- **Hardware Baseline**: Standardized minimum NPU throughput at **85 TOPS** for Genesis participation.
- **Docker Integration**: Hardened container environment with read-only volume mounts for audit configurations.

### Fixed
- **CI/CD Pipeline**: Resolved `docker-compose` command deprecation by migrating to `docker compose` V2 syntax.
- **Python Environment**: Standardized GitHub Action runners to Python 3.10 to ensure deterministic audit results.
