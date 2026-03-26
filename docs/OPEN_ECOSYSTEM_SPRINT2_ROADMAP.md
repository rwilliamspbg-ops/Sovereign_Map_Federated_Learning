# Open Ecosystem Sprint 2 Roadmap

## Sprint Goal

Deliver trust and governance transparency for marketplace operations:

1. Explainable scoring in matching outcomes.
2. Local dispute workflow for contract issues.
3. Governance action logging surfaces.
4. Dashboard visibility for dispute and governance activity.

## Duration

- 2 weeks

## Scope

### P0 (Must Deliver)

1. Match score breakdown included per selected offer.
2. Governance activity endpoints (create/list).
3. Dispute endpoints (create/list/update).
4. Metrics summary governance snapshot.
5. UI rendering of score breakdown and governance activity.
6. Automated test coverage for new endpoints.

### P1 (If Capacity Allows)

1. Dispute SLA timers and escalation status.
2. Governance action filtering by actor and source.
3. Policy proposal voting workflow stub.

## User Stories

1. As a buyer, I can see why an offer was selected using score components.
2. As an operator, I can submit disputes for problematic contracts.
3. As a moderator, I can update dispute status and leave resolution notes.
4. As a governance observer, I can see recent governance actions in one view.

## Acceptance Criteria

1. `/marketplace/match` includes `score_breakdown` and `selection_diagnostics`.
2. `/marketplace/disputes` supports create/list.
3. `/marketplace/disputes/<id>` supports status updates.
4. `/governance/actions` supports create/list.
5. `/metrics_summary` includes governance snapshot.
6. Frontend displays score breakdown and recent governance actions.
7. Local tests and frontend build pass.

## Definition of Done

1. Backend API endpoints implemented and documented.
2. Frontend views expose explainability and governance visibility.
3. Tests validate score, dispute, and governance workflows.
4. No diagnostics errors in touched files.
