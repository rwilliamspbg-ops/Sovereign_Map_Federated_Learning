package slam

import (
	"testing"
	"time"
)

func TestValidateAlignmentConfig(t *testing.T) {
	if err := ValidateAlignmentConfig(AlignmentConfig{Window: 250 * time.Millisecond, MaxSkew: 100 * time.Millisecond}); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if err := ValidateAlignmentConfig(AlignmentConfig{Window: 0, MaxSkew: 100 * time.Millisecond}); err == nil {
		t.Fatalf("expected invalid zero window")
	}
}

func TestValidateCalibrationManifest(t *testing.T) {
	valid := CalibrationManifest{
		Version:            "v1",
		CameraIntrinsics:   true,
		ExtrinsicsLoaded:   true,
		CoordinateFrameSet: true,
	}
	if err := ValidateCalibrationManifest(valid); err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}

	invalid := valid
	invalid.ExtrinsicsLoaded = false
	if err := ValidateCalibrationManifest(invalid); err == nil {
		t.Fatalf("expected invalid manifest")
	}
}

func TestNormalizeConfidence(t *testing.T) {
	c := NormalizeConfidence(PoseConfidence{Score: 1.7})
	if c.Score != 1 {
		t.Fatalf("expected score clamp to 1, got %f", c.Score)
	}
	if c.Source != "unknown" {
		t.Fatalf("expected unknown source default, got %s", c.Source)
	}
}
