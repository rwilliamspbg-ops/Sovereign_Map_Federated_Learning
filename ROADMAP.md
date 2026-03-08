# Sovereign Map Roadmap (2026)

## Scope
This roadmap tracks execution priorities after the v1.1.0 readiness update on branch `fork/Full-Testnet-Deployment-v1.1.0`.

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

## In Progress
- End-to-end validation across mesh, storage, and sensor pipeline.
- Operational hardening for production-like deployments.

## Next Milestones

### Milestone 1: Integration Confidence
- Add integration tests for:
  - node join/discovery under mixed network conditions
  - checkpoint write/read flows for filesystem + IPFS
  - camera/SLAM -> tile generation pipeline
  - mobile/drone ingest -> aggregation input path
- Exit criteria:
  - CI green on integration suite
  - deterministic test fixtures for sensor data

### Milestone 2: OpenCV and Build Reliability
- Add OpenCV-capable CI path for `gocv` packages.
- Document supported environments for camera/SLAM builds.
- Exit criteria:
  - reproducible build docs for Linux/macOS
  - one CI lane compiling camera and SLAM packages

### Milestone 3: Observability and SLOs
- Add dashboards/alerts for:
  - peer count, join success, relay usage
  - checkpoint storage latency/error rates
  - sensor ingestion queue depth and drop rate
- Exit criteria:
  - baseline SLO dashboards published
  - alert runbook linked in docs

### Milestone 4: Scale and Readiness Gate
- Run staged scale tests (10 -> 100 -> 1000 nodes).
- Validate throughput, convergence, and stability at each stage.
- Exit criteria:
  - scale report captured in repository
  - release gate checklist fully checked

## Risks
- OpenCV dependency variance across environments.
- IPFS daemon/network behavior in constrained environments.
- NAT traversal behavior in heterogeneous network topologies.

## Tracking
- Issue labels recommended: `roadmap`, `integration`, `performance`, `observability`, `testnet`.
- Keep this document aligned with `README.md` and release notes.
