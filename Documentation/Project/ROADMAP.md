<!-- markdownlint-disable MD005 MD007 MD013 MD022 MD032 -->

# Sovereign Map Roadmap (2026)

## Scope
This roadmap tracks execution priorities for the current `v1.2.0` platform baseline on `main`.

## Current Release Baseline
- Version `v1.2.0` published on 2026-03-16.
- Release notes: `docs/RELEASE_NOTES_v1.2.0.md`.
- Current readiness summary: `Documentation/Project/RELEASE_READINESS_2026-03-17.md`.

## Completed
- P2P mesh runtime connected to libp2p default transports (TCP + QUIC).
- NAT traversal service scaffolded with AutoNAT/AutoRelay/Relay hooks.
- Model checkpoint storage expanded to multi-backend mode (filesystem + IPFS).
- Sensor ingestion modules expanded:
  - Camera frame capture
  - SLAM feature extraction/bridge
  - Mobile client ingestion API
  - Drone telemetry ingestion
- Map tile encoding and cache scaffolding added.
- Documentation updates published:
  - `NETWORK_READINESS_ASSESSMENT.md`
  - `IMPLEMENTATION_SUMMARY.md`
  - `OPENCV_INSTALL.md`

## Completed in v1.1.0 Finalization
- End-to-end validation across mesh, storage, and sensor pipeline.
- Operational hardening for production-like deployments.
- OpenCV-tagged CI lane validated green on `main`.

## Completed in Post-v1.1.0 Hardening (2026-03)
- Node-agent API listener wiring with versioned capability/proof/hybrid/ledger routes.
- Proof API auth middleware added with token-file validation and role enforcement.
- Proof verification ledger added as a bounded in-memory ring buffer with API access.
- Hybrid verifier package expanded with FRI/Winterfell verification and backend registry tests.
- Secure channel ECDH migration and TLS 1.3 handshake/deadlock regression coverage added.
- End-to-end node-agent integration test updated to validate auth-gated proof and ledger behavior.
- Capability contract test (`TestCapabilitiesContractV1`) added and wired into CI build workflow.
- Versioned capability schema and release-readiness summaries published in project documentation.
- Real tokenomics telemetry exporter integrated with Prometheus/Grafana compatibility rules.
- Event-driven TPM attestation and message telemetry ingestion wired with exporter endpoints.
- FL SLO alert rules (`fl_slo_alerts.yml`) and observability query validation checks added.
- Security supply-chain workflow added (SBOM + Trivy SARIF upload).
- Local participant onboarding API added (invite/register/list/revoke) with cert issuance.
- Windows laptop client launcher + automated Windows EXE build workflow added.

## Completed in Sprint 10 Documentation and HUD Cleanup (2026-03-23)
- Frontend mode navigation consolidated to HUD-first operation (Browser FL Studio removed from primary app mode selector).
- Web runtime telemetry surfaced directly inside HUD (browser RTT/downlink, heap, page load, viewport, device profile).
- Privacy-Utility dashboard now consumes backend convergence metrics as canonical data source.
- Dashboard query validator allowlist updated for consensus/aggregation metrics used in operations dashboard panels.

## Completed in v1.2.0 Cycle (2026-03)
- Wallet-enabled blockchain core integrated with FL/coordinator pathways.
- Smart contract runtime and governance execution hooks wired through consensus flow.
- Validator trust/reputation controls expanded with policy-aware metrics.
- Testnet wallet readiness gate standardized for release operations.
- Root npm build/test workflow paths corrected to build both SDK packages (`privacy`, `core`) in CI.
- SDK publish and reproducibility workflows aligned to lockfile-driven npm installs.

## Completed in Post-v1.2.0 Reliability Sweep (2026-04)
- Observability CI extended with scheduled drift checks and release-event validation (`published`, `prereleased`).
- Added macOS client smoke workflow for cross-platform client confidence.
- Windows client EXE lane and macOS client smoke lane both validated green on `main` during cross-platform closure.
- Contributor baseline checks re-validated and documented (`npm run test:ci`, Python formatting/lint gates, targeted monitoring tests).
- Runtime performance profile controls added (`ultra_latency`, `balanced`, `throughput`) with profile-aware cadence/retry/backpressure behavior.
- Provider-aware execution policy publication added to runtime summaries (`hardware_class`, `provider`, optimizer flags, fallback order).
- Memory-pressure adaptive control loop telemetry added with dedicated Prometheus alerts for critical pressure and sustained backpressure.
- Deterministic workflow progress contracts added across training and chaos guard phases.

## Next Milestones

### Milestone 1: API and Security Confidence
- Status: completed
- Added negative-path API tests for malformed payloads, missing tokens, and role mismatches.
- Added stress-style ledger tests to validate retention behavior at high verification event volume.
- Exit criteria:
  - CI includes API failure-path checks for proof/hybrid/ledger routes: met
  - ledger retention behavior validated under sustained event generation: met

