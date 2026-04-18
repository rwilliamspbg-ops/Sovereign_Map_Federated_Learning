package autonomy

import "fmt"

// ReadinessChecklist captures production hardening requirements.
type ReadinessChecklist struct {
	HasChaosTests     bool
	HasRollbackPlan   bool
	HasSafetyGateway  bool
	HasAuditTrail     bool
	HasCanaryStrategy bool
	HasSLODashboards  bool
}

// ValidateReadiness enforces phase 6 minimum production gates.
func ValidateReadiness(c ReadinessChecklist) error {
	if !c.HasChaosTests {
		return fmt.Errorf("missing chaos test validation")
	}
	if !c.HasRollbackPlan {
		return fmt.Errorf("missing rollback plan")
	}
	if !c.HasSafetyGateway {
		return fmt.Errorf("missing safety gateway")
	}
	if !c.HasAuditTrail {
		return fmt.Errorf("missing audit trail")
	}
	if !c.HasCanaryStrategy {
		return fmt.Errorf("missing canary rollout strategy")
	}
	if !c.HasSLODashboards {
		return fmt.Errorf("missing SLO dashboards")
	}
	return nil
}
