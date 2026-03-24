# Ops Optimization Roadmap (Started 2026-03-24)

## Objective

Improve operations maturity across visual clarity, runtime performance, and operator usability while preserving production safety.

## Success Metrics

- Time-to-detect critical regressions: < 3 minutes.
- Time-to-triage major incidents: < 10 minutes.
- Dashboard p95 panel render latency: < 1.5 seconds for default window.
- Alert-to-runbook match coverage: 100 percent for critical and warning alerts.
- Incident evidence bundle generation time: < 60 seconds.

## Phase 1 (Week 1): Operator Clarity and Fast Triage

Status: Completed (2026-03-24).

Planned outcomes:

- Add "What Changed" panel for short-window trend deltas.
- Add one-command incident bundle export for postmortem artifacts.
- Finalize docs lint and runbook structure consistency for on-call reliability.

Acceptance criteria:

- Operators can identify top changing signals without manually comparing windows.
- Incident bundle script produces timestamped artifacts and summary metadata.
- No CI regressions in dashboard query validation or lint guard workflows.

## Phase 2 (Week 2): Performance and Query Efficiency

Status: In progress.

Planned outcomes:

- Add recording rules for expensive repeated dashboard expressions.
- Introduce query budget checks for heavy panels and add fallback behavior.
- Optimize label-cardinality hotspots for TPM and FL dimensions.

Acceptance criteria:

- Dashboard p95 render latency reduced by at least 25 percent vs baseline.
- No panel timeout in default Operations overview at 15-minute window.

## Phase 3 (Week 3): Guided Operations and HUD Actions

Status: In progress.

Planned outcomes:

- Add runbook-match cards in HUD for active alerts.
- Add guarded action controls for common mitigations (temporary, auditable).
- Add timeline annotations for deploy/config changes in Grafana.

Acceptance criteria:

- Operators can complete first-response flow without leaving HUD/Grafana.
- Every guarded action is logged with actor, reason, and expiry.

## Phase 4 (Week 4): Reliability and Governance Hardening

Status: Planned.

Planned outcomes:

- Add pre-release synthetic query smoke tests for all starred dashboards.
- Expand incident evidence artifacts and retention policy alignment.
- Add periodic dashboard/query drift checks to CI.

Acceptance criteria:

- New dashboard/query drift is detected before production deploy.
- Evidence bundle supports standard post-incident review template.

## Work Started Today

- Added a "What Changed (Current vs Prior Window)" panel to Operations dashboard.
- Added incident-bundle export helper script for operator workflows.
- Added runbook-match cards in HUD for active alerts.
- Added timeline annotations for config/control changes in Grafana.
- Added recording-rule groundwork for expensive TPM/operator dashboard queries.
- Added CI guard workflow for incident bundle tooling.
- Added roadmap and execution checkpoints in this document.
