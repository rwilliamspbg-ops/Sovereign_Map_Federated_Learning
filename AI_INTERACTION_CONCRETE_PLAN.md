# AI Interaction and Ease-of-Use Plan

## Objective

Make the system easier to use, easier to trust, and easier to drive with AI-assisted interactions by reducing manual setup, exposing structured recommendations, and keeping advanced controls available but not dominant.

## Progress Update

Implemented so far:

1. Shared AI interaction summary endpoint for HUD and C2.
2. Natural-language command bar and structured recommendation cards in the main HUD.
3. Quick-action routing and summary previews in C2.

Remaining follow-up work:

1. Add explicit approve/edit/reject/undo loops for AI-suggested actions.
2. Expand model-routing heuristics beyond the current summary/planner split.
3. Persist user interaction history so operators can review prior AI decisions.

## Product Principles

1. Minimize the number of user decisions required for the common path.
2. Show a clear reason, confidence, and consequence for every AI recommendation.
3. Keep human approval in the loop for any action that changes mission state.
4. Hide prompt engineering behind product actions and model routing.
5. Default to the simplest path, but let experts expand into full control.

## Current Surface To Build On

1. Primary HUD at `frontend/src/HUD.jsx`.
2. C2 view at `frontend/src/C2SwarmHUD.jsx`.
3. Control-plane backend in `sovereignmap_production_backend_v2.py`.
4. Autonomy core in `internal/autonomy/*`.
5. Live ops events from SSE and metrics endpoints already in place.

## Recommendation Map

### 1) Make the main interaction path one-step

What to build:

1. A single natural-language command bar for the primary HUD.
2. Pinned quick actions for the most common workflows.
3. Context auto-fill from current mission, map, policy, and twin state.

Implementation details:

1. Parse user intent into a structured action request.
2. Show the interpreted action before execution.
3. Require one confirmation click for state-changing actions.

Acceptance criteria:

1. Users can request a common task in one sentence.
2. The UI shows the interpreted action, confidence, and required inputs.
3. No manual navigation is needed for the default workflow.

### 2) Expose structured AI suggestions, not free-form text

What to build:

1. Recommendation cards with action, reason, confidence, and expected outcome.
2. Safe alternatives when the top action is blocked by policy.
3. Clear rejection reasons when the AI declines to act.

Implementation details:

1. Use a typed payload for suggestions from backend to frontend.
2. Normalize outputs into a common shape: `action`, `reason`, `confidence`, `risk`, `expected_gain`, `blocked_reason`.
3. Render the same structure in HUD and C2.

Acceptance criteria:

1. Every recommendation is explainable in one glance.
2. The same recommendation format is reused across views.
3. Blocked actions always show why they were blocked.

### 3) Add approve/edit/reject loops for every AI action

What to build:

1. Buttons for approve, edit, reject, and undo.
2. A review drawer that lets users modify the AI-suggested action before execution.
3. A visible audit trail of user decisions.

Implementation details:

1. Keep state changes behind a confirmation boundary.
2. Save the user override reason with the action event.
3. Include rollback hooks for reversible actions.

Acceptance criteria:

1. The user can accept, change, or reject any AI suggestion.
2. Every decision is logged with who approved it and why.
3. Undo is available for reversible operations.

### 4) Use progressive disclosure for expert controls

What to build:

1. A simple default mode.
2. An expert mode that reveals model selection, thresholds, and raw telemetry.
3. A “Do it for me” mode for routine tasks and a manual override mode for advanced users.

Implementation details:

1. Keep advanced controls collapsed until explicitly requested.
2. Preserve a single route to the same underlying action, regardless of mode.
3. Make mode switches persistent per user preference.

Acceptance criteria:

1. New users see a smaller, less intimidating surface.
2. Power users can still access raw telemetry and model controls.
3. Mode changes never change the underlying safety checks.

### 5) Route requests to the right model automatically

What to build:

1. A lightweight model-router layer that selects a model by task type.
2. Default routing for classification, summarization, planning, and map reasoning.
3. Manual override for power users and debugging.

Implementation details:

1. Define task classes such as `summary`, `planner`, `map_reasoning`, `safety_review`, and `search`.
2. Route based on latency budget, cost, and output shape.
3. Keep routing decisions visible in debug mode.

Acceptance criteria:

