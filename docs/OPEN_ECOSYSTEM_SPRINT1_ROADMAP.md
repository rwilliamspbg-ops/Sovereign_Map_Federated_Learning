# Open Ecosystem Sprint 1 Roadmap

## Sprint Goal

Deliver a user-friendly local-first marketplace loop that a new user can complete in one session:

1. Create offer
2. Create intent
3. Match contract
4. Trigger training round
5. Release escrow
6. Inspect timeline and metrics

## Duration

- 2 weeks

## Scope

### P0 (Must Deliver)

1. Deterministic marketplace API error codes and messages.
2. Match failure diagnostics to explain why offers did not match.
3. Contract lifecycle timeline (created -> bound_to_round -> escrow_released).
4. Dashboard visibility for pending contracts and marketplace summary.
5. Positive and negative-path local tests.
6. First-10-minutes onboarding guide.

### P1 (If Capacity Allows)

1. UI score breakdown view (quality, cost, trust).
2. Intent status workflow guardrails in controls.
3. Lightweight policy preview before matching.

## User Stories

1. As a data provider, I can create an offer with clear field guidance and immediate validation.
2. As a model owner, I can understand exactly why matching failed.
3. As an operator, I can see contract state transitions in chronological order.
4. As a reviewer, I can verify ecosystem activity through metrics and operation events.
5. As a new integrator, I can finish the full local flow in 10 minutes.

## Acceptance Criteria

1. Marketplace endpoints return stable error codes and messages on all validation failures.
2. Match failure includes machine-readable rejection reasons and counts.
3. Every matched contract includes timeline events.
4. `training/status` includes the pending contract summary.
5. `metrics_summary` includes marketplace snapshot.
6. Local smoke and negative tests pass.
7. Frontend build passes.

## Risks and Mitigations

1. Risk: ambiguous match outcomes.

- Mitigation: include rejection reason counters and budget rejection count.

1. Risk: accidental status misuse.

- Mitigation: enforce intent status transitions server-side.

1. Risk: duplicate escrow release.

- Mitigation: explicit `contract_already_released` error.

## Definition of Done

1. Backend, frontend, tests, and docs updated.
2. No diagnostics errors in touched files.
3. Local backend marketplace smoke test passes.
4. Local backend negative-path test passes.
5. Frontend production build passes.

## Demo Checklist

1. Create an offer from the UI.
2. Create an intent from the UI.
3. Run a match and inspect status.
4. Trigger one FL round and inspect contract binding.
5. Release escrow and confirm updated timeline.
6. Inspect marketplace section in `/metrics_summary`.
