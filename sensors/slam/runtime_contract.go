package slam

import (
	"fmt"
	"time"
)

// AlignmentConfig defines timestamp alignment policy for fused localization.
type AlignmentConfig struct {
	Window      time.Duration
	MaxSkew     time.Duration
	DropTooLate bool
}

// CalibrationManifest captures minimum calibration metadata required at runtime.
type CalibrationManifest struct {
	Version            string
	CameraIntrinsics   bool
	ExtrinsicsLoaded   bool
	CoordinateFrameSet bool
}

// PoseConfidence captures confidence score contract for fused pose outputs.
type PoseConfidence struct {
	Score   float64
	Source  string
	Reasons []string
}

// ValidateAlignmentConfig enforces sane alignment defaults and bounds.
func ValidateAlignmentConfig(cfg AlignmentConfig) error {
	if cfg.Window <= 0 {
		return fmt.Errorf("alignment window must be > 0")
	}
	if cfg.MaxSkew <= 0 {
		return fmt.Errorf("alignment max skew must be > 0")
	}
	if cfg.MaxSkew > cfg.Window {
		return fmt.Errorf("alignment max skew must be <= window")
	}
	return nil
}

// ValidateCalibrationManifest checks whether runtime calibration requirements are met.
func ValidateCalibrationManifest(m CalibrationManifest) error {
	if m.Version == "" {
		return fmt.Errorf("calibration manifest version is required")
	}
	if !m.CameraIntrinsics {
		return fmt.Errorf("camera intrinsics are required")
	}
	if !m.ExtrinsicsLoaded {
		return fmt.Errorf("extrinsics are required")
	}
	if !m.CoordinateFrameSet {
		return fmt.Errorf("coordinate frame set is required")
	}
	return nil
}

// NormalizeConfidence clamps confidence into [0,1] for downstream contracts.
func NormalizeConfidence(conf PoseConfidence) PoseConfidence {
	if conf.Score < 0 {
		conf.Score = 0
	}
	if conf.Score > 1 {
		conf.Score = 1
	}
	if conf.Source == "" {
		conf.Source = "unknown"
	}
	return conf
}