1. The user does not need to choose a model for common tasks.
2. The system picks an appropriate model automatically.
3. Routing decisions are explainable when expanded.

### 6) Return structured outputs and metadata every time

What to build:

1. Standard response envelopes for AI answers.
2. Confidence, assumptions, sources, and freshness fields.
3. Support for multimodal output where appropriate.

Implementation details:

1. Enforce a shared schema across backend responses.
2. Include `confidence`, `assumptions`, `source_span`, `freshness_secs`, and `next_action`.
3. Make the UI render charts, tables, and map overlays from the same response family.

Acceptance criteria:

1. AI responses are consistent and machine-readable.
2. Users can see trust signals without opening developer tools.
3. The UI can render the same answer as text, card, or overlay.

### 7) Improve mission awareness and search

What to build:

1. Live context panels that auto-load map, telemetry, policy, and recent events.
2. Fast search over past decisions and operator actions.
3. A timeline that explains why the AI changed its recommendation.

Implementation details:

1. Load context automatically on page open.
2. Index decisions, replan triggers, overrides, and twin changes.
3. Let the user ask questions like “why did the system reroute here?”

Acceptance criteria:

1. Users can inspect prior decisions in seconds.
2. The system can answer “why” questions using recent state and audit history.
3. The timeline links cause, decision, and outcome.

## Implementation Phases

### Phase 1: Command Bar and Recommendation Cards

Goal: reduce the first action path to one sentence and one confirmation.

Deliverables:

1. Command bar in `frontend/src/HUD.jsx`.
2. Recommendation cards for action/reason/confidence/risk.
3. Backend payload shape for structured AI suggestions.

### Phase 2: Explainability and Safety Review

Goal: make every AI action inspectable before execution.

Deliverables:

1. Review drawer with approve/edit/reject/undo.
2. Safety panel with policy state and blocked-reason display.
3. Audit trail events for all AI-assisted decisions.

### Phase 3: Model Routing and Structured Metadata

Goal: remove manual model choice from the common path.

Deliverables:

1. Task-type router service.
2. Standard AI response envelope.
3. Debug mode that shows model choice and fallback logic.

### Phase 4: Mission Context and Search

Goal: make the system answerable and self-explanatory.

Deliverables:

1. Context auto-loading panels.
2. Decision search and replay.
3. Timeline linking inputs to outputs and outcomes.

### Phase 5: Rollout and Hardening

Goal: deploy safely with measurable usability gains.

Deliverables:

1. Feature flags for command bar, recommendation cards, and expert mode.
2. Canary rollout plan.
3. UX telemetry dashboard.

## Suggested Ticket Breakdown

1. UX-001: Build the natural-language command bar.
2. UX-002: Add structured recommendation cards.
3. UX-003: Add approve/edit/reject/undo workflow.
4. UX-004: Add expert mode and progressive disclosure.
5. ML-001: Add automatic model routing.
6. ML-002: Add structured response envelopes.
7. OPS-001: Add context auto-loading and search.
8. OPS-002: Add decision timeline and replay.

## Metrics To Track

1. time_to_first_action_secs
2. command_success_rate
3. recommendation_acceptance_rate
4. edit_after_suggestion_rate
5. override_rate
6. model_routing_accuracy
7. answer_confidence_display_rate
8. decision_replay_usage

## Rollout Criteria

1. New users can complete the common path without a tutorial.
2. Recommendation cards reduce back-and-forth interaction.
3. Operator overrides stay within acceptable bounds.
4. No safety regressions are introduced by model routing or structured outputs.

## Validation Plan

1. Unit tests for command parsing, recommendation ranking, and response schema validation.
2. Integration tests for backend payloads and HUD rendering.
3. E2E tests for command bar -> suggestion -> confirm -> execute -> audit trail.
4. Usability review with the simplest path measured in clicks and time-to-action.

## First Branch Execution Order

1. Create the command bar and recommendation card components.
2. Add structured AI response envelopes in the backend.
3. Wire approve/edit/reject/undo actions to the audit trail.
4. Add model routing and expert-mode toggles.
5. Add search and timeline replay after the core flow is stable.

## Definition of Done

1. The common path is shorter and clearer for new users.
2. AI suggestions are structured, explainable, and actionable.
3. Advanced controls remain available without overwhelming the default view.
4. The system can justify model choice, action choice, and safety decisions.