### Milestone 2: Operator Readiness and Docs
- Status: completed
- Published token rotation and role policy runbook for compose and testnet setups.
- Added troubleshooting matrix for auth and role failures (`401`/`403`) in runtime docs.
- Exit criteria:
  - operator runbook published and linked from README: met
  - documented rollback and recovery steps for auth misconfiguration: met

### Milestone 3: Observability and SLOs
- Status: completed
- Added dashboards/alerts for:
  - proof acceptance/rejection rates by backend and mode
  - ledger event volume and verification latency bands
  - existing peer count, join success, relay usage, and checkpoint storage errors
- Added tokenomics and TPM event compatibility layers to keep dashboards stable during metric migrations.
- Added observability CI workflow to fail fast on invalid dashboard queries.
- Exit criteria:
  - baseline SLO dashboards published: met
  - alert runbook linked in docs: met

### Milestone 6: Participant Onboarding and Client Distribution
- Status: completed
- Added local testnet join APIs for invite minting, registration, and certificate-driven participant onboarding.
- Added participant bootstrap tooling and compose profile for local onboarding tests.
- Added Windows EXE launcher docs and CI workflow for reproducible laptop client packaging.
- Exit criteria:
  - participant onboarding documented and runnable locally: met
  - Windows executable build automated in CI: met

### Milestone 7: CI and Documentation Reliability
- Status: completed
- Standardized root project status reporting and badge surface in `README.md`.
- Updated roadmap and release-readiness documentation to reflect `v1.2.0` baseline and workflow behavior.
- Corrected npm workflow assumptions in CI so package build/test jobs use repository lockfiles and deterministic install paths.
- Exit criteria:
  - root status docs and roadmap aligned with current release: met
  - sdk build/test CI paths verified against local baseline commands: met
  - documentation index points to current readiness artifacts: met

### Milestone 8: HUD Consolidation and Metrics Surface
- Status: completed (2026-03-23)
- Removed Browser FL Studio from primary frontend navigation and made HUD the default entry view.
- Added browser runtime telemetry cards/panels to HUD for operator visibility.
- Realigned privacy analysis mode to backend convergence metrics for consistency.
- Exit criteria:
  - HUD is default and Browser FL Studio is no longer a primary mode: met
  - browser telemetry is visible in HUD KPI/panel surfaces: met
  - observability dashboard query validation includes all referenced metrics: met

### Milestone 9: AV-Ready v1.0 One-Pass Sprint
- Status: planned
- Execute a single-pass sprint to harden AV mapping runtime and validation gates.
- Canonical sprint document:
  - `docs/AV_READY_V1_ONE_PASS_SPRINT.md`
- Primary outcomes:
  - sensor contracts and ingestion validation
  - time alignment and calibration enforcement
  - fused pose output replacing fallback nominal path
  - end-to-end map update and deterministic tile generation
  - AV-specific observability and CI merge gate coverage
- Exit criteria:
  - one-pass definition of done in sprint document fully met
  - CI AV lane required and green on main
  - release-readiness summary updated with AV v1.0 evidence bundle

### Milestone 4: Scale and Readiness Gate
- Status: completed (2026-03-17)
- 10-node scale test executed on 4-core/15 GiB host; 10/10 agents confirmed running.
- FL metrics validated: `sovereignmap_fl_round=800`, accuracy=99.5%, loss=0.1.
- Tokenomics pipeline active: supply=7920, mint_rate=3.47/min.
- TPM pipeline connected; series present in Prometheus.
- Auto-accelerator detection (NPU→GPU→CPU) implemented in `deploy_demo.sh`.
- Host constraint documented: 25+ nodes require ≥32 GiB RAM (OOM risk on dev container).
- Linear extrapolation to 25/50/100/1000 nodes captured in `results/SCALE_REPORT_2026-03-17.md`.
- Exit criteria:
  - scale report captured in repository: **met** (`results/SCALE_REPORT_2026-03-17.md`)
  - release gate checklist fully checked: **met** (see report §6)

### Milestone 5: Capability Contract Stabilization
 - Status: completed
- Freeze `/api/v1/capabilities` as the runtime contract for API, auth, and observability metadata.
- Add contract tests to detect accidental capability schema regressions.
- Exit criteria:
  - capability schema documented and versioned: met
  - capability contract tests enabled in CI: met

## Risks
- OpenCV dependency variance across environments.
- IPFS daemon/network behavior in constrained environments.
- NAT traversal behavior in heterogeneous network topologies.

## Tracking
- Issue labels recommended: `roadmap`, `integration`, `performance`, `observability`, `testnet`.
- Additional labels recommended for current stream: `api`, `auth`, `ledger`, `hybrid-proof`.
- Keep this document aligned with `README.md`, current readiness summaries, and release notes.
