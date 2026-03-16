<!-- markdownlint-disable MD005 MD007 MD022 MD032 -->

# Sovereign Map Roadmap (2026)

## Scope
This roadmap tracks execution priorities for the current `v1.2.0` platform baseline on `main`.

## Current Release Baseline
- Version `v1.2.0` published on 2026-03-16.
- Release notes: `docs/RELEASE_NOTES_v1.2.0.md`.
- Current readiness summary: `Documentation/Project/RELEASE_READINESS_2026-03-16.md`.

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

## Completed in v1.2.0 Cycle (2026-03)
- Wallet-enabled blockchain core integrated with FL/coordinator pathways.
- Smart contract runtime and governance execution hooks wired through consensus flow.
- Validator trust/reputation controls expanded with policy-aware metrics.
- Testnet wallet readiness gate standardized for release operations.
- Root npm build/test workflow paths corrected to build both SDK packages (`privacy`, `core`) in CI.
- SDK publish and reproducibility workflows aligned to lockfile-driven npm installs.

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

### Milestone 4: Scale and Readiness Gate
- Run staged scale tests (10 -> 100 -> 1000 nodes) with updated API/auth settings.
- Validate throughput, convergence, and stability at each stage.
- Exit criteria:
  - scale report captured in repository
  - release gate checklist fully checked

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
