# AI Interaction Phase 4 Concrete Execution Plan

## Scope and Objective

This plan defines the next concrete execution slice after PR #99 and focuses on three outcomes:

1. Production-safe auth and policy for interaction decision logging.
2. Searchable, replayable digital-twin decision context for operators.
3. Stability gates for merge confidence (tests, lint, and strict soak checks).

## Why This Phase

The current interaction workflow is operational, but merge confidence and operator trust require:

1. Explicit auth boundaries for `/ai/interaction/decision`.
2. Backend contract tests for interaction APIs.
3. Search/replay quality improvements in the HUD twin perspective.

## Deliverables

### D1. Interaction Decision Auth Hardening

Files expected to change:

1. `sovereignmap_production_backend_v2.py`
2. `README.md`
3. `docs/api/openapi.yaml`

Concrete implementation:

1. Add an explicit policy mode for interaction decisions:
   - `AI_INTERACTION_DECISION_AUTH_MODE=public_local|admin_required`
2. Default to `admin_required` for production-safe behavior.
3. Allow `public_local` only for local/test flows with explicit env opt-in.
4. Return structured auth failure payload with remediation hint.

Acceptance criteria:

1. Decision POST is blocked unless policy mode allows request context.
2. Auth mode is documented with examples.
3. Existing local dev flow still works when enabled explicitly.

### D2. Backend Contract Tests for Interaction APIs

Files expected to change:

1. `tests/scripts/python/` (new focused test module)
2. Optional test helpers under existing test utilities

Concrete implementation:

1. Add tests for:
   - `GET /ai/interaction/summary`
   - `GET /ai/interaction/history`
   - `POST /ai/interaction/decision`
2. Cover both positive and negative paths:
   - valid decision values (`approve`, `edit`, `reject`, `undo`)
   - invalid decision rejection
   - auth-restricted behavior under strict mode
3. Assert response envelope fields are present and typed.

Acceptance criteria:

1. Tests fail if schema-critical keys are removed.
2. Tests fail if auth mode behavior regresses.
3. Tests run in CI without requiring external services.

### D3. Digital Twin Search and Replay UX

Files expected to change:

1. `frontend/src/HUD.jsx`
2. Optional helper/style files if needed

Concrete implementation:

1. Add a decision search input (filter by action, decision, route, reason).
2. Add replay list controls:
   - sort by time (newest/oldest)
   - quick filters (`approve`, `reject`, `undo`)
3. Add a replay detail panel for selected decision:
   - model route
   - operator reason
   - associated prompt
   - timestamp and review ID

Acceptance criteria:

1. Operators can find a prior decision in under 10 seconds.
2. Search and filters work with both backend and locally-added entries.
3. Replay detail remains readable on mobile and desktop.

### D4. Stability and Merge Gates

Files expected to change:

1. `AI_INTERACTION_PHASE4_CONCRETE_EXECUTION_PLAN.md` (status updates)
2. Optional docs updates if commands change

Concrete implementation:

1. Run and record targeted checks:
   - `python3 -m py_compile sovereignmap_production_backend_v2.py`
   - `npm --prefix frontend run lint`
   - `npm --prefix frontend run test -- --run src/HUD.test.jsx src/C2SwarmHUD.test.jsx`
   - `npm --prefix frontend run build`
2. Run strict chaos guard and attach pass/fail details:
   - `SOAK_CHAOS_ENABLED=1 SOAK_CHAOS_STRICT=1 CHAOS_MIN_CLIENT_QUORUM=1 python3 tests/scripts/python/test_soak_chaos_guard.py`

Acceptance criteria:

1. Required checks are documented in PR body.
2. Any remaining warnings are explicitly called out with owner/follow-up.
3. Strict soak result is reported, not omitted.

## Execution Timeline

### Week 1

1. D1 auth hardening + docs.
2. D2 backend contract tests.

### Week 2

1. D3 search/replay UI.
2. D4 full validation and PR evidence updates.

## Task Breakdown

1. `AI-AUTH-001`: Add interaction decision auth mode and enforcement.
2. `AI-AUTH-002`: Document auth mode and testing examples.
3. `AI-TEST-001`: Add summary/history/decision backend contract tests.
4. `AI-UX-001`: Implement decision search and quick filters in HUD.
5. `AI-UX-002`: Implement replay detail panel.
6. `AI-OPS-001`: Run validation matrix and publish evidence.

## Risk Register

1. Risk: Auth hardening may break local test UX.
   - Mitigation: explicit `public_local` mode and startup logs.
2. Risk: Replay UX increases HUD complexity.
   - Mitigation: default collapsed panel and simple filter presets.
3. Risk: Strict soak remains unstable.
   - Mitigation: capture first failing assertion and gate merge on documented disposition.

## PR Completion Checklist

1. [ ] Auth mode implemented and documented.
2. [ ] Interaction API contract tests added and passing.
3. [ ] HUD replay search/filter/detail shipped.
4. [ ] Lint/test/build checks completed.
5. [ ] Strict soak run completed with outcome included.
6. [ ] PR description includes validation commands and residual risk notes.
