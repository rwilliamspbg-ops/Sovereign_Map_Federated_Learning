package autonomy

import (
	"testing"
	"time"
)

func TestLayeredMapStoreAndDecay(t *testing.T) {
	store := NewLayeredMapStore("traversability")
	store.UpsertCell(MapCell{X: 1, Y: 2, Confidence: 0.9, Occupancy: 0.2, LastUpdatedAt: time.Now().UTC().Add(-2 * time.Hour)})
	before := store.Snapshot()
	if before.CellCount != 1 {
		t.Fatalf("expected one cell, got %d", before.CellCount)
	}
	store.ApplyConfidenceDecay(time.Now().UTC(), 1*time.Hour)
	after := store.Snapshot()
	if after.MeanConfidence >= before.MeanConfidence {
		t.Fatalf("expected confidence decay: before=%f after=%f", before.MeanConfidence, after.MeanConfidence)
	}
}

func TestTwinStoreLag(t *testing.T) {
	twin := NewTwinStore()
	twin.Upsert(TwinEntityState{EntityID: "drone-01", ObservedAt: time.Now().UTC().Add(-1 * time.Second), PoseConfidence: 1.2, RiskScore: -1})
	state, ok := twin.Get("drone-01")
	if !ok {
		t.Fatal("expected twin state")
	}
	if state.PoseConfidence != 1 {
		t.Fatalf("expected confidence clamp to 1, got %f", state.PoseConfidence)
	}
	if state.RiskScore != 0 {
		t.Fatalf("expected risk clamp to 0, got %f", state.RiskScore)
	}
	lag, ok := twin.StateLag("drone-01", time.Now().UTC())
	if !ok || lag <= 0 {
		t.Fatalf("expected positive lag, got %v", lag)
	}
}

func TestSelectBestCorrection(t *testing.T) {
	decision := SelectBestCorrection([]CorrectionOption{
		{Action: "reroute_left", EstimatedRisk: 0.2, EstimatedMissionGain: 0.8, EstimatedCost: 0.3},
		{Action: "accelerate", EstimatedRisk: 0.7, EstimatedMissionGain: 0.9, EstimatedCost: 0.4},
	}, PlannerPolicy{MaxAcceptedRisk: 0.4, MinMissionGain: 0.5})
	if decision.SelectedAction != "reroute_left" {
		t.Fatalf("unexpected action: %s", decision.SelectedAction)
	}
	if decision.RejectedUnsafe {
		t.Fatal("expected safe action selection")
	}
}

func TestPredictLinearMotion(t *testing.T) {
	pred := PredictLinearMotion(TelemetryEvent{
		EntityID:      "drone-01",
		Latitude:      37.0,
		Longitude:     -122.0,
		GroundSpeedMS: 10,
		Confidence:    0.9,
	}, 60)
	if pred.PredictedLatitude <= 37.0 {
		t.Fatalf("expected predicted latitude to increase, got %f", pred.PredictedLatitude)
	}
	if pred.Confidence <= 0 || pred.Confidence > 1 {
		t.Fatalf("expected confidence in (0,1], got %f", pred.Confidence)
	}
}

func TestSelectComputeTarget(t *testing.T) {
	matrix := CapabilityMatrix{
		Version: "v1",
		Sensors: []SensorCapability{{ID: "cam-1", Class: SensorClassSpatial, Enabled: true}},
		Compute: []ComputeCapability{
			{ID: "backend-a", Class: ComputeClassBackend, GPUEnabled: true, MaxWorkers: 32, LatencyMsP95: 120},
			{ID: "edge-a", Class: ComputeClassEdge, NPUEnabled: true, MaxWorkers: 4, LatencyMsP95: 20},
		},
	}
	decision := SelectComputeTarget(WorkloadProfile{Name: "realtime-planner", RequiresAccelerator: true, LatencyBudgetMs: 40}, matrix)
	if decision.TargetID != "edge-a" {
		t.Fatalf("expected edge-a target, got %s", decision.TargetID)
	}
}

func TestValidateReadiness(t *testing.T) {
	check := ReadinessChecklist{
		HasChaosTests:     true,
		HasRollbackPlan:   true,
		HasSafetyGateway:  true,
		HasAuditTrail:     true,
		HasCanaryStrategy: true,
		HasSLODashboards:  true,
	}
	if err := ValidateReadiness(check); err != nil {
		t.Fatalf("expected readiness valid: %v", err)
	}

	check.HasChaosTests = false
	if err := ValidateReadiness(check); err == nil {
		t.Fatal("expected readiness validation error")
	}
}
