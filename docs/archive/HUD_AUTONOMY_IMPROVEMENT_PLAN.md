# HUD + Autonomy Improvement Plan

## Objective

Strengthen autonomous mapping and digital twin operations by upgrading HUD decision support, map intelligence, and closed-loop course-correction visibility using the existing frontend, C2 HUD, and backend telemetry stack.

## Current Strengths

1. Multi-view frontend already supports primary HUD and C2 HUD modes.
2. SSE ops-event stream already exists for live updates.
3. C2 map already renders nodes, links, and risk zones.
4. Backend exposes health, metrics, trust, training, and swarm endpoints.
5. New autonomy core package provides contracts, map/twin stores, planner logic, and readiness checks.

## Priority Improvements

## P0: Mission-Critical HUD Upgrades (Immediate)

1. Add an Autonomy Control Strip in HUD:
- Live twin lag and freshness badge.
- Coverage confidence badge.
- Course-correction status badge (stable, warning, rerouting).
- Planner safety state (policy pass/fail).

2. Add C2 map intelligence overlays:
- Coverage heatmap layer.
- Confidence decay layer.
- Predicted trajectory polylines for top N nodes.
- Replan trigger markers when risk thresholds are crossed.

3. Add recommendation panel:
- Top 3 safe corrective actions with rationale.
- Action confidence and expected mission gain.
- Guardrail reason when actions are blocked.

4. Add event correlation timeline:
- Correlate ops events with map shifts, policy changes, and training rounds.
- Show cause/effect chain for each correction.

## P1: Operational Accuracy and Explainability

1. Add explainability cards:
- Why an action was chosen.
- Which signals influenced decision score.
- Which policy constraints filtered alternatives.

2. Add sensor quality matrix widget:
- Per-source confidence, packet freshness, and anomaly score.
- Auto-failover indicators.

3. Add digital twin scenario runner:
- Preview expected state for 30s, 60s, 120s horizons.
- Compare candidate action outcomes side-by-side.

4. Add compute allocation HUD panel:
- Edge/node/backend execution placement.
- Accelerator use (GPU/NPU) and fallback mode status.

## P2: Hardening and Decision Reliability

1. Add anomaly detection indicators:
- Drift spikes.
- Confidence cliffs.
- Sensor disagreement alarms.

2. Add simulation confidence envelopes:
- Render uncertainty bounds around predicted paths.

3. Add mission SLO scoreboard:
- Twin lag SLO.
- Correction success SLO.
- Coverage SLO.
- API/control latency SLO.

4. Add incident and rollback assist:
- Recommended rollback actions when safety gates fail.
- Last-known-good control profile restore button.

## Proposed HUD Function Additions

## In frontend/src/HUD.jsx

1. buildAutonomyKPIModel(input):
- Compute and return normalized KPI model for twin lag, coverage confidence, risk, correction rate, and safety state.

2. derivePlannerSafetyState(trustStatus, opsHealth, policyDraft):
- Return policy pass/fail state with reason labels.

3. deriveRecommendedCorrections(opsEvents, trainingStatus, mapSnapshot):
- Return ranked corrections for display with confidence and mission gain estimates.

4. buildEventCorrelationTimeline(opsEvents, policyHistory, trends):
- Return timeline objects linking event causes to observed outcomes.

5. computeSLOStatusBadges(metrics):
- Return per-SLO badge color, label, and breach duration.

6. computeSensorQualityMatrix(opsHealth, mapTelemetry):
- Return source-level confidence, freshness, anomaly scores.

## In frontend/src/C2SwarmHUD.jsx

1. buildCoverageHeatmap(nodes, coverageCells):
- Convert map data into normalized overlay cells for rendering.

2. buildConfidenceDecayOverlay(mapState, now):
- Generate visual fade state from map confidence decay.

3. buildPredictedTrajectories(nodes, predictionPayload):
- Return polyline points and uncertainty widths per node.

4. detectReplanTriggers(nodes, riskZones, policyState):
- Return trigger markers and reason labels for map annotations.

5. summarizeCommandImpact(commandLog, mapState):
- Compute quick post-command impact summary.

