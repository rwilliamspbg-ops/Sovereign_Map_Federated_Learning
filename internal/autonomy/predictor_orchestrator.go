package autonomy

import "time"

// PredictedState represents short-horizon digital twin forecast output.
type PredictedState struct {
	EntityID           string    `json:"entity_id"`
	PredictedAt        time.Time `json:"predicted_at"`
	HorizonSeconds     int       `json:"horizon_seconds"`
	PredictedLatitude  float64   `json:"predicted_latitude"`
	PredictedLongitude float64   `json:"predicted_longitude"`
	Confidence         float64   `json:"confidence"`
}

// PredictLinearMotion provides a low-compute baseline for short horizons.
func PredictLinearMotion(event TelemetryEvent, horizonSeconds int) PredictedState {
	delta := float64(horizonSeconds) * event.GroundSpeedMS
	latStep := delta / 111111.0
	return PredictedState{
		EntityID:           event.EntityID,
		PredictedAt:        time.Now().UTC(),
		HorizonSeconds:     horizonSeconds,
		PredictedLatitude:  event.Latitude + latStep,
		PredictedLongitude: event.Longitude,
		Confidence:         ClampUnitInterval(event.Confidence * 0.95),
	}
}

// WorkloadProfile defines execution constraints for compute scheduling.
type WorkloadProfile struct {
	Name                string
	RequiresAccelerator bool
	LatencyBudgetMs     int
	Priority            int
}

// SchedulingDecision identifies selected compute target for workload.
type SchedulingDecision struct {
	TargetID string `json:"target_id"`
	Reason   string `json:"reason"`
}

// SelectComputeTarget uses latency and accelerator constraints for routing.
func SelectComputeTarget(workload WorkloadProfile, inventory CapabilityMatrix) SchedulingDecision {
	selected := SchedulingDecision{TargetID: "", Reason: "no suitable compute target"}
	for _, target := range inventory.Compute {
		if target.MaxWorkers <= 0 {
			continue
		}
		if workload.RequiresAccelerator && !target.GPUEnabled && !target.NPUEnabled {
			continue
		}
		if workload.LatencyBudgetMs > 0 && target.LatencyMsP95 > workload.LatencyBudgetMs {
			continue
		}
		selected.TargetID = target.ID
		selected.Reason = "matched latency and accelerator constraints"
		if target.Class == ComputeClassEdge {
			selected.Reason = "selected edge for low-latency workload"
			return selected
		}
	}
	return selected
}
