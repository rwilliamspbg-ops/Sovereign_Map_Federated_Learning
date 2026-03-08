//go:build opencv

package slam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gocv.io/x/gocv"
)

func TestNewORBSLAMBridgeRemoteHealthCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	bridge, err := NewORBSLAMBridge(ORBSLAMBridgeConfig{
		Enabled:  true,
		Endpoint: ts.URL,
		Timeout:  2 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected healthy endpoint, got error: %v", err)
	}
	if !bridge.initialized {
		t.Fatalf("bridge should be initialized")
	}
}

func TestProcessFrameRemoteEndpoint(t *testing.T) {
	trackCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/track":
			trackCalled = true
			var req trackRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.FrameJPEG == "" {
				t.Fatalf("frame payload should not be empty")
			}
			resp := trackResponse{
				Pose: CameraPose{
					X: 1.25, Y: -0.5, Z: 3.0,
					Qw: 1, Qx: 0, Qy: 0, Qz: 0,
					TrackingState: "REMOTE_TRACKING",
				},
				MapPoints: []MapPoint{{ID: 1, X: 0.1, Y: 0.2, Z: 0.3, Observations: 4}},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	bridge, err := NewORBSLAMBridge(ORBSLAMBridgeConfig{
		Enabled:  true,
		Endpoint: ts.URL,
		Timeout:  2 * time.Second,
	})
	if err != nil {
		t.Fatalf("init bridge: %v", err)
	}

	img := gocv.NewMatWithSize(8, 8, gocv.MatTypeCV8UC3)
	defer img.Close()

	pose, err := bridge.ProcessFrame(context.Background(), img, time.Now().UTC())
	if err != nil {
		t.Fatalf("process frame: %v", err)
	}
	if !trackCalled {
		t.Fatalf("track endpoint was not called")
	}
	if pose.TrackingState != "REMOTE_TRACKING" {
		t.Fatalf("unexpected tracking state: %s", pose.TrackingState)
	}
	if len(bridge.GetMap()) != 1 {
		t.Fatalf("expected one map point, got %d", len(bridge.GetMap()))
	}
}

func TestProcessFrameLocalFallback(t *testing.T) {
	bridge, err := NewORBSLAMBridge(ORBSLAMBridgeConfig{Enabled: true})
	if err != nil {
		t.Fatalf("init bridge fallback: %v", err)
	}

	img := gocv.NewMatWithSize(4, 4, gocv.MatTypeCV8UC3)
	defer img.Close()
	pose, err := bridge.ProcessFrame(context.Background(), img, time.Now().UTC())
	if err != nil {
		t.Fatalf("process frame fallback: %v", err)
	}
	if pose.TrackingState != "LOCAL_FALLBACK" {
		t.Fatalf("expected LOCAL_FALLBACK, got %s", pose.TrackingState)
	}
}