6. deriveOperatorAssistCards(status, mapState, auditLog):
- Return top assist recommendations and warnings.

## Backend/API Enhancements Needed for HUD

1. Extend map payload endpoint with:
- map_version
- mean_confidence
- stale_cell_count
- replan_trigger_count
- predicted_paths[]
- confidence_envelopes[]

2. Add twin summary endpoint:
- /autonomy/twin/summary
- per-entity lag, confidence, risk, and prediction error.

3. Add planner insight endpoint:
- /autonomy/planner/insights
- candidate actions, rejected reasons, selected action score.

4. Add sensor quality endpoint:
- /autonomy/sensors/quality
- health, freshness, anomaly and failover state.

5. Add SLO endpoint:
- /autonomy/slo/status
- target, current, breach state, and burn rate.

## Execution Plan (6 Weeks)

## Week 1: HUD KPI Foundation

1. Implement buildAutonomyKPIModel and computeSLOStatusBadges.
2. Add Autonomy Control Strip to primary HUD.
3. Wire data from existing health/trend/trust payloads.
4. Add unit tests for KPI normalization and badge thresholds.

## Week 2: C2 Map Overlay Upgrade

1. Implement coverage and confidence overlays.
2. Add predicted trajectory rendering primitives.
3. Add replan trigger annotations and tooltip reasons.
4. Add visual regression snapshots for map layer rendering.

## Week 3: Recommendation and Explainability Layer

1. Implement deriveRecommendedCorrections.
2. Implement planner safety and explainability cards.
3. Add event correlation timeline model and UI.
4. Add contract tests for recommendation ranking logic.

## Week 4: Backend Autonomy Endpoints

1. Add twin summary endpoint.
2. Add planner insight endpoint.
3. Add sensor quality endpoint.
4. Add SLO status endpoint.
5. Add API auth and role checks consistent with existing policy enforcement.

## Week 5: Reliability and Performance

1. Add anomaly indicators and uncertainty envelopes.
2. Add client-side caching and stale-data fallbacks.
3. Add rate-safe polling fallback when SSE degrades.
4. Add performance budgets for HUD render and fetch cycles.

## Week 6: Hardening and Rollout

1. Add end-to-end scenario tests for correction loops.
2. Run chaos tests for endpoint degradation and sensor dropouts.
3. Add operator runbook and rollback controls.
4. Roll out behind feature flag with staged enablement.

## Metrics to Track

1. twin_lag_ms_p95
2. correction_success_rate
3. correction_reject_rate_policy
4. map_mean_confidence
5. stale_cell_ratio
6. hud_data_freshness_seconds
7. recommendation_acceptance_rate
8. operator_override_rate

## Testing Strategy

1. Unit tests:
- KPI builders, ranking logic, trigger detection, confidence overlays.

2. Integration tests:
- Endpoint contracts for autonomy summary and planner insights.

3. E2E tests:
- C2 command submit -> planner decision -> map update -> HUD recommendation verification.

4. Failure tests:
- SSE disconnect, stale telemetry, missing risk data, policy denial states.

## Rollout Strategy

1. Feature flags:
- HUD_AUTONOMY_STRIP
- HUD_C2_ADVANCED_OVERLAYS
- HUD_RECOMMENDATION_ASSIST

2. Canary stages:
- Stage 1: internal operators only.
- Stage 2: selected mission profiles.
- Stage 3: full rollout with override fallback.

3. Exit criteria:
- No critical UI regressions.
- SLOs stable for 7 consecutive days.
- Operator override rate trending down.

## Suggested First Implementation Tickets

1. HUD-101: Add buildAutonomyKPIModel + control strip.
2. HUD-102: Add SLO badges and breach timers.
3. C2-201: Coverage heatmap + confidence decay overlay.
4. C2-202: Predicted path rendering + uncertainty envelope.
5. API-301: /autonomy/twin/summary endpoint.
6. API-302: /autonomy/planner/insights endpoint.
7. API-303: /autonomy/sensors/quality endpoint.
8. QA-401: E2E correction loop scenario suite.
