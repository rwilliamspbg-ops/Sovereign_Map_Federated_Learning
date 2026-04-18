package autonomy

import (
	"testing"
	"time"
)

func TestCapabilityMatrixValidate(t *testing.T) {
	valid := CapabilityMatrix{
		Version:  "v1",
		Captured: time.Now().UTC(),
		Sensors: []SensorCapability{{
			ID:             "drone-gps",
			Class:          SensorClassPose,
			NominalHz:      20,
			SupportsReplay: true,
			Enabled:        true,
		}},
		Compute: []ComputeCapability{{
			ID:           "edge-0",
			Class:        ComputeClassEdge,
			GPUEnabled:   false,
			NPUEnabled:   true,
			MaxWorkers:   4,
			LatencyMsP95: 25,
		}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid matrix: %v", err)
	}

	invalid := valid
	invalid.Sensors = nil
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid matrix with no sensors")
	}
}

func TestClampUnitInterval(t *testing.T) {
	if got := ClampUnitInterval(-1.5); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
	if got := ClampUnitInterval(2.5); got != 1 {
		t.Fatalf("expected 1, got %f", got)
	}
	if got := ClampUnitInterval(0.42); got != 0.42 {
		t.Fatalf("expected identity, got %f", got)
	}
}
