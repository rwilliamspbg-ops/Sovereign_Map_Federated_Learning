package slam

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gocv.io/x/gocv"
)

type serializedMap struct {
	Version     int         `json:"version"`
	Config      map[string]any `json:"config"`
	CurrentPose CameraPose  `json:"current_pose"`
	MapPoints   []MapPoint  `json:"map_points"`
	FrameCount  int         `json:"frame_count"`
	SavedAt     time.Time   `json:"saved_at"`
}

// SLAMMode identifies SLAM operating mode.
type SLAMMode string

const (
	ModeMono   SLAMMode = "monocular"
	ModeStereo SLAMMode = "stereo"
	ModeRGBD   SLAMMode = "rgbd"
)

// ORBSLAMBridgeConfig stores bridge settings to an external SLAM engine.
type ORBSLAMBridgeConfig struct {
	Enabled        bool
	Mode           SLAMMode
	VocabularyPath string
	SettingsPath   string
	Endpoint       string
	Timeout        time.Duration
}

// CameraPose tracks 6DOF position and orientation.
type CameraPose struct {
	X, Y, Z        float64
	Qw, Qx, Qy, Qz float64
	Timestamp      time.Time
	TrackingState  string
}

// MapPoint represents a 3D landmark in the SLAM map.
type MapPoint struct {
	ID           int
	X, Y, Z      float64
	Observations int
	Descriptor   []byte
}

// ORBSLAMBridge interfaces with ORB-SLAM3 SLAM system.
type ORBSLAMBridge struct {
	config      ORBSLAMBridgeConfig
	initialized bool
	currentPose CameraPose
	mapPoints   []MapPoint
	frameCount  int
}

// NewORBSLAMBridge creates ORB-SLAM bridge connection.
func NewORBSLAMBridge(cfg ORBSLAMBridgeConfig) (*ORBSLAMBridge, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeMono
	}

	bridge := &ORBSLAMBridge{
		config:    cfg,
		mapPoints: make([]MapPoint, 0, 10000),
	}

	if cfg.Enabled {
		if err := bridge.initialize(); err != nil {
			return nil, fmt.Errorf("initialize ORB-SLAM: %w", err)
		}
	}

	return bridge, nil
}

// initialize starts ORB-SLAM system with vocabulary and camera settings.
func (b *ORBSLAMBridge) initialize() error {
	// In production, this would load ORB vocabulary and camera calibration
	// For now, mark as initialized for stub mode
	b.initialized = true
	return nil
}

// ProcessFrame feeds image frame to SLAM and returns camera pose.
func (b *ORBSLAMBridge) ProcessFrame(ctx context.Context, img gocv.Mat, timestamp time.Time) (CameraPose, error) {
	if !b.initialized {
		return CameraPose{}, fmt.Errorf("SLAM bridge not initialized")
	}

	b.frameCount++

	// In production, this would call ORB-SLAM3 TrackMonocular/TrackStereo/TrackRGBD
	// For stub mode, return identity pose
	pose := CameraPose{
		X: 0, Y: 0, Z: 0,
		Qw: 1, Qx: 0, Qy: 0, Qz: 0,
		Timestamp:     timestamp,
		TrackingState: "OK",
	}

	b.currentPose = pose
	return pose, nil
}

// GetMap retrieves current SLAM map points.
func (b *ORBSLAMBridge) GetMap() []MapPoint {
	return b.mapPoints
}

// GetPose returns most recent camera pose estimate.
func (b *ORBSLAMBridge) GetPose() CameraPose {
	return b.currentPose
}

// Reset clears SLAM map and tracking state.
func (b *ORBSLAMBridge) Reset() error {
	b.mapPoints = make([]MapPoint, 0, 10000)
	b.currentPose = CameraPose{}
	b.frameCount = 0
	return nil
}

// SaveMap exports SLAM map to file for later reuse.
func (b *ORBSLAMBridge) SaveMap(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create map directory: %w", err)
	}

	payload := serializedMap{
		Version: 1,
		Config: map[string]any{
			"mode":            b.config.Mode,
			"vocabulary_path": b.config.VocabularyPath,
			"settings_path":   b.config.SettingsPath,
			"endpoint":        b.config.Endpoint,
		},
		CurrentPose: b.currentPose,
		MapPoints:   append([]MapPoint(nil), b.mapPoints...),
		FrameCount:  b.frameCount,
		SavedAt:     time.Now().UTC(),
	}

	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal map: %w", err)
	}

	if err := os.WriteFile(filepath.Clean(path), raw, 0o644); err != nil {
		return fmt.Errorf("write map file: %w", err)
	}

	return nil
}

// LoadMap imports previously saved SLAM map.
func (b *ORBSLAMBridge) LoadMap(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read map file: %w", err)
	}

	var payload serializedMap
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("unmarshal map file: %w", err)
	}
	if payload.Version <= 0 {
		return fmt.Errorf("unsupported map format version: %d", payload.Version)
	}

	b.currentPose = payload.CurrentPose
	b.mapPoints = append([]MapPoint(nil), payload.MapPoints...)
	b.frameCount = payload.FrameCount
	b.initialized = true

	return nil
}

// Shutdown cleanly stops SLAM system.
func (b *ORBSLAMBridge) Shutdown() error {
	b.initialized = false
	return nil
}

// GetStats returns SLAM performance metrics.
func (b *ORBSLAMBridge) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"frames_processed": b.frameCount,
		"map_points":       len(b.mapPoints),
		"tracking_state":   b.currentPose.TrackingState,
	}
}
