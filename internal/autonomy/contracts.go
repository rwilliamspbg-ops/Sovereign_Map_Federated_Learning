package autonomy

import (
	"fmt"
	"time"
)

// SensorClass categorizes data sources for normalization and policy.
type SensorClass string

const (
	SensorClassSpatial    SensorClass = "spatial"
	SensorClassPose       SensorClass = "pose"
	SensorClassEnv        SensorClass = "environment"
	SensorClassSystem     SensorClass = "system"
	SensorClassExternal   SensorClass = "external"
	SensorClassUnassigned SensorClass = "unassigned"
)

// ComputeClass identifies compute targets used by workload scheduling.
type ComputeClass string

const (
	ComputeClassEdge    ComputeClass = "edge"
	ComputeClassNode    ComputeClass = "node"
	ComputeClassBackend ComputeClass = "backend"
)

// SensorCapability records a sensor endpoint and health envelope.
type SensorCapability struct {
	ID             string      `json:"id"`
	Class          SensorClass `json:"class"`
	NominalHz      float64     `json:"nominal_hz"`
	SupportsReplay bool        `json:"supports_replay"`
	Enabled        bool        `json:"enabled"`
}

// ComputeCapability records an execution target available to the scheduler.
type ComputeCapability struct {
	ID           string       `json:"id"`
	Class        ComputeClass `json:"class"`
	GPUEnabled   bool         `json:"gpu_enabled"`
	NPUEnabled   bool         `json:"npu_enabled"`
	MaxWorkers   int          `json:"max_workers"`
	LatencyMsP95 int          `json:"latency_ms_p95"`
}

// CapabilityMatrix defines sensor and compute inventory at runtime.
type CapabilityMatrix struct {
	Version  string              `json:"version"`
	Captured time.Time           `json:"captured"`
	Sensors  []SensorCapability  `json:"sensors"`
	Compute  []ComputeCapability `json:"compute"`
}

// TelemetryEvent is a normalized event contract consumed by fusion/planning.
type TelemetryEvent struct {
	EntityID         string            `json:"entity_id"`
	Timestamp        time.Time         `json:"timestamp"`
	Source           string            `json:"source"`
	SensorClass      SensorClass       `json:"sensor_class"`
	Confidence       float64           `json:"confidence"`
	Health           string            `json:"health"`
	Latitude         float64           `json:"latitude,omitempty"`
	Longitude        float64           `json:"longitude,omitempty"`
	AltitudeMeters   float64           `json:"altitude_meters,omitempty"`
	HeadingDegrees   float64           `json:"heading_degrees,omitempty"`
	GroundSpeedMS    float64           `json:"ground_speed_ms,omitempty"`
	AdditionalFields map[string]string `json:"additional_fields,omitempty"`
}

// TwinEntityState is the canonical state format persisted in the digital twin.
type TwinEntityState struct {
	EntityID       string            `json:"entity_id"`
	ObservedAt     time.Time         `json:"observed_at"`
	PoseConfidence float64           `json:"pose_confidence"`
	RiskScore      float64           `json:"risk_score"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

// Validate checks minimum baseline requirements for inventory readiness.
func (m CapabilityMatrix) Validate() error {
	if m.Version == "" {
		return fmt.Errorf("capability matrix version is required")
	}
	if len(m.Sensors) == 0 {
		return fmt.Errorf("capability matrix requires at least one sensor")
	}
	if len(m.Compute) == 0 {
		return fmt.Errorf("capability matrix requires at least one compute target")
	}
	for _, s := range m.Sensors {
		if s.ID == "" {
			return fmt.Errorf("sensor capability id is required")
		}
		if s.Class == "" {
			return fmt.Errorf("sensor %s class is required", s.ID)
		}
	}
	for _, c := range m.Compute {
		if c.ID == "" {
			return fmt.Errorf("compute capability id is required")
		}
		if c.Class == "" {
			return fmt.Errorf("compute %s class is required", c.ID)
		}
		if c.MaxWorkers <= 0 {
			return fmt.Errorf("compute %s max workers must be > 0", c.ID)
		}
	}
	return nil
}

// ClampUnitInterval normalizes confidence/risk values into [0,1].
func ClampUnitInterval(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
