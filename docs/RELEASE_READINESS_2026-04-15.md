# Release Readiness Note - 2026-04-15

## Scope

This note captures the latest AV smoke and chaos guard release-readiness evidence gathered from local runtime validation.

## Evidence Bundle

- Artifact root: `test-results/release-evidence/20260415T160033Z`
- Key files:
  - `av_e2e_smoke.log`
  - `chaos_guard.log`
  - `chaos_guard_final.log`
  - `chaos_guard_runtime.log`
  - `chaos_final_exit_code.txt`
  - `chaos_runtime_exit_code.txt`
  - `backend_runtime.log`
  - `node-agent_runtime.log`
  - `prometheus_runtime.log`

## CI and Validation References

- Full validation PR gate workflow:
  - https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/full-validation-pr-gate.yml
- AV ingestion checks workflow:
  - https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/av-ingestion-checks.yml
- Full validation workflow:
  - https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/full-validation.yml

## What Passed

- AV ingest and pose path checks passed in smoke evidence:
  - ingest contracts
  - alignment path checks
  - SLAM pose contract checks
- Runtime stack startup succeeded in latest chaos run:
  - backend healthy
  - two node-agent containers started
  - Prometheus started

## What Failed

- Chaos guard final and runtime attempts failed strict gate conditions.
- Latest runtime failure mode:
  - `cadence gate not satisfied before next restart`
  - observed state at failure: `round=2.0`, `baseline=2.0`, `active_nodes=0.0`, `quorum=1.0`
- Exit codes:
  - `chaos_final_exit_code.txt = 1`
  - `chaos_runtime_exit_code.txt = 1`

## Known Risks

1. Chaos recovery behavior is not yet stable enough for release sign-off under strict cadence gate settings.
2. Tile-stage validation is environment-gated where OpenCV-tagged modules are not enabled.
3. Rebuild-heavy local runs can cause disk pressure and extend validation turnaround time.

## Release Decision (Current)

- Status: **NOT READY FOR FULL RELEASE SIGN-OFF**
- Reason: strict chaos cadence gate still failing, and tile-stage remains environment-gated in default local validation.

## Required Follow-Up Before Sign-Off

1. Run hardware-backed GPU validation and attach report artifacts.
2. Expand deterministic replay fixtures for AV timing/dropout edge cases.
3. Harden CI gates to classify and fail independently on:
   - preflight quorum failure
   - cadence gate failure
   - sustained active-node loss during restart cycle
