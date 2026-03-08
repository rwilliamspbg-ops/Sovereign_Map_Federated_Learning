package slam

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

type serializedMap struct {
	Version     int            `json:"version"`
	Config      map[string]any `json:"config"`
	CurrentPose CameraPose     `json:"current_pose"`
	MapPoints   []MapPoint     `json:"map_points"`
	FrameCount  int            `json:"frame_count"`
	SavedAt     time.Time      `json:"saved_at"`
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
	httpClient  *http.Client
}

type trackRequest struct {
	Mode      SLAMMode `json:"mode"`
	Timestamp string   `json:"timestamp"`
	FrameJPEG string   `json:"frame_jpeg_b64"`
}

type trackResponse struct {
	Pose      CameraPose `json:"pose"`
	MapPoints []MapPoint `json:"map_points"`
	Status    string     `json:"status"`
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
	b.httpClient = &http.Client{Timeout: b.config.Timeout}

	// If endpoint is not configured, bridge remains in local fallback mode.
	if strings.TrimSpace(b.config.Endpoint) == "" {
		b.initialized = true
		return nil
	}

	// Probe endpoint health when remote SLAM backend is configured.
	ctx, cancel := context.WithTimeout(context.Background(), b.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.endpointURL("/health"), nil)
	if err != nil {
		return fmt.Errorf("create health request: %w", err)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect ORB-SLAM endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ORB-SLAM endpoint unhealthy: status=%d", resp.StatusCode)
	}

	b.initialized = true
	return nil
}

// ProcessFrame feeds image frame to SLAM and returns camera pose.
func (b *ORBSLAMBridge) ProcessFrame(ctx context.Context, img gocv.Mat, timestamp time.Time) (CameraPose, error) {
	if !b.initialized {
		return CameraPose{}, fmt.Errorf("SLAM bridge not initialized")
	}
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	b.frameCount++
	if strings.TrimSpace(b.config.Endpoint) == "" {
		// Local fallback mode keeps deterministic behavior without remote engine.
		pose := CameraPose{
			X: 0, Y: 0, Z: 0,
			Qw: 1, Qx: 0, Qy: 0, Qz: 0,
			Timestamp:     timestamp,
			TrackingState: "LOCAL_FALLBACK",
		}
		b.currentPose = pose
		return pose, nil
	}

	if img.Empty() {
		return CameraPose{}, fmt.Errorf("empty frame")
	}

	encoded, err := gocv.IMEncode(".jpg", img)
	if err != nil {
		return CameraPose{}, fmt.Errorf("encode frame: %w", err)
	}
	defer encoded.Close()

	reqPayload := trackRequest{
		Mode:      b.config.Mode,
		Timestamp: timestamp.UTC().Format(time.RFC3339Nano),
		FrameJPEG: base64.StdEncoding.EncodeToString(encoded.GetBytes()),
	}
	body, err := json.Marshal(reqPayload)
	if err != nil {
		return CameraPose{}, fmt.Errorf("marshal track request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.endpointURL("/track"), bytes.NewReader(body))
	if err != nil {
		return CameraPose{}, fmt.Errorf("create track request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return CameraPose{}, fmt.Errorf("track frame via ORB-SLAM endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return CameraPose{}, fmt.Errorf("track request failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var out trackResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CameraPose{}, fmt.Errorf("decode track response: %w", err)
	}

	pose := out.Pose
	if pose.Timestamp.IsZero() {
		pose.Timestamp = timestamp
	}
	if strings.TrimSpace(pose.TrackingState) == "" {
		if strings.TrimSpace(out.Status) != "" {
			pose.TrackingState = out.Status
		} else {
			pose.TrackingState = "REMOTE_OK"
		}
	}

	b.mapPoints = append([]MapPoint(nil), out.MapPoints...)
	b.currentPose = pose
	return pose, nil
}

func (b *ORBSLAMBridge) endpointURL(path string) string {
	base := strings.TrimRight(strings.TrimSpace(b.config.Endpoint), "/")
	if base == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
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